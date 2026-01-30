package git

import (
	"context"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"
)

var (
	ErrNotRepo    = errors.New("not a git repository")
	ErrGitMissing = errors.New("git executable not found")
)

const defaultTimeout = 30 * time.Second

// FindRepoRoot walks up from start to find a .git directory or file.
func FindRepoRoot(start string) (string, error) {
	if start == "" {
		start = "."
	}
	abs, err := filepath.Abs(start)
	if err != nil {
		return "", err
	}
	for {
		gitPath := filepath.Join(abs, ".git")
		if _, err := os.Stat(gitPath); err == nil {
			return abs, nil
		}
		parent := filepath.Dir(abs)
		if parent == abs {
			return "", ErrNotRepo
		}
		abs = parent
	}
}

// EnsureGit checks if git is available.
func EnsureGit() error {
	if _, err := exec.LookPath("git"); err != nil {
		return ErrGitMissing
	}
	return nil
}

// HeadSHA returns the HEAD commit sha.
func HeadSHA(ctx context.Context, repo string) (string, error) {
	out, err := runGit(ctx, repo, []string{"rev-parse", "HEAD"})
	if err != nil {
		return "", err
	}
	return strings.TrimSpace(out), nil
}

// runGit executes git and returns stdout.
func runGit(ctx context.Context, repo string, args []string) (string, error) {
	if ctx == nil {
		ctx = context.Background()
	}
	if _, hasDeadline := ctx.Deadline(); !hasDeadline {
		var cancel context.CancelFunc
		ctx, cancel = context.WithTimeout(ctx, defaultTimeout)
		defer cancel()
	}
	cmd := exec.CommandContext(ctx, "git", append([]string{"-C", repo}, args...)...)
	cmd.Env = os.Environ()
	out, err := cmd.CombinedOutput()
	if err != nil {
		return "", fmt.Errorf("git %s: %w: %s", strings.Join(args, " "), err, strings.TrimSpace(string(out)))
	}
	return string(out), nil
}

// runGitStreaming returns an exec.Cmd ready to start with stdout pipe.
func runGitStreaming(ctx context.Context, repo string, args []string) (*exec.Cmd, func(), error) {
	if ctx == nil {
		ctx = context.Background()
	}
	cancel := func() {}
	if _, hasDeadline := ctx.Deadline(); !hasDeadline {
		ctx, cancel = context.WithTimeout(ctx, defaultTimeout)
	}
	cmd := exec.CommandContext(ctx, "git", append([]string{"-C", repo}, args...)...)
	cmd.Env = os.Environ()
	return cmd, cancel, nil
}

// IsGitRepo verifies if repo is a git work tree.
func IsGitRepo(ctx context.Context, repo string) bool {
	out, err := runGit(ctx, repo, []string{"rev-parse", "--is-inside-work-tree"})
	if err != nil {
		return false
	}
	return strings.TrimSpace(out) == "true"
}
