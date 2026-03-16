package infra

import (
	"fmt"
	"io"
	"io/fs"
	"os"
	"path/filepath"
	"sort"
	"strings"
)

type Mode string

const (
	ModeGlobal Mode = "global"
	ModeLocal  Mode = "local"
)

type Layout struct {
	Mode      Mode
	SourceDir string
	RootDir   string
	AgentsDir string
	ClaudeDir string
	CodexDir  string
	BinDir    string
}

type Options struct {
	Layout Layout
	NoSync bool
	Stdout io.Writer
}

type Report struct {
	Layout         Layout
	AgentsGitFree  bool
	ClaudeLinked   bool
	CodexLinked    bool
	HelpersLinked  bool
	InfraSkillLink bool
}

func GlobalLayout(sourceDir, homeDir string) (Layout, error) {
	if homeDir == "" {
		return Layout{}, fmt.Errorf("home directory is required")
	}
	homeDir, err := filepath.Abs(homeDir)
	if err != nil {
		return Layout{}, fmt.Errorf("resolve home dir: %w", err)
	}
	return Layout{
		Mode:      ModeGlobal,
		SourceDir: sourceDir,
		RootDir:   homeDir,
		AgentsDir: filepath.Join(homeDir, ".agents"),
		ClaudeDir: filepath.Join(homeDir, ".claude"),
		CodexDir:  filepath.Join(homeDir, ".codex"),
		BinDir:    filepath.Join(homeDir, ".local", "bin"),
	}, nil
}

func LocalLayout(sourceDir, projectDir string) (Layout, error) {
	if projectDir == "" {
		return Layout{}, fmt.Errorf("project directory is required")
	}
	projectDir, err := filepath.Abs(projectDir)
	if err != nil {
		return Layout{}, fmt.Errorf("resolve project dir: %w", err)
	}
	return Layout{
		Mode:      ModeLocal,
		SourceDir: sourceDir,
		RootDir:   projectDir,
		AgentsDir: filepath.Join(projectDir, ".agents"),
		ClaudeDir: filepath.Join(projectDir, ".claude"),
		CodexDir:  filepath.Join(projectDir, ".codex"),
		BinDir:    filepath.Join(projectDir, ".local", "bin"),
	}, nil
}

func Setup(opts Options) error {
	if opts.Layout.SourceDir == "" {
		return fmt.Errorf("source dir is required")
	}
	if !opts.NoSync {
		if err := syncRepo(opts.Layout.SourceDir, opts.Layout.AgentsDir); err != nil {
			return err
		}
		logf(opts.Stdout, "Synced source repo into %s", opts.Layout.AgentsDir)
	}
	if err := scrubInstalledGitMetadata(opts.Layout, opts.Stdout); err != nil {
		return err
	}
	if err := scrubGeneratedArtifacts(opts.Layout, opts.Stdout); err != nil {
		return err
	}
	return RefreshLinks(opts)
}

func RefreshLinks(opts Options) error {
	if err := ensureRepoSkillLinks(opts.Layout, opts.Stdout); err != nil {
		return err
	}
	if err := setupClaude(opts.Layout, opts.Stdout); err != nil {
		return err
	}
	if err := setupCodex(opts.Layout, opts.Stdout); err != nil {
		return err
	}
	if err := setupHelpers(opts.Layout, opts.Stdout); err != nil {
		return err
	}
	return installCLIWrapper(opts.Layout, opts.Stdout)
}

func Doctor(layout Layout) Report {
	report := Report{Layout: layout}
	if _, err := os.Stat(filepath.Join(layout.AgentsDir, ".git")); err == nil {
		report.AgentsGitFree = false
	} else {
		report.AgentsGitFree = true
	}
	report.ClaudeLinked = isLinkTo(filepath.Join(layout.ClaudeDir, "instructions"), filepath.Join(layout.AgentsDir, ".instructions"))
	report.CodexLinked = isLinkTo(filepath.Join(layout.CodexDir, "AGENTS.md"), filepath.Join(layout.AgentsDir, ".instructions", "AGENTS.md"))
	report.HelpersLinked = isLinkTo(filepath.Join(layout.BinDir, "agents-attachments"), filepath.Join(layout.AgentsDir, ".scripts", "agents-attachments"))
	report.InfraSkillLink = isLinkTo(filepath.Join(layout.AgentsDir, "skills", "alexis-agents-infra"), layout.AgentsDir)
	return report
}

func logf(w io.Writer, format string, args ...any) {
	if w == nil {
		return
	}
	fmt.Fprintf(w, format+"\n", args...)
}

func syncRepo(sourceDir, agentsDir string) error {
	if err := os.MkdirAll(agentsDir, 0o755); err != nil {
		return fmt.Errorf("create agents dir: %w", err)
	}
	return filepath.WalkDir(sourceDir, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		rel, err := filepath.Rel(sourceDir, path)
		if err != nil {
			return err
		}
		if rel == "." {
			return nil
		}
		if shouldSkip(rel, d.IsDir()) {
			if d.IsDir() {
				return filepath.SkipDir
			}
			return nil
		}

		dst := filepath.Join(agentsDir, rel)
		info, err := d.Info()
		if err != nil {
			return err
		}

		if d.Type()&os.ModeSymlink != 0 {
			target, err := os.Readlink(path)
			if err != nil {
				return fmt.Errorf("read symlink %s: %w", path, err)
			}
			return forceSymlink(target, dst)
		}

		if d.IsDir() {
			return os.MkdirAll(dst, info.Mode())
		}

		if err := os.MkdirAll(filepath.Dir(dst), 0o755); err != nil {
			return err
		}
		return copyFile(path, dst, info.Mode())
	})
}

func shouldSkip(rel string, isDir bool) bool {
	rel = filepath.ToSlash(rel)
	base := filepath.Base(rel)
	switch {
	case rel == ".git" || strings.HasPrefix(rel, ".git/"):
		return true
	case rel == ".temp" || strings.HasPrefix(rel, ".temp/"):
		return true
	case base == ".DS_Store", base == ".skill-lock.json", base == ".gitignore", base == ".gitattributes", base == ".gitmodules":
		return true
	default:
		return false
	}
}

func scrubInstalledGitMetadata(layout Layout, out io.Writer) error {
	if samePath(layout.SourceDir, layout.AgentsDir) {
		logf(out, "Skipping git metadata cleanup because source dir equals install dir: %s", layout.AgentsDir)
		return nil
	}
	removed := 0
	for _, rel := range []string{".git", ".gitignore", ".gitattributes", ".gitmodules"} {
		path := filepath.Join(layout.AgentsDir, rel)
		if _, err := os.Lstat(path); err == nil {
			if err := os.RemoveAll(path); err != nil {
				return fmt.Errorf("remove %s: %w", path, err)
			}
			removed++
			logf(out, "Removed installed git metadata: %s", path)
		}
	}
	if removed == 0 {
		logf(out, "Installed git metadata already absent from %s", layout.AgentsDir)
	}
	return nil
}

func scrubGeneratedArtifacts(layout Layout, out io.Writer) error {
	roots := []string{layout.ClaudeDir, layout.CodexDir, layout.BinDir}
	if !samePath(layout.SourceDir, layout.AgentsDir) {
		roots = append([]string{layout.AgentsDir}, roots...)
	} else {
		logf(out, "Skipping generated artifact cleanup in source dir: %s", layout.AgentsDir)
	}
	for _, root := range roots {
		if err := removeGeneratedArtifacts(root, out); err != nil {
			return err
		}
	}
	return nil
}

func removeGeneratedArtifacts(root string, out io.Writer) error {
	if _, err := os.Lstat(root); os.IsNotExist(err) {
		return nil
	} else if err != nil {
		return fmt.Errorf("stat %s: %w", root, err)
	}

	var paths []string
	if err := filepath.WalkDir(root, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if path == root {
			return nil
		}
		if isGeneratedArtifact(filepath.Base(path)) {
			paths = append(paths, path)
		}
		return nil
	}); err != nil {
		return fmt.Errorf("walk %s: %w", root, err)
	}

	sort.Slice(paths, func(i, j int) bool {
		return len(paths[i]) > len(paths[j])
	})
	for _, path := range paths {
		if _, err := os.Lstat(path); os.IsNotExist(err) {
			continue
		} else if err != nil {
			return fmt.Errorf("stat generated artifact %s: %w", path, err)
		}
		if err := os.RemoveAll(path); err != nil {
			return fmt.Errorf("remove generated artifact %s: %w", path, err)
		}
		logf(out, "Removed generated artifact: %s", path)
	}
	return nil
}

func ensureRepoSkillLinks(layout Layout, out io.Writer) error {
	skillsDir := filepath.Join(layout.AgentsDir, "skills")
	if err := os.MkdirAll(skillsDir, 0o755); err != nil {
		return fmt.Errorf("create skills dir: %w", err)
	}
	if err := createSymlink(layout.AgentsDir, filepath.Join(skillsDir, "alexis-agents-infra"), out); err != nil {
		return err
	}
	skillCreator := filepath.Join(layout.AgentsDir, ".skills", "skill-creator")
	if st, err := os.Stat(skillCreator); err == nil && st.IsDir() {
		if err := createSymlink(skillCreator, filepath.Join(skillsDir, "skill-creator"), out); err != nil {
			return err
		}
	}
	return nil
}

func setupClaude(layout Layout, out io.Writer) error {
	if err := os.MkdirAll(filepath.Join(layout.ClaudeDir, "skills"), 0o755); err != nil {
		return err
	}
	if err := createSymlink(filepath.Join(layout.AgentsDir, ".instructions"), filepath.Join(layout.ClaudeDir, "instructions"), out); err != nil {
		return err
	}
	if err := writeClaudeEntrypoint(layout); err != nil {
		return err
	}
	skillsDir := filepath.Join(layout.AgentsDir, "skills")
	entries, err := os.ReadDir(skillsDir)
	if err != nil {
		return err
	}
	for _, entry := range entries {
		if shouldIgnoreGeneratedEntry(entry.Name()) {
			continue
		}
		if err := createSymlink(filepath.Join(skillsDir, entry.Name()), filepath.Join(layout.ClaudeDir, "skills", entry.Name()), out); err != nil {
			return err
		}
	}
	return createSymlink(filepath.Join(layout.AgentsDir, ".configs", "claude-settings.json"), filepath.Join(layout.ClaudeDir, "settings.json"), out)
}

func setupCodex(layout Layout, out io.Writer) error {
	if err := os.MkdirAll(filepath.Join(layout.CodexDir, "skills"), 0o755); err != nil {
		return err
	}
	if err := os.MkdirAll(filepath.Join(layout.CodexDir, "rules"), 0o755); err != nil {
		return err
	}
	if err := createSymlink(filepath.Join(layout.AgentsDir, ".instructions", "AGENTS.md"), filepath.Join(layout.CodexDir, "AGENTS.md"), out); err != nil {
		return err
	}
	skillsDir := filepath.Join(layout.AgentsDir, "skills")
	entries, err := os.ReadDir(skillsDir)
	if err != nil {
		return err
	}
	for _, entry := range entries {
		if shouldIgnoreGeneratedEntry(entry.Name()) {
			continue
		}
		systemSkill := filepath.Join(layout.CodexDir, "skills", ".system", entry.Name())
		if st, err := os.Stat(systemSkill); err == nil && st.IsDir() {
			logf(out, "Skipping %s because it exists in %s", entry.Name(), systemSkill)
			continue
		}
		if err := createSymlink(filepath.Join(skillsDir, entry.Name()), filepath.Join(layout.CodexDir, "skills", entry.Name()), out); err != nil {
			return err
		}
	}
	if err := createSymlink(filepath.Join(layout.AgentsDir, ".configs", "codex-config.toml"), filepath.Join(layout.CodexDir, "config.toml"), out); err != nil {
		return err
	}
	rulesDir := filepath.Join(layout.AgentsDir, ".rules")
	rules, err := os.ReadDir(rulesDir)
	if err != nil {
		return err
	}
	for _, entry := range rules {
		if entry.IsDir() || shouldIgnoreGeneratedEntry(entry.Name()) {
			continue
		}
		if err := createSymlink(filepath.Join(rulesDir, entry.Name()), filepath.Join(layout.CodexDir, "rules", entry.Name()), out); err != nil {
			return err
		}
	}
	return nil
}

func setupHelpers(layout Layout, out io.Writer) error {
	if err := os.MkdirAll(layout.BinDir, 0o755); err != nil {
		return err
	}
	if err := createSymlink(filepath.Join(layout.AgentsDir, ".scripts", "agents-attachments"), filepath.Join(layout.BinDir, "agents-attachments"), out); err != nil {
		return err
	}
	return nil
}

func installCLIWrapper(layout Layout, out io.Writer) error {
	if err := os.MkdirAll(layout.BinDir, 0o755); err != nil {
		return err
	}
	path := filepath.Join(layout.BinDir, "agents-infra")
	sourceDir := layout.SourceDir
	if sourceDir == "" {
		sourceDir = layout.AgentsDir
	}
	body := fmt.Sprintf(`#!/usr/bin/env sh
set -eu
export AGENTS_INFRA_SOURCE_DIR=%q
cd "$AGENTS_INFRA_SOURCE_DIR/tools/agents-infra"
exec go run . "$@"
`, sourceDir)
	if existing, err := os.ReadFile(path); err == nil && string(existing) == body {
		logf(out, "CLI launcher already up to date: %s", path)
		return nil
	}
	if err := removeManagedPath(path, out); err != nil {
		return err
	}
	if err := os.WriteFile(path, []byte(body), 0o755); err != nil {
		return fmt.Errorf("write cli wrapper: %w", err)
	}
	logf(out, "Installed CLI launcher: %s", path)
	return nil
}

func writeClaudeEntrypoint(layout Layout) error {
	ref, err := filepath.Rel(layout.ClaudeDir, filepath.Join(layout.AgentsDir, ".instructions", "INSTRUCTIONS.md"))
	if err != nil || strings.HasPrefix(ref, ".."+string(filepath.Separator)+"..") {
		ref = filepath.Join(layout.AgentsDir, ".instructions", "INSTRUCTIONS.md")
	}
	ref = filepath.ToSlash(ref)
	body := fmt.Sprintf("# Claude Instructions\n\nLoad all instructions from the installed agents runtime:\n\n@%s\n", ref)
	return os.WriteFile(filepath.Join(layout.ClaudeDir, "CLAUDE.md"), []byte(body), 0o644)
}

func createSymlink(target, link string, out io.Writer) error {
	if err := os.MkdirAll(filepath.Dir(link), 0o755); err != nil {
		return err
	}
	if existingTarget, err := os.Readlink(link); err == nil {
		if existingTarget == target {
			return nil
		}
		if err := os.Remove(link); err != nil {
			return err
		}
		logf(out, "Removed existing symlink: %s", link)
	} else if !os.IsNotExist(err) {
		if err := removeManagedPath(link, out); err != nil {
			return err
		}
	}
	if err := os.Symlink(target, link); err != nil {
		return fmt.Errorf("create symlink %s -> %s: %w", link, target, err)
	}
	logf(out, "Created symlink: %s -> %s", link, target)
	return nil
}

func removeManagedPath(path string, out io.Writer) error {
	st, err := os.Lstat(path)
	if os.IsNotExist(err) {
		return nil
	} else if err != nil {
		return err
	}
	if st.IsDir() && st.Mode()&os.ModeSymlink == 0 {
		if err := os.RemoveAll(path); err != nil {
			return err
		}
		logf(out, "Removed existing directory: %s", path)
		return nil
	}
	if err := os.Remove(path); err != nil {
		return err
	}
	logf(out, "Removed existing managed path: %s", path)
	return nil
}

func forceSymlink(target, path string) error {
	if existingTarget, err := os.Readlink(path); err == nil {
		if existingTarget == target {
			return nil
		}
		if err := os.Remove(path); err != nil {
			return err
		}
	} else if !os.IsNotExist(err) {
		if err := os.RemoveAll(path); err != nil {
			return err
		}
	}
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return err
	}
	return os.Symlink(target, path)
}

func copyFile(src, dst string, mode fs.FileMode) error {
	in, err := os.Open(src)
	if err != nil {
		return err
	}
	defer in.Close()
	if st, err := os.Lstat(dst); err == nil {
		if st.IsDir() {
			if err := os.RemoveAll(dst); err != nil {
				return err
			}
		} else if st.Mode()&os.ModeSymlink != 0 {
			if err := os.Remove(dst); err != nil {
				return err
			}
		}
	}
	out, err := os.OpenFile(dst, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, mode)
	if err != nil {
		return err
	}
	defer out.Close()
	if _, err := io.Copy(out, in); err != nil {
		return err
	}
	return out.Close()
}

func samePath(a, b string) bool {
	aa, errA := filepath.Abs(a)
	bb, errB := filepath.Abs(b)
	if errA != nil || errB != nil {
		return a == b
	}
	return aa == bb
}

func isLinkTo(path, target string) bool {
	got, err := os.Readlink(path)
	if err != nil {
		return false
	}
	return got == target
}

func shouldIgnoreGeneratedEntry(name string) bool {
	return strings.HasPrefix(name, ".") || strings.Contains(name, ".bak.")
}

func isGeneratedArtifact(name string) bool {
	return name == ".DS_Store" || strings.Contains(name, ".bak.")
}
