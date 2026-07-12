package infra

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"
)

const claudeDangerouslySkipPermissions = "--dangerously-skip-permissions"

type ClaudeLaunchPlan struct {
	StartDir                 string
	HomeDir                  string
	ProjectConfigs           []ClaudeProjectConfigSource
	RegistrySources          []CodexMCPRegistrySource
	MCPServers               []ClaudeMCPLaunchServer
	MCPConfigJSON            string
	ConfigArgs               []string
	UserArgs                 []string
	Args                     []string
	PrintConfig              bool
	WrapperExpandedShortcuts []CodexWrapperShortcut
}

type ClaudeProjectConfigSource struct {
	Path           string
	EnabledServers []string
}

type ClaudeMCPLaunchServer struct {
	Name              string
	URL               string
	BearerTokenEnvVar string
	Command           string
	Args              []string
	DefinitionSource  string
	EnabledBy         []string
}

// BuildClaudeLaunchPlan mirrors BuildCodexLaunchPlan: it walks the same
// ancestor .agents/.configs/project-config.toml files, reads the same
// shared [mcp] enabled_servers opt-in and the same codex-mcp-servers.toml
// registries, but renders the result as a Claude Code --mcp-config JSON
// payload instead of Codex `-c` overrides. The opt-in list is intentionally
// agent-agnostic — one list drives which servers agents-infra codex and
// agents-infra claude each compose into their own format.
func BuildClaudeLaunchPlan(startDir, homeDir string, args []string) (ClaudeLaunchPlan, error) {
	parsed, err := parseClaudeWrapperArgs(args)
	if err != nil {
		return ClaudeLaunchPlan{}, err
	}
	if startDir == "" {
		var err error
		startDir, err = os.Getwd()
		if err != nil {
			return ClaudeLaunchPlan{}, fmt.Errorf("resolve cwd: %w", err)
		}
	}
	startDir, err = filepath.Abs(startDir)
	if err != nil {
		return ClaudeLaunchPlan{}, fmt.Errorf("resolve start dir: %w", err)
	}
	if homeDir == "" {
		var err error
		homeDir, err = os.UserHomeDir()
		if err != nil {
			return ClaudeLaunchPlan{}, fmt.Errorf("resolve home dir: %w", err)
		}
	}
	homeDir, err = filepath.Abs(homeDir)
	if err != nil {
		return ClaudeLaunchPlan{}, fmt.Errorf("resolve home dir: %w", err)
	}

	plan := ClaudeLaunchPlan{
		StartDir:                 startDir,
		HomeDir:                  homeDir,
		UserArgs:                 parsed.claudeArgs,
		PrintConfig:              parsed.printConfig,
		WrapperExpandedShortcuts: parsed.expandedShortcuts,
	}

	ancestors := ancestorDirsRootFirst(startDir)
	globalProjectConfigPath := filepath.Join(homeDir, ".agents", ".configs", projectConfigFileName)
	enabledOrder, enabledBy, projectConfigs, err := loadCompositeMCPEnablement(ancestors, globalProjectConfigPath, "mcp")
	if err != nil {
		return ClaudeLaunchPlan{}, err
	}
	for _, source := range projectConfigs {
		plan.ProjectConfigs = append(plan.ProjectConfigs, ClaudeProjectConfigSource{
			Path:           source.Path,
			EnabledServers: source.EnabledServers,
		})
	}

	definitions, registrySources, err := loadCompositeMCPRegistry(homeDir, ancestors)
	if err != nil {
		return ClaudeLaunchPlan{}, err
	}
	plan.RegistrySources = registrySources

	mcpServers := map[string]claudeMCPConfigServer{}
	for _, name := range enabledOrder {
		def, ok := definitions[name]
		if !ok {
			return ClaudeLaunchPlan{}, fmt.Errorf("MCP server %q is enabled by %s but no definition was found in codex-mcp-servers.toml registries", name, strings.Join(enabledBy[name], ", "))
		}
		if err := validateCodexMCPDefinition(name, def); err != nil {
			return ClaudeLaunchPlan{}, err
		}
		server := ClaudeMCPLaunchServer{
			Name:              name,
			URL:               def.Server.URL,
			BearerTokenEnvVar: def.Server.BearerTokenEnvVar,
			Command:           def.Server.Command,
			Args:              append([]string(nil), def.Server.Args...),
			DefinitionSource:  def.Source,
			EnabledBy:         append([]string(nil), enabledBy[name]...),
		}
		plan.MCPServers = append(plan.MCPServers, server)
		mcpServers[name] = claudeMCPConfigServer(server)
	}

	if len(mcpServers) > 0 {
		configJSON, err := marshalClaudeMCPConfig(mcpServers)
		if err != nil {
			return ClaudeLaunchPlan{}, fmt.Errorf("encode Claude MCP config: %w", err)
		}
		plan.MCPConfigJSON = configJSON
		plan.ConfigArgs = []string{"--mcp-config", configJSON}
	}
	plan.Args = append(append([]string(nil), plan.ConfigArgs...), plan.UserArgs...)
	return plan, nil
}

type claudeMCPConfigServer ClaudeMCPLaunchServer

type claudeMCPConfigEntry struct {
	Type    string            `json:"type"`
	URL     string            `json:"url,omitempty"`
	Headers map[string]string `json:"headers,omitempty"`
	Command string            `json:"command,omitempty"`
	Args    []string          `json:"args,omitempty"`
}

func marshalClaudeMCPConfig(servers map[string]claudeMCPConfigServer) (string, error) {
	entries := make(map[string]claudeMCPConfigEntry, len(servers))
	for name, server := range servers {
		if server.URL != "" {
			entry := claudeMCPConfigEntry{Type: "http", URL: server.URL}
			if server.BearerTokenEnvVar != "" {
				entry.Headers = map[string]string{
					"Authorization": fmt.Sprintf("Bearer ${%s}", server.BearerTokenEnvVar),
				}
			}
			entries[name] = entry
			continue
		}
		entries[name] = claudeMCPConfigEntry{
			Type:    "stdio",
			Command: server.Command,
			Args:    server.Args,
		}
	}
	payload, err := json.Marshal(struct {
		MCPServers map[string]claudeMCPConfigEntry `json:"mcpServers"`
	}{MCPServers: entries})
	if err != nil {
		return "", err
	}
	return string(payload), nil
}

func RenderClaudeLaunchPlan(plan ClaudeLaunchPlan) string {
	var out strings.Builder
	out.WriteString("agents-infra claude config\n")
	fmt.Fprintf(&out, "cwd: %s\n", plan.StartDir)

	out.WriteString("project_configs:\n")
	if len(plan.ProjectConfigs) == 0 {
		out.WriteString("  - (none)\n")
	} else {
		for _, source := range plan.ProjectConfigs {
			if len(source.EnabledServers) == 0 {
				fmt.Fprintf(&out, "  - %s: enabled_mcp=(none)\n", source.Path)
			} else {
				fmt.Fprintf(&out, "  - %s: enabled_mcp=%s\n", source.Path, strings.Join(source.EnabledServers, ","))
			}
		}
	}

	out.WriteString("mcp_registries:\n")
	if len(plan.RegistrySources) == 0 {
		out.WriteString("  - (none)\n")
	} else {
		for _, source := range plan.RegistrySources {
			if len(source.ServerNames) == 0 {
				fmt.Fprintf(&out, "  - %s (%s): servers=(none)\n", source.Path, source.Scope)
			} else {
				fmt.Fprintf(&out, "  - %s (%s): servers=%s\n", source.Path, source.Scope, strings.Join(source.ServerNames, ","))
			}
		}
	}

	out.WriteString("enabled_mcp:\n")
	if len(plan.MCPServers) == 0 {
		out.WriteString("  - (none)\n")
	} else {
		for _, server := range plan.MCPServers {
			fmt.Fprintf(&out, "  - %s\n", server.Name)
			fmt.Fprintf(&out, "    enabled_by: %s\n", strings.Join(server.EnabledBy, ", "))
			fmt.Fprintf(&out, "    definition: %s\n", server.DefinitionSource)
			if server.URL != "" {
				fmt.Fprintf(&out, "    url: %s\n", server.URL)
				if server.BearerTokenEnvVar != "" {
					fmt.Fprintf(&out, "    bearer_token_env_var: %s\n", server.BearerTokenEnvVar)
				}
			} else {
				fmt.Fprintf(&out, "    command: %s\n", server.Command)
				if len(server.Args) > 0 {
					fmt.Fprintf(&out, "    args: %s\n", formatTOMLStringArray(server.Args))
				}
			}
		}
	}

	out.WriteString("wrapper_expansions:\n")
	if len(plan.WrapperExpandedShortcuts) == 0 {
		out.WriteString("  - (none)\n")
	} else {
		for _, shortcut := range plan.WrapperExpandedShortcuts {
			fmt.Fprintf(&out, "  - %s => %s\n", shortcut.From, shortcut.To)
		}
	}

	out.WriteString("claude_args:\n")
	if len(plan.Args) == 0 {
		out.WriteString("  - (none)\n")
	} else {
		for _, arg := range plan.Args {
			fmt.Fprintf(&out, "  - %s\n", strconv.Quote(arg))
		}
	}
	return out.String()
}

type parsedClaudeWrapperArgs struct {
	claudeArgs        []string
	printConfig       bool
	expandedShortcuts []CodexWrapperShortcut
}

func parseClaudeWrapperArgs(args []string) (parsedClaudeWrapperArgs, error) {
	var parsed parsedClaudeWrapperArgs
	passThrough := false
	for _, arg := range args {
		if passThrough {
			parsed.claudeArgs = append(parsed.claudeArgs, arg)
			continue
		}
		switch arg {
		case "--":
			passThrough = true
		case "--print-config":
			parsed.printConfig = true
		case "-d", "--danger", "--yolo":
			parsed.claudeArgs = append(parsed.claudeArgs, claudeDangerouslySkipPermissions)
			parsed.expandedShortcuts = append(parsed.expandedShortcuts, CodexWrapperShortcut{
				From: arg,
				To:   claudeDangerouslySkipPermissions,
			})
		default:
			parsed.claudeArgs = append(parsed.claudeArgs, arg)
		}
	}
	return parsed, nil
}
