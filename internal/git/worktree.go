package git

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

// Worktree represents a single git worktree.
type Worktree struct {
	Path   string
	Branch string
	HEAD   string
	IsBare bool
	IsMain bool // first worktree = main repository
}

// FindRepoRoot returns the top-level directory of the git repository.
func FindRepoRoot() (string, error) {
	out, err := exec.Command("git", "rev-parse", "--show-toplevel").CombinedOutput()
	if err != nil {
		return "", fmt.Errorf("not a git repository: %w", err)
	}
	return strings.TrimSpace(string(out)), nil
}

// List returns all worktrees for the current repository.
func List() ([]Worktree, error) {
	out, err := exec.Command("git", "worktree", "list", "--porcelain").CombinedOutput()
	if err != nil {
		return nil, fmt.Errorf("git worktree list: %w\n%s", err, out)
	}
	return parsePorcelain(string(out)), nil
}

func parsePorcelain(raw string) []Worktree {
	var result []Worktree
	var current Worktree

	for _, line := range strings.Split(raw, "\n") {
		line = strings.TrimSpace(line)
		switch {
		case strings.HasPrefix(line, "worktree "):
			current = Worktree{Path: strings.TrimPrefix(line, "worktree ")}
		case strings.HasPrefix(line, "HEAD "):
			current.HEAD = strings.TrimPrefix(line, "HEAD ")
		case strings.HasPrefix(line, "branch "):
			ref := strings.TrimPrefix(line, "branch ")
			current.Branch = strings.TrimPrefix(ref, "refs/heads/")
		case line == "bare":
			current.IsBare = true
		case line == "":
			if current.Path != "" {
				result = append(result, current)
				current = Worktree{}
			}
		}
	}
	if current.Path != "" {
		result = append(result, current)
	}
	if len(result) > 0 {
		result[0].IsMain = true
	}
	return result
}

// Add creates a new worktree.
// Auto-detects whether the branch exists locally, on remote, or needs to be created.
// baseBranch is used as the starting point for new branches (e.g. "origin/release", "main").
func Add(basePath, branchName, baseBranch string) error {
	wtPath := filepath.Join(basePath, branchName)

	newBranch := false
	args := []string{"worktree", "add"}
	if BranchExists(branchName) {
		// Branch exists locally — just check it out
		args = append(args, wtPath, branchName)
	} else if remoteRef := findRemoteRef(branchName); remoteRef != "" {
		// Branch exists on remote — create local tracking branch
		args = append(args, "-b", branchName, wtPath, remoteRef)
	} else {
		// New branch — create from baseBranch
		newBranch = true
		args = append(args, "-b", branchName, wtPath)
		if baseBranch != "" && baseBranch != "HEAD" {
			args = append(args, baseBranch)
		}
	}

	out, err := exec.Command("git", args...).CombinedOutput()
	if err != nil {
		return fmt.Errorf("git worktree add: %w\n%s", err, strings.TrimSpace(string(out)))
	}

	// New branch created from another base (e.g. origin/release):
	// unset inherited upstream so push goes to origin/<branchName>, not the base
	if newBranch {
		_ = exec.Command("git", "-C", wtPath, "branch", "--unset-upstream").Run()
	}

	return nil
}

// Remove fully cleans up a worktree: removes worktree, deletes the directory,
// prunes stale references, and deletes the branch.
func Remove(wtPath, branch string, force bool) error {
	args := []string{"worktree", "remove"}
	if force {
		args = append(args, "--force")
	}
	args = append(args, wtPath)

	out, err := exec.Command("git", args...).CombinedOutput()
	if err != nil {
		_ = os.RemoveAll(wtPath)
		_ = exec.Command("git", "worktree", "prune").Run()
	} else {
		if _, statErr := os.Stat(wtPath); statErr == nil {
			_ = os.RemoveAll(wtPath)
		}
	}

	_ = exec.Command("git", "worktree", "prune").Run()

	if branch != "" && branch != "main" && branch != "master" {
		delArgs := []string{"branch", "-D", branch}
		if delOut, delErr := exec.Command("git", delArgs...).CombinedOutput(); delErr != nil {
			_ = delOut
		}
	}

	if err != nil {
		return fmt.Errorf("git worktree remove: %w\n%s", err, strings.TrimSpace(string(out)))
	}
	return nil
}

// findRemoteRef looks for a remote tracking branch matching the given short name.
// Returns e.g. "origin/feature/foo" or "" if not found.
func findRemoteRef(shortName string) string {
	for _, b := range ListRemoteBranches() {
		// Strip remote prefix (e.g. "origin/") and compare
		short := b
		if idx := strings.Index(b, "/"); idx >= 0 {
			short = b[idx+1:]
		}
		if short == shortName {
			return b
		}
	}
	return ""
}

// BranchExists checks if a branch already exists.
func BranchExists(name string) bool {
	err := exec.Command("git", "rev-parse", "--verify", "refs/heads/"+name).Run()
	return err == nil
}

// ShortStatus returns a brief git status for a worktree path.
func ShortStatus(wtPath string) (string, error) {
	cmd := exec.Command("git", "-C", wtPath, "status", "--short")
	out, err := cmd.CombinedOutput()
	if err != nil {
		return "", err
	}
	lines := strings.Split(strings.TrimSpace(string(out)), "\n")
	if len(lines) == 1 && lines[0] == "" {
		return "clean", nil
	}
	return fmt.Sprintf("%d changed", len(lines)), nil
}

// SyncStatus describes how a branch relates to its upstream.
type SyncStatus struct {
	Ahead    int
	Behind   int
	Upstream string // e.g. "origin/main"
	NoRemote bool   // no tracking branch set
}

// Summary returns a human-readable sync status.
func (s SyncStatus) Summary() string {
	if s.NoRemote {
		return "no remote"
	}
	if s.Ahead == 0 && s.Behind == 0 {
		return "up to date"
	}
	parts := ""
	if s.Behind > 0 {
		parts += fmt.Sprintf("↓%d behind", s.Behind)
	}
	if s.Ahead > 0 {
		if parts != "" {
			parts += " "
		}
		parts += fmt.Sprintf("↑%d ahead", s.Ahead)
	}
	return parts
}

// Short returns a compact indicator for the list view.
func (s SyncStatus) Short() string {
	if s.NoRemote {
		return ""
	}
	if s.Ahead == 0 && s.Behind == 0 {
		return "✓"
	}
	parts := ""
	if s.Behind > 0 {
		parts += fmt.Sprintf("↓%d", s.Behind)
	}
	if s.Ahead > 0 {
		parts += fmt.Sprintf("↑%d", s.Ahead)
	}
	return parts
}

// GetSyncStatus returns ahead/behind status for a worktree branch vs upstream.
func GetSyncStatus(wtPath string) SyncStatus {
	// Get tracking branch
	upstreamCmd := exec.Command("git", "-C", wtPath, "rev-parse", "--abbrev-ref", "--symbolic-full-name", "@{u}")
	upstreamOut, err := upstreamCmd.CombinedOutput()
	if err != nil {
		return SyncStatus{NoRemote: true}
	}
	upstream := strings.TrimSpace(string(upstreamOut))
	if upstream == "" {
		return SyncStatus{NoRemote: true}
	}

	// Get ahead/behind counts
	cmd := exec.Command("git", "-C", wtPath, "rev-list", "--left-right", "--count", "HEAD...@{u}")
	out, err := cmd.CombinedOutput()
	if err != nil {
		return SyncStatus{Upstream: upstream, NoRemote: false}
	}

	parts := strings.Fields(strings.TrimSpace(string(out)))
	if len(parts) != 2 {
		return SyncStatus{Upstream: upstream}
	}

	ahead := 0
	behind := 0
	fmt.Sscanf(parts[0], "%d", &ahead)
	fmt.Sscanf(parts[1], "%d", &behind)

	return SyncStatus{
		Ahead:    ahead,
		Behind:   behind,
		Upstream: upstream,
	}
}

// IsDirty checks if a worktree has uncommitted changes.
func IsDirty(wtPath string) bool {
	s, err := ShortStatus(wtPath)
	if err != nil {
		return false
	}
	return s != "clean"
}

// ListBranches returns all local branch names, sorted.
func ListBranches() []string {
	out, err := exec.Command("git", "branch", "--format=%(refname:short)").CombinedOutput()
	if err != nil {
		return nil
	}
	var branches []string
	for _, line := range strings.Split(strings.TrimSpace(string(out)), "\n") {
		line = strings.TrimSpace(line)
		if line != "" {
			branches = append(branches, line)
		}
	}
	return branches
}

// ListRemoteBranches returns remote branch names as-is (e.g. "origin/main").
func ListRemoteBranches() []string {
	out, err := exec.Command("git", "branch", "-r", "--format=%(refname:short)").CombinedOutput()
	if err != nil {
		return nil
	}
	var branches []string
	for _, line := range strings.Split(strings.TrimSpace(string(out)), "\n") {
		line = strings.TrimSpace(line)
		if line == "" || strings.Contains(line, "HEAD") || line == "origin" {
			continue
		}
		branches = append(branches, line)
	}
	return branches
}

// BranchOrigin indicates where a branch exists.
type BranchOrigin int

const (
	BranchLocal      BranchOrigin = iota // local only
	BranchRemote                         // remote only
	BranchBoth                           // local + remote
)

// BranchInfo holds a branch name and its origin.
type BranchInfo struct {
	Name   string
	Origin BranchOrigin
}

// AllBranches returns merged list of local + remote branches with origin info.
func AllBranches() []BranchInfo {
	localSet := make(map[string]bool)
	for _, b := range ListBranches() {
		localSet[b] = true
	}

	remoteSet := make(map[string]bool)
	remoteMap := make(map[string]string) // short name → full remote name
	for _, b := range ListRemoteBranches() {
		// Strip "origin/" to get short name for matching
		short := b
		if idx := strings.Index(b, "/"); idx >= 0 {
			short = b[idx+1:]
		}
		remoteSet[short] = true
		remoteMap[short] = b
	}

	seen := make(map[string]bool)
	var result []BranchInfo

	// Local branches first
	for _, b := range ListBranches() {
		origin := BranchLocal
		if remoteSet[b] {
			origin = BranchBoth
		}
		result = append(result, BranchInfo{Name: b, Origin: origin})
		seen[b] = true
	}

	// Remote-only branches
	for _, b := range ListRemoteBranches() {
		short := b
		if idx := strings.Index(b, "/"); idx >= 0 {
			short = b[idx+1:]
		}
		if seen[short] {
			continue
		}
		seen[short] = true
		result = append(result, BranchInfo{Name: b, Origin: BranchRemote})
	}

	return result
}

// Pull runs git pull in the given worktree.
func Pull(wtPath string) error {
	cmd := exec.Command("git", "-C", wtPath, "pull")
	out, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("git pull: %w\n%s", err, strings.TrimSpace(string(out)))
	}
	return nil
}

// Push runs git push origin HEAD in the given worktree.
func Push(wtPath string) error {
	cmd := exec.Command("git", "-C", wtPath, "push", "origin", "HEAD")
	out, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("git push: %w\n%s", err, strings.TrimSpace(string(out)))
	}
	return nil
}

// Fetch runs git fetch in the given worktree.
func Fetch(wtPath string) error {
	cmd := exec.Command("git", "-C", wtPath, "fetch")
	out, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("git fetch: %w\n%s", err, strings.TrimSpace(string(out)))
	}
	return nil
}
