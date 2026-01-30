package git

import (
	"bufio"
	"context"
	"io"
	"strconv"
	"strings"

	"roastgit/internal/model"
)

// NumstatForCommits returns numstat data for a list of commits.
func NumstatForCommits(ctx context.Context, repo string, shas []string) (map[string]model.CommitSize, error) {
	if len(shas) == 0 {
		return map[string]model.CommitSize{}, nil
	}
	result := make(map[string]model.CommitSize, len(shas))
	batchSize := 200
	for i := 0; i < len(shas); i += batchSize {
		end := i + batchSize
		if end > len(shas) {
			end = len(shas)
		}
		args := []string{"show", "--numstat", "--format=%H%x1f"}
		args = append(args, shas[i:end]...)
		cmd, cancel, err := runGitStreaming(ctx, repo, args)
		if err != nil {
			return nil, err
		}
		defer cancel()
		stdout, err := cmd.StdoutPipe()
		if err != nil {
			return nil, err
		}
		stderr, err := cmd.StderrPipe()
		if err != nil {
			return nil, err
		}
		if err := cmd.Start(); err != nil {
			return nil, err
		}
		parsed, parseErr := ParseNumstat(stdout)
		_, _ = io.ReadAll(stderr)
		if err := cmd.Wait(); err != nil {
			return nil, err
		}
		if parseErr != nil {
			return nil, parseErr
		}
		for k, v := range parsed {
			result[k] = v
		}
	}
	return result, nil
}

// NumstatForLog returns numstat data for git log with filters.
func NumstatForLog(ctx context.Context, repo string, opts LogOptions) (map[string]model.CommitSize, error) {
	args := []string{"log", "--numstat", "--pretty=format:%H%x1f"}
	if opts.Since != "" {
		args = append(args, "--since", opts.Since)
	}
	if opts.Until != "" {
		args = append(args, "--until", opts.Until)
	}
	if opts.Author != "" {
		args = append(args, "--author", opts.Author)
	}
	if opts.MaxCommits > 0 {
		args = append(args, "-n", intToString(opts.MaxCommits))
	}
	cmd, cancel, err := runGitStreaming(ctx, repo, args)
	if err != nil {
		return nil, err
	}
	defer cancel()
	stdout, err := cmd.StdoutPipe()
	if err != nil {
		return nil, err
	}
	stderr, err := cmd.StderrPipe()
	if err != nil {
		return nil, err
	}
	if err := cmd.Start(); err != nil {
		return nil, err
	}
	parsed, parseErr := ParseNumstat(stdout)
	_, _ = io.ReadAll(stderr)
	if err := cmd.Wait(); err != nil {
		return nil, err
	}
	if parseErr != nil {
		return nil, parseErr
	}
	return parsed, nil
}

// ParseNumstat parses numstat output into commit sizes.
func ParseNumstat(r io.Reader) (map[string]model.CommitSize, error) {
	scanner := bufio.NewScanner(r)
	scanner.Buffer(make([]byte, 0, 64*1024), 1024*1024)
	result := map[string]model.CommitSize{}
	current := ""
	for scanner.Scan() {
		line := scanner.Text()
		if strings.ContainsRune(line, rune(unitSep)) {
			parts := strings.Split(line, string([]byte{unitSep}))
			current = strings.TrimSpace(parts[0])
			if current != "" {
				if _, ok := result[current]; !ok {
					result[current] = model.CommitSize{}
				}
			}
			continue
		}
		if strings.TrimSpace(line) == "" {
			continue
		}
		fields := strings.Split(line, "\t")
		if len(fields) < 3 || current == "" {
			continue
		}
		size := result[current]
		size.Files++
		if fields[0] == "-" || fields[1] == "-" {
			size.BinaryFiles++
			result[current] = size
			continue
		}
		added, errA := strconv.Atoi(fields[0])
		deleted, errD := strconv.Atoi(fields[1])
		if errA == nil {
			size.Added += added
		}
		if errD == nil {
			size.Deleted += deleted
		}
		result[current] = size
	}
	if err := scanner.Err(); err != nil {
		return nil, err
	}
	return result, nil
}
