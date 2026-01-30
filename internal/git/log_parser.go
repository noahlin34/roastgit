package git

import (
	"bufio"
	"context"
	"io"
	"strconv"
	"strings"
	"time"

	"roastgit/internal/model"
)

type LogOptions struct {
	Since      string
	Until      string
	Author     string
	MaxCommits int
}

const (
	recordSep byte = 0x1e
	unitSep   byte = 0x1f
)

// LogCommits streams git log and returns parsed commits.
func LogCommits(ctx context.Context, repo string, opts LogOptions) ([]model.Commit, error) {
	args := []string{"log", "--date=iso-strict", "--pretty=format:%H%x1f%an%x1f%ae%x1f%ad%x1f%s%x1f%P%x1e"}
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
	commits, parseErr := ParseLog(stdout)
	// drain stderr for git errors
	_, _ = io.ReadAll(stderr)
	if err := cmd.Wait(); err != nil {
		return nil, err
	}
	if parseErr != nil {
		return nil, parseErr
	}
	return commits, nil
}

// ParseLog parses commits from git log output.
func ParseLog(r io.Reader) ([]model.Commit, error) {
	scanner := bufio.NewScanner(r)
	scanner.Buffer(make([]byte, 0, 64*1024), 1024*1024)
	scanner.Split(splitOnRecordSep)
	commits := make([]model.Commit, 0, 256)
	for scanner.Scan() {
		rec := strings.TrimSpace(scanner.Text())
		if rec == "" {
			continue
		}
		fields := splitFields(rec, unitSep)
		if len(fields) < 6 {
			continue
		}
		date, err := time.Parse(time.RFC3339, fields[3])
		if err != nil {
			return nil, err
		}
		parents := []string{}
		if strings.TrimSpace(fields[5]) != "" {
			parents = strings.Fields(fields[5])
		}
		commit := model.Commit{
			SHA:         fields[0],
			AuthorName:  fields[1],
			AuthorEmail: fields[2],
			Date:        date,
			Subject:     fields[4],
			Parents:     parents,
		}
		commits = append(commits, commit)
	}
	if err := scanner.Err(); err != nil {
		return nil, err
	}
	return commits, nil
}

func splitOnRecordSep(data []byte, atEOF bool) (advance int, token []byte, err error) {
	for i, b := range data {
		if b == recordSep {
			return i + 1, data[:i], nil
		}
	}
	if atEOF && len(data) > 0 {
		return len(data), data, nil
	}
	return 0, nil, nil
}

func splitFields(s string, sep byte) []string {
	parts := []string{}
	start := 0
	for i := 0; i < len(s); i++ {
		if s[i] == sep {
			parts = append(parts, s[start:i])
			start = i + 1
		}
	}
	parts = append(parts, s[start:])
	return parts
}

func intToString(v int) string {
	return strconv.FormatInt(int64(v), 10)
}
