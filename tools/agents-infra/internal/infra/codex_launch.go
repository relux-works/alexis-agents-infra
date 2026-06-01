package infra

import (
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strconv"
	"strings"
)

const codexDangerouslyBypassApprovalsAndSandbox = "--dangerously-bypass-approvals-and-sandbox"

type CodexLaunchPlan struct {
	StartDir                 string
	HomeDir                  string
	BaseCodexConfigPath      string
	BaseCodexConfigPresent   bool
	ProjectConfigs           []CodexProjectConfigSource
	RegistrySources          []CodexMCPRegistrySource
	MCPServers               []CodexMCPLaunchServer
	ConfigArgs               []string
	UserArgs                 []string
	Args                     []string
	PrintConfig              bool
	WrapperExpandedShortcuts []CodexWrapperShortcut
}

type CodexProjectConfigSource struct {
	Path           string
	EnabledServers []string
}

type CodexMCPRegistrySource struct {
	Path        string
	Scope       string
	ServerNames []string
}

type CodexMCPLaunchServer struct {
	Name              string
	URL               string
	BearerTokenEnvVar string
	DefinitionSource  string
	EnabledBy         []string
}

type CodexWrapperShortcut struct {
	From string
	To   string
}

type codexMCPDefinition struct {
	Server codexMCPServer
	Source string
}

func BuildCodexLaunchPlan(startDir, homeDir string, args []string) (CodexLaunchPlan, error) {
	parsed, err := parseCodexWrapperArgs(args)
	if err != nil {
		return CodexLaunchPlan{}, err
	}
	if startDir == "" {
		startDir, err = os.Getwd()
		if err != nil {
			return CodexLaunchPlan{}, fmt.Errorf("resolve cwd: %w", err)
		}
	}
	startDir, err = filepath.Abs(startDir)
	if err != nil {
		return CodexLaunchPlan{}, fmt.Errorf("resolve start dir: %w", err)
	}
	if homeDir == "" {
		homeDir, err = os.UserHomeDir()
		if err != nil {
			return CodexLaunchPlan{}, fmt.Errorf("resolve home dir: %w", err)
		}
	}
	homeDir, err = filepath.Abs(homeDir)
	if err != nil {
		return CodexLaunchPlan{}, fmt.Errorf("resolve home dir: %w", err)
	}

	plan := CodexLaunchPlan{
		StartDir:                 startDir,
		HomeDir:                  homeDir,
		BaseCodexConfigPath:      filepath.Join(homeDir, ".codex", "config.toml"),
		BaseCodexConfigPresent:   pathExists(filepath.Join(homeDir, ".codex", "config.toml")),
		UserArgs:                 parsed.codexArgs,
		PrintConfig:              parsed.printConfig,
		WrapperExpandedShortcuts: parsed.expandedShortcuts,
	}

	ancestors := ancestorDirsRootFirst(startDir)
	enabledOrder, enabledBy, projectConfigs, err := loadCompositeMCPEnablement(ancestors)
	if err != nil {
		return CodexLaunchPlan{}, err
	}
	plan.ProjectConfigs = projectConfigs

	definitions, registrySources, err := loadCompositeMCPRegistry(homeDir, ancestors)
	if err != nil {
		return CodexLaunchPlan{}, err
	}
	plan.RegistrySources = registrySources

	for _, name := range enabledOrder {
		def, ok := definitions[name]
		if !ok {
			return CodexLaunchPlan{}, fmt.Errorf("MCP server %q is enabled by %s but no definition was found in codex-mcp-servers.toml registries", name, strings.Join(enabledBy[name], ", "))
		}
		if def.Server.URL == "" {
			return CodexLaunchPlan{}, fmt.Errorf("MCP server %q is defined by %s but is missing url", name, def.Source)
		}
		server := CodexMCPLaunchServer{
			Name:              name,
			URL:               def.Server.URL,
			BearerTokenEnvVar: def.Server.BearerTokenEnvVar,
			DefinitionSource:  def.Source,
			EnabledBy:         append([]string(nil), enabledBy[name]...),
		}
		plan.MCPServers = append(plan.MCPServers, server)
		plan.ConfigArgs = append(plan.ConfigArgs, "-c", fmt.Sprintf("mcp_servers.%s.url=%q", name, server.URL))
		if server.BearerTokenEnvVar != "" {
			plan.ConfigArgs = append(plan.ConfigArgs, "-c", fmt.Sprintf("mcp_servers.%s.bearer_token_env_var=%q", name, server.BearerTokenEnvVar))
		}
	}
	plan.Args = append(append([]string(nil), plan.ConfigArgs...), plan.UserArgs...)
	return plan, nil
}

func RenderCodexLaunchPlan(plan CodexLaunchPlan) string {
	var out strings.Builder
	out.WriteString("agents-infra codex config\n")
	fmt.Fprintf(&out, "cwd: %s\n", plan.StartDir)
	if plan.BaseCodexConfigPresent {
		fmt.Fprintf(&out, "base_codex_config: %s\n", plan.BaseCodexConfigPath)
	} else {
		fmt.Fprintf(&out, "base_codex_config: %s (missing)\n", plan.BaseCodexConfigPath)
	}

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
			fmt.Fprintf(&out, "    url: %s\n", server.URL)
			if server.BearerTokenEnvVar != "" {
				fmt.Fprintf(&out, "    bearer_token_env_var: %s\n", server.BearerTokenEnvVar)
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

	out.WriteString("codex_args:\n")
	if len(plan.Args) == 0 {
		out.WriteString("  - (none)\n")
	} else {
		for _, arg := range plan.Args {
			fmt.Fprintf(&out, "  - %s\n", strconv.Quote(arg))
		}
	}
	return out.String()
}

type parsedCodexWrapperArgs struct {
	codexArgs         []string
	printConfig       bool
	expandedShortcuts []CodexWrapperShortcut
}

func parseCodexWrapperArgs(args []string) (parsedCodexWrapperArgs, error) {
	var parsed parsedCodexWrapperArgs
	passThrough := false
	for _, arg := range args {
		if passThrough {
			parsed.codexArgs = append(parsed.codexArgs, arg)
			continue
		}
		switch arg {
		case "--":
			passThrough = true
		case "--print-config":
			parsed.printConfig = true
		case "-d", "--danger", "--yolo":
			parsed.codexArgs = append(parsed.codexArgs, codexDangerouslyBypassApprovalsAndSandbox)
			parsed.expandedShortcuts = append(parsed.expandedShortcuts, CodexWrapperShortcut{
				From: arg,
				To:   codexDangerouslyBypassApprovalsAndSandbox,
			})
		default:
			parsed.codexArgs = append(parsed.codexArgs, arg)
		}
	}
	return parsed, nil
}

func ancestorDirsRootFirst(startDir string) []string {
	dir := filepath.Clean(startDir)
	var cwdFirst []string
	for {
		cwdFirst = append(cwdFirst, dir)
		parent := filepath.Dir(dir)
		if parent == dir {
			break
		}
		dir = parent
	}
	rootFirst := make([]string, 0, len(cwdFirst))
	for i := len(cwdFirst) - 1; i >= 0; i-- {
		rootFirst = append(rootFirst, cwdFirst[i])
	}
	return rootFirst
}

func loadCompositeMCPEnablement(ancestors []string) ([]string, map[string][]string, []CodexProjectConfigSource, error) {
	var enabledOrder []string
	enabledSeen := map[string]bool{}
	enabledBy := map[string][]string{}
	var sources []CodexProjectConfigSource

	for _, dir := range ancestors {
		path := filepath.Join(dir, ".agents", ".configs", projectConfigFileName)
		data, err := os.ReadFile(path)
		if os.IsNotExist(err) {
			continue
		}
		if err != nil {
			return nil, nil, nil, fmt.Errorf("read project config %s: %w", path, err)
		}
		servers, err := parseEnabledMCPServers(data, path)
		if err != nil {
			return nil, nil, nil, err
		}
		source := CodexProjectConfigSource{
			Path:           path,
			EnabledServers: append([]string(nil), servers...),
		}
		sources = append(sources, source)
		for _, name := range servers {
			if !isBareTOMLKey(name) {
				return nil, nil, nil, fmt.Errorf("MCP server name %q in %s is not a supported TOML bare key", name, path)
			}
			if !enabledSeen[name] {
				enabledOrder = append(enabledOrder, name)
				enabledSeen[name] = true
			}
			enabledBy[name] = append(enabledBy[name], path)
		}
	}
	return enabledOrder, enabledBy, sources, nil
}

func loadCompositeMCPRegistry(homeDir string, ancestors []string) (map[string]codexMCPDefinition, []CodexMCPRegistrySource, error) {
	definitions := map[string]codexMCPDefinition{}
	var sources []CodexMCPRegistrySource

	globalPath := filepath.Join(homeDir, ".agents", ".configs", "codex-mcp-servers.toml")
	if err := mergeMCPRegistry(definitions, &sources, globalPath, "global"); err != nil {
		return nil, nil, err
	}
	for _, dir := range ancestors {
		path := filepath.Join(dir, ".agents", ".configs", "codex-mcp-servers.toml")
		if samePath(path, globalPath) {
			continue
		}
		if err := mergeMCPRegistry(definitions, &sources, path, "project"); err != nil {
			return nil, nil, err
		}
	}
	return definitions, sources, nil
}

func mergeMCPRegistry(definitions map[string]codexMCPDefinition, sources *[]CodexMCPRegistrySource, path, scope string) error {
	data, err := os.ReadFile(path)
	if os.IsNotExist(err) {
		return nil
	}
	if err != nil {
		return fmt.Errorf("read Codex MCP registry %s: %w", path, err)
	}
	registry, err := parseCodexMCPRegistry(data, path)
	if err != nil {
		return err
	}
	names := make([]string, 0, len(registry))
	for name, server := range registry {
		names = append(names, name)
		definitions[name] = codexMCPDefinition{
			Server: server,
			Source: path,
		}
	}
	sort.Strings(names)
	*sources = append(*sources, CodexMCPRegistrySource{
		Path:        path,
		Scope:       scope,
		ServerNames: names,
	})
	return nil
}
