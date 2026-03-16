package infra

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestLocalLayout(t *testing.T) {
	layout, err := LocalLayout("/src/repo", "/tmp/project")
	if err != nil {
		t.Fatalf("LocalLayout: %v", err)
	}
	if layout.AgentsDir != "/tmp/project/.agents" {
		t.Fatalf("AgentsDir = %q", layout.AgentsDir)
	}
	if layout.ClaudeDir != "/tmp/project/.claude" {
		t.Fatalf("ClaudeDir = %q", layout.ClaudeDir)
	}
	if layout.CodexDir != "/tmp/project/.codex" {
		t.Fatalf("CodexDir = %q", layout.CodexDir)
	}
}

func TestSetupLocalCreatesInstalledRuntime(t *testing.T) {
	source := seedSourceRepo(t)
	project := t.TempDir()
	layout, err := LocalLayout(source, project)
	if err != nil {
		t.Fatalf("LocalLayout: %v", err)
	}

	var logs bytes.Buffer
	if err := Setup(Options{Layout: layout, Stdout: &logs}); err != nil {
		t.Fatalf("Setup: %v", err)
	}

	assertExists(t, filepath.Join(project, ".agents", ".instructions", "INSTRUCTIONS.md"))
	assertNoPath(t, filepath.Join(project, ".agents", ".git"))
	assertSymlink(t, filepath.Join(project, ".claude", "instructions"), filepath.Join(project, ".agents", ".instructions"))
	assertSymlink(t, filepath.Join(project, ".codex", "AGENTS.md"), filepath.Join(project, ".agents", ".instructions", "AGENTS.md"))
	assertSymlink(t, filepath.Join(project, ".local", "bin", "agents-attachments"), filepath.Join(project, ".agents", ".scripts", "agents-attachments"))

	launcher := filepath.Join(project, ".local", "bin", "agents-infra")
	data, err := os.ReadFile(launcher)
	if err != nil {
		t.Fatalf("ReadFile(%s): %v", launcher, err)
	}
	if !strings.Contains(string(data), source) {
		t.Fatalf("launcher does not reference source repo: %q", string(data))
	}

	claudeEntry := filepath.Join(project, ".claude", "CLAUDE.md")
	entry, err := os.ReadFile(claudeEntry)
	if err != nil {
		t.Fatalf("ReadFile(%s): %v", claudeEntry, err)
	}
	if !strings.Contains(string(entry), "@../.agents/.instructions/INSTRUCTIONS.md") {
		t.Fatalf("CLAUDE.md should reference local runtime relatively: %q", string(entry))
	}
}

func TestSyncSkipsGitAndTemp(t *testing.T) {
	source := seedSourceRepo(t)
	project := t.TempDir()
	layout, err := LocalLayout(source, project)
	if err != nil {
		t.Fatalf("LocalLayout: %v", err)
	}

	if err := Setup(Options{Layout: layout}); err != nil {
		t.Fatalf("Setup: %v", err)
	}

	assertNoPath(t, filepath.Join(project, ".agents", ".git"))
	assertNoPath(t, filepath.Join(project, ".agents", ".temp"))
	assertNoPath(t, filepath.Join(project, ".agents", ".gitignore"))
}

func TestDoctor(t *testing.T) {
	source := seedSourceRepo(t)
	project := t.TempDir()
	layout, err := LocalLayout(source, project)
	if err != nil {
		t.Fatalf("LocalLayout: %v", err)
	}
	if err := Setup(Options{Layout: layout}); err != nil {
		t.Fatalf("Setup: %v", err)
	}

	report := Doctor(layout)
	if !report.AgentsGitFree || !report.ClaudeLinked || !report.CodexLinked || !report.HelpersLinked || !report.InfraSkillLink {
		t.Fatalf("unexpected doctor report: %+v", report)
	}
}

func TestSetupRemovesGeneratedArtifacts(t *testing.T) {
	source := seedSourceRepo(t)
	project := t.TempDir()
	layout, err := LocalLayout(source, project)
	if err != nil {
		t.Fatalf("LocalLayout: %v", err)
	}

	mustMkdir(t, filepath.Join(project, ".agents", ".rules"))
	mustMkdir(t, filepath.Join(project, ".claude"))
	mustMkdir(t, filepath.Join(project, ".codex", "rules"))
	mustMkdir(t, filepath.Join(project, ".local", "bin"))

	mustWrite(t, filepath.Join(project, ".agents", ".rules", "default.rules.bak.1"), "stale")
	mustWrite(t, filepath.Join(project, ".agents", ".DS_Store"), "junk")
	mustWrite(t, filepath.Join(project, ".claude", "settings.json.bak.1"), "stale")
	mustWrite(t, filepath.Join(project, ".codex", "rules", "default.rules.bak.1"), "stale")
	mustWrite(t, filepath.Join(project, ".local", "bin", "agents-infra.bak.1"), "stale")

	if err := Setup(Options{Layout: layout}); err != nil {
		t.Fatalf("Setup: %v", err)
	}

	assertNoPath(t, filepath.Join(project, ".agents", ".rules", "default.rules.bak.1"))
	assertNoPath(t, filepath.Join(project, ".agents", ".DS_Store"))
	assertNoPath(t, filepath.Join(project, ".claude", "settings.json.bak.1"))
	assertNoPath(t, filepath.Join(project, ".codex", "rules", "default.rules.bak.1"))
	assertNoPath(t, filepath.Join(project, ".local", "bin", "agents-infra.bak.1"))
}

func TestSetupReplacesManagedPathsWithoutBackups(t *testing.T) {
	source := seedSourceRepo(t)
	project := t.TempDir()
	layout, err := LocalLayout(source, project)
	if err != nil {
		t.Fatalf("LocalLayout: %v", err)
	}

	mustMkdir(t, filepath.Join(project, ".claude"))
	mustMkdir(t, filepath.Join(project, ".codex", "rules"))
	mustMkdir(t, filepath.Join(project, ".local", "bin"))

	mustWrite(t, filepath.Join(project, ".claude", "settings.json"), "custom")
	mustWrite(t, filepath.Join(project, ".codex", "config.toml"), "custom")
	mustWrite(t, filepath.Join(project, ".codex", "rules", "default.rules"), "custom")
	mustWrite(t, filepath.Join(project, ".local", "bin", "agents-infra"), "#!/bin/sh\nexit 0\n")

	if err := Setup(Options{Layout: layout}); err != nil {
		t.Fatalf("Setup: %v", err)
	}

	assertSymlink(t, filepath.Join(project, ".claude", "settings.json"), filepath.Join(project, ".agents", ".configs", "claude-settings.json"))
	assertSymlink(t, filepath.Join(project, ".codex", "config.toml"), filepath.Join(project, ".agents", ".configs", "codex-config.toml"))
	assertSymlink(t, filepath.Join(project, ".codex", "rules", "default.rules"), filepath.Join(project, ".agents", ".rules", "default.rules"))
	assertNoGeneratedArtifacts(t, project)
}

func seedSourceRepo(t *testing.T) string {
	t.Helper()
	root := t.TempDir()
	mustMkdir(t, filepath.Join(root, ".instructions"))
	mustMkdir(t, filepath.Join(root, ".configs"))
	mustMkdir(t, filepath.Join(root, ".rules"))
	mustMkdir(t, filepath.Join(root, ".scripts"))
	mustMkdir(t, filepath.Join(root, ".skills", "skill-creator"))
	mustMkdir(t, filepath.Join(root, "tools", "agents-infra"))
	mustMkdir(t, filepath.Join(root, ".temp"))
	mustMkdir(t, filepath.Join(root, ".git"))

	mustWrite(t, filepath.Join(root, ".instructions", "INSTRUCTIONS.md"), "instructions")
	mustWrite(t, filepath.Join(root, ".instructions", "AGENTS.md"), "agents")
	mustWrite(t, filepath.Join(root, ".configs", "claude-settings.json"), "{}")
	mustWrite(t, filepath.Join(root, ".configs", "codex-config.toml"), "model = \"gpt-5.4\"")
	mustWrite(t, filepath.Join(root, ".rules", "default.rules"), "allow")
	mustWrite(t, filepath.Join(root, ".scripts", "agents-attachments"), "#!/bin/sh\nexit 0\n")
	mustWrite(t, filepath.Join(root, ".skills", "skill-creator", "SKILL.md"), "creator")
	mustWrite(t, filepath.Join(root, ".gitignore"), "ignored")
	mustWrite(t, filepath.Join(root, ".temp", "junk.txt"), "junk")
	mustWrite(t, filepath.Join(root, "tools", "agents-infra", "go.mod"), "module example\n")
	return root
}

func mustMkdir(t *testing.T, path string) {
	t.Helper()
	if err := os.MkdirAll(path, 0o755); err != nil {
		t.Fatalf("MkdirAll(%s): %v", path, err)
	}
}

func mustWrite(t *testing.T, path, body string) {
	t.Helper()
	if err := os.WriteFile(path, []byte(body), 0o755); err != nil {
		t.Fatalf("WriteFile(%s): %v", path, err)
	}
}

func assertExists(t *testing.T, path string) {
	t.Helper()
	if _, err := os.Stat(path); err != nil {
		t.Fatalf("expected %s to exist: %v", path, err)
	}
}

func assertNoPath(t *testing.T, path string) {
	t.Helper()
	if _, err := os.Lstat(path); err == nil {
		t.Fatalf("expected %s to be absent", path)
	}
}

func assertSymlink(t *testing.T, path, target string) {
	t.Helper()
	got, err := os.Readlink(path)
	if err != nil {
		t.Fatalf("Readlink(%s): %v", path, err)
	}
	if got != target {
		t.Fatalf("%s -> %s, want %s", path, got, target)
	}
}

func assertNoGeneratedArtifacts(t *testing.T, root string) {
	t.Helper()
	err := filepath.WalkDir(root, func(path string, d os.DirEntry, err error) error {
		if err != nil {
			return err
		}
		base := filepath.Base(path)
		if base == ".DS_Store" || strings.Contains(base, ".bak.") {
			t.Fatalf("unexpected generated artifact left behind: %s", path)
		}
		return nil
	})
	if err != nil {
		t.Fatalf("WalkDir(%s): %v", root, err)
	}
}
