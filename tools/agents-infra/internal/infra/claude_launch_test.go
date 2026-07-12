package infra

import (
	"encoding/json"
	"path/filepath"
	"reflect"
	"strings"
	"testing"
)

func TestBuildClaudeLaunchPlanComposesAncestorConfigsAndProvenance(t *testing.T) {
	home := t.TempDir()
	root := t.TempDir()
	parent := filepath.Join(root, "parent")
	child := filepath.Join(parent, "child")
	mustMkdir(t, child)

	mustMkdir(t, filepath.Join(home, ".agents", ".configs"))
	mustWrite(t, filepath.Join(home, ".agents", ".configs", "codex-mcp-servers.toml"), "[servers.figma]\nurl = \"https://global.example/figma\"\n")
	mustMkdir(t, filepath.Join(parent, ".agents", ".configs"))
	mustWrite(t, filepath.Join(parent, ".agents", ".configs", "project-config.toml"), "[mcp]\nenabled_servers = [\"figma\"]\n")
	mustWrite(t, filepath.Join(parent, ".agents", ".configs", "codex-mcp-servers.toml"), "[servers.figma]\nurl = \"https://parent.example/figma\"\n")
	mustMkdir(t, filepath.Join(child, ".agents", ".configs"))
	mustWrite(t, filepath.Join(child, ".agents", ".configs", "project-config.toml"), "[mcp]\nenabled_servers = [\"jira\", \"figma\"]\n")
	mustWrite(t, filepath.Join(child, ".agents", ".configs", "codex-mcp-servers.toml"), "[servers.jira]\nurl = \"https://child.example/jira\"\nbearer_token_env_var = \"JIRA_TOKEN\"\n")

	plan, err := BuildClaudeLaunchPlan(child, home, []string{"-d", "-"})
	if err != nil {
		t.Fatalf("BuildClaudeLaunchPlan: %v", err)
	}

	if len(plan.MCPServers) != 2 {
		t.Fatalf("MCPServers = %#v, want 2 entries", plan.MCPServers)
	}
	if plan.MCPServers[0].Name != "figma" || plan.MCPServers[0].URL != "https://parent.example/figma" {
		t.Fatalf("figma server = %#v", plan.MCPServers[0])
	}
	if len(plan.MCPServers[0].EnabledBy) != 2 {
		t.Fatalf("figma EnabledBy = %#v, want parent and child configs", plan.MCPServers[0].EnabledBy)
	}
	if plan.MCPServers[1].Name != "jira" || plan.MCPServers[1].BearerTokenEnvVar != "JIRA_TOKEN" {
		t.Fatalf("jira server = %#v", plan.MCPServers[1])
	}

	wantArgs := []string{
		"--mcp-config", plan.MCPConfigJSON,
		claudeDangerouslySkipPermissions,
		"-",
	}
	if !reflect.DeepEqual(plan.Args, wantArgs) {
		t.Fatalf("Args = %#v, want %#v", plan.Args, wantArgs)
	}

	var decoded struct {
		MCPServers map[string]struct {
			Type    string            `json:"type"`
			URL     string            `json:"url"`
			Headers map[string]string `json:"headers"`
		} `json:"mcpServers"`
	}
	if err := json.Unmarshal([]byte(plan.MCPConfigJSON), &decoded); err != nil {
		t.Fatalf("unmarshal MCPConfigJSON: %v", err)
	}
	if decoded.MCPServers["figma"].Type != "http" || decoded.MCPServers["figma"].URL != "https://parent.example/figma" {
		t.Fatalf("figma config = %#v", decoded.MCPServers["figma"])
	}
	if decoded.MCPServers["jira"].Headers["Authorization"] != "Bearer ${JIRA_TOKEN}" {
		t.Fatalf("jira headers = %#v", decoded.MCPServers["jira"].Headers)
	}

	rendered := RenderClaudeLaunchPlan(plan)
	for _, want := range []string{
		"agents-infra claude config",
		"enabled_mcp=figma",
		"enabled_mcp=jira,figma",
		"definition: " + filepath.Join(parent, ".agents", ".configs", "codex-mcp-servers.toml"),
		"definition: " + filepath.Join(child, ".agents", ".configs", "codex-mcp-servers.toml"),
		"-d => " + claudeDangerouslySkipPermissions,
	} {
		if !strings.Contains(rendered, want) {
			t.Fatalf("rendered plan missing %q:\n%s", want, rendered)
		}
	}
}

func TestBuildClaudeLaunchPlanIgnoresHomeAgentsProjectConfigWithoutProjectOptIn(t *testing.T) {
	home := t.TempDir()
	start := filepath.Join(home, "project", "subdir")
	mustMkdir(t, start)
	mustMkdir(t, filepath.Join(home, ".agents", ".configs"))
	mustWrite(t, filepath.Join(home, ".agents", ".configs", "codex-mcp-servers.toml"), "[servers.figma]\nurl = \"https://global.example/figma\"\n")
	mustWrite(t, filepath.Join(home, ".agents", ".configs", "project-config.toml"), "[mcp]\nenabled_servers = [\"figma\"]\n")

	plan, err := BuildClaudeLaunchPlan(start, home, []string{"--resume"})
	if err != nil {
		t.Fatalf("BuildClaudeLaunchPlan: %v", err)
	}
	if len(plan.MCPServers) != 0 || len(plan.ConfigArgs) != 0 {
		t.Fatalf("home agents registry/config should not enable MCP without project opt-in: %+v", plan)
	}
	wantArgs := []string{"--resume"}
	if !reflect.DeepEqual(plan.Args, wantArgs) {
		t.Fatalf("Args = %#v, want %#v", plan.Args, wantArgs)
	}
	if !strings.Contains(RenderClaudeLaunchPlan(plan), "enabled_mcp:\n  - (none)") {
		t.Fatalf("rendered plan should report no enabled MCP:\n%s", RenderClaudeLaunchPlan(plan))
	}
}

func TestBuildClaudeLaunchPlanSupportsStdioMCPServers(t *testing.T) {
	home := t.TempDir()
	start := t.TempDir()
	mustMkdir(t, filepath.Join(home, ".agents", ".configs"))
	mustWrite(t, filepath.Join(home, ".agents", ".configs", "codex-mcp-servers.toml"), "[servers.lldb]\ncommand = \"lldb-mcp\"\nargs = [\"--socket\", \"auto\"]\n")
	mustMkdir(t, filepath.Join(start, ".agents", ".configs"))
	mustWrite(t, filepath.Join(start, ".agents", ".configs", "project-config.toml"), "[mcp]\nenabled_servers = [\"lldb\"]\n")

	plan, err := BuildClaudeLaunchPlan(start, home, nil)
	if err != nil {
		t.Fatalf("BuildClaudeLaunchPlan: %v", err)
	}

	if len(plan.MCPServers) != 1 {
		t.Fatalf("MCPServers = %#v, want 1 entry", plan.MCPServers)
	}
	server := plan.MCPServers[0]
	if server.Name != "lldb" || server.Command != "lldb-mcp" || !reflect.DeepEqual(server.Args, []string{"--socket", "auto"}) {
		t.Fatalf("lldb server = %#v", server)
	}

	var decoded struct {
		MCPServers map[string]struct {
			Type    string   `json:"type"`
			Command string   `json:"command"`
			Args    []string `json:"args"`
		} `json:"mcpServers"`
	}
	if err := json.Unmarshal([]byte(plan.MCPConfigJSON), &decoded); err != nil {
		t.Fatalf("unmarshal MCPConfigJSON: %v", err)
	}
	if decoded.MCPServers["lldb"].Type != "stdio" || decoded.MCPServers["lldb"].Command != "lldb-mcp" {
		t.Fatalf("lldb config = %#v", decoded.MCPServers["lldb"])
	}

	rendered := RenderClaudeLaunchPlan(plan)
	for _, want := range []string{
		"enabled_mcp:\n  - lldb",
		"command: lldb-mcp",
		"args: [\"--socket\", \"auto\"]",
	} {
		if !strings.Contains(rendered, want) {
			t.Fatalf("rendered plan missing %q:\n%s", want, rendered)
		}
	}
}

func TestBuildClaudeLaunchPlanPrintConfigStopsWrapperParsingAfterSeparator(t *testing.T) {
	home := t.TempDir()
	start := t.TempDir()

	plan, err := BuildClaudeLaunchPlan(start, home, []string{"--print-config", "--", "-d"})
	if err != nil {
		t.Fatalf("BuildClaudeLaunchPlan: %v", err)
	}
	if !plan.PrintConfig {
		t.Fatalf("PrintConfig = false, want true")
	}
	wantArgs := []string{"-d"}
	if !reflect.DeepEqual(plan.Args, wantArgs) {
		t.Fatalf("Args = %#v, want %#v", plan.Args, wantArgs)
	}
	if len(plan.WrapperExpandedShortcuts) != 0 {
		t.Fatalf("WrapperExpandedShortcuts = %#v, want none", plan.WrapperExpandedShortcuts)
	}
}

func TestBuildClaudeLaunchPlanFailsOnUnknownEnabledMCP(t *testing.T) {
	home := t.TempDir()
	start := t.TempDir()
	mustMkdir(t, filepath.Join(start, ".agents", ".configs"))
	projectConfig := filepath.Join(start, ".agents", ".configs", "project-config.toml")
	mustWrite(t, projectConfig, "[mcp]\nenabled_servers = [\"missing\"]\n")

	_, err := BuildClaudeLaunchPlan(start, home, nil)
	if err == nil {
		t.Fatal("expected unknown enabled MCP to fail")
	}
	if !strings.Contains(err.Error(), "MCP server \"missing\"") || !strings.Contains(err.Error(), projectConfig) {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestSharedMCPOptInDrivesBothCodexAndClaudeIdentically(t *testing.T) {
	home := t.TempDir()
	start := t.TempDir()
	mustMkdir(t, filepath.Join(start, ".agents", ".configs"))
	mustWrite(t, filepath.Join(start, ".agents", ".configs", "codex-mcp-servers.toml"), "[servers.figma]\nurl = \"https://global.example/figma\"\n[servers.lldb]\ncommand = \"lldb-mcp\"\n")
	mustWrite(t, filepath.Join(start, ".agents", ".configs", "project-config.toml"), "[mcp]\nenabled_servers = [\"figma\", \"lldb\"]\n")

	codexPlan, err := BuildCodexLaunchPlan(start, home, nil)
	if err != nil {
		t.Fatalf("BuildCodexLaunchPlan: %v", err)
	}
	claudePlan, err := BuildClaudeLaunchPlan(start, home, nil)
	if err != nil {
		t.Fatalf("BuildClaudeLaunchPlan: %v", err)
	}

	if len(codexPlan.MCPServers) != 2 || len(claudePlan.MCPServers) != 2 {
		t.Fatalf("expected the single [mcp] opt-in to enable both servers for both agents: codex=%#v claude=%#v", codexPlan.MCPServers, claudePlan.MCPServers)
	}
	for i := range codexPlan.MCPServers {
		if codexPlan.MCPServers[i].Name != claudePlan.MCPServers[i].Name {
			t.Fatalf("codex/claude enabled server name mismatch at %d: %q vs %q", i, codexPlan.MCPServers[i].Name, claudePlan.MCPServers[i].Name)
		}
	}
}
