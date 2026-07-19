package infra

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

const (
	ChildLaunchCompositionContract      = "agents-infra.child-launch-composition"
	ChildLaunchCompositionSchemaVersion = 1
)

type ChildLaunchCompositionProducer struct {
	Version string `json:"version"`
	Commit  string `json:"commit"`
}

type ChildLaunchComposition struct {
	Contract        string                            `json:"contract"`
	SchemaVersion   int                               `json:"schema_version"`
	Status          string                            `json:"status"`
	Producer        ChildLaunchCompositionProducer    `json:"producer"`
	Agent           string                            `json:"agent"`
	ProjectDir      string                            `json:"project_dir"`
	ArgvPrefix      []string                          `json:"argv_prefix"`
	MCPServers      []ChildLaunchCompositionMCPServer `json:"mcp_servers"`
	RequiredEnvVars []string                          `json:"required_env_vars"`
	Sources         ChildLaunchCompositionSources     `json:"sources"`
}

type ChildLaunchCompositionMCPServer struct {
	Name              string   `json:"name"`
	Transport         string   `json:"transport"`
	DefinitionSource  string   `json:"definition_source"`
	EnabledBy         []string `json:"enabled_by"`
	BearerTokenEnvVar string   `json:"bearer_token_env_var,omitempty"`
}

type ChildLaunchCompositionSources struct {
	ProjectConfigs []string `json:"project_configs"`
	Registries     []string `json:"registries"`
}

type ChildLaunchCompositionErrorEnvelope struct {
	Contract      string                         `json:"contract"`
	SchemaVersion int                            `json:"schema_version"`
	Status        string                         `json:"status"`
	Producer      ChildLaunchCompositionProducer `json:"producer"`
	Agent         string                         `json:"agent"`
	ProjectDir    string                         `json:"project_dir"`
	Error         ChildLaunchCompositionError    `json:"error"`
}

type ChildLaunchCompositionError struct {
	Code string `json:"code"`
}

func NewChildLaunchCompositionErrorEnvelope(agent, projectDir string, producer ChildLaunchCompositionProducer, code string) ChildLaunchCompositionErrorEnvelope {
	return ChildLaunchCompositionErrorEnvelope{
		Contract:      ChildLaunchCompositionContract,
		SchemaVersion: ChildLaunchCompositionSchemaVersion,
		Status:        "error",
		Producer:      producer,
		Agent:         agent,
		ProjectDir:    projectDir,
		Error:         ChildLaunchCompositionError{Code: code},
	}
}

func CanonicalProjectDir(projectDir string) (string, error) {
	if strings.TrimSpace(projectDir) == "" {
		return "", fmt.Errorf("project directory is required")
	}
	abs, err := filepath.Abs(projectDir)
	if err != nil {
		return "", fmt.Errorf("resolve project directory: %w", err)
	}
	resolved, err := filepath.EvalSymlinks(filepath.Clean(abs))
	if err != nil {
		return "", fmt.Errorf("resolve project directory %s: %w", abs, err)
	}
	info, err := os.Stat(resolved)
	if err != nil {
		return "", fmt.Errorf("stat project directory %s: %w", resolved, err)
	}
	if !info.IsDir() {
		return "", fmt.Errorf("project path is not a directory: %s", resolved)
	}
	return filepath.Clean(resolved), nil
}

func BuildChildLaunchComposition(agent, projectDir, homeDir string, producer ChildLaunchCompositionProducer) (ChildLaunchComposition, error) {
	if agent != "codex" && agent != "claude" {
		return ChildLaunchComposition{}, fmt.Errorf("unsupported agent %q", agent)
	}
	canonicalProjectDir, err := CanonicalProjectDir(projectDir)
	if err != nil {
		return ChildLaunchComposition{}, err
	}
	if homeDir == "" {
		homeDir, err = os.UserHomeDir()
		if err != nil {
			return ChildLaunchComposition{}, fmt.Errorf("resolve home directory: %w", err)
		}
	}
	homeDir, err = filepath.Abs(homeDir)
	if err != nil {
		return ChildLaunchComposition{}, fmt.Errorf("resolve home directory: %w", err)
	}

	composition := ChildLaunchComposition{
		Contract:        ChildLaunchCompositionContract,
		SchemaVersion:   ChildLaunchCompositionSchemaVersion,
		Status:          "ok",
		Producer:        producer,
		Agent:           agent,
		ProjectDir:      canonicalProjectDir,
		ArgvPrefix:      []string{},
		MCPServers:      []ChildLaunchCompositionMCPServer{},
		RequiredEnvVars: []string{},
		Sources: ChildLaunchCompositionSources{
			ProjectConfigs: []string{},
			Registries:     []string{},
		},
	}

	ancestors := ancestorDirsRootFirst(canonicalProjectDir)
	globalProjectConfigPath := filepath.Join(homeDir, ".agents", ".configs", projectConfigFileName)
	projectConfig, err := loadCompositeProjectConfig(ancestors, globalProjectConfigPath)
	if err != nil {
		return ChildLaunchComposition{}, err
	}
	for _, source := range projectConfig.Sources {
		composition.Sources.ProjectConfigs = append(composition.Sources.ProjectConfigs, source.Path)
	}

	definitions, registrySources, err := loadCompositeMCPRegistry(homeDir, ancestors)
	if err != nil {
		return ChildLaunchComposition{}, err
	}
	for _, source := range registrySources {
		composition.Sources.Registries = append(composition.Sources.Registries, source.Path)
	}

	requiredEnvSeen := map[string]bool{}
	claudeServers := map[string]claudeMCPConfigServer{}
	for _, name := range projectConfig.EnabledOrder {
		definition, ok := definitions[name]
		if !ok {
			return ChildLaunchComposition{}, fmt.Errorf("MCP server %q is enabled by %s but no definition was found in codex-mcp-servers.toml registries", name, strings.Join(projectConfig.EnabledBy[name], ", "))
		}
		if err := validateCodexMCPDefinition(name, definition); err != nil {
			return ChildLaunchComposition{}, err
		}

		transport := "stdio"
		if definition.Server.URL != "" {
			transport = "http"
		}
		composition.MCPServers = append(composition.MCPServers, ChildLaunchCompositionMCPServer{
			Name:              name,
			Transport:         transport,
			DefinitionSource:  definition.Source,
			EnabledBy:         append([]string(nil), projectConfig.EnabledBy[name]...),
			BearerTokenEnvVar: definition.Server.BearerTokenEnvVar,
		})
		if envName := definition.Server.BearerTokenEnvVar; envName != "" && !requiredEnvSeen[envName] {
			composition.RequiredEnvVars = append(composition.RequiredEnvVars, envName)
			requiredEnvSeen[envName] = true
		}

		server := CodexMCPLaunchServer{
			Name:              name,
			URL:               definition.Server.URL,
			BearerTokenEnvVar: definition.Server.BearerTokenEnvVar,
			Command:           definition.Server.Command,
			Args:              append([]string(nil), definition.Server.Args...),
			DefinitionSource:  definition.Source,
			EnabledBy:         append([]string(nil), projectConfig.EnabledBy[name]...),
		}
		if agent == "codex" {
			composition.ArgvPrefix = append(composition.ArgvPrefix, codexMCPConfigArgs(server)...)
		} else {
			claudeServers[name] = claudeMCPConfigServer(ClaudeMCPLaunchServer(server))
		}
	}

	if agent == "claude" && len(claudeServers) > 0 {
		configJSON, err := marshalClaudeMCPConfig(claudeServers)
		if err != nil {
			return ChildLaunchComposition{}, fmt.Errorf("encode Claude MCP config: %w", err)
		}
		composition.ArgvPrefix = []string{"--mcp-config", configJSON}
	}
	return composition, nil
}
