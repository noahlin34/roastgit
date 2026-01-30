package git

import (
	"context"
	"strings"
)

// Branches returns local branch names.
func Branches(ctx context.Context, repo string) ([]string, error) {
	out, err := runGit(ctx, repo, []string{"for-each-ref", "refs/heads", "--format=%(refname:short)"})
	if err != nil {
		return nil, err
	}
	lines := strings.Split(strings.TrimSpace(out), "\n")
	branches := []string{}
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}
		branches = append(branches, line)
	}
	return branches, nil
}
