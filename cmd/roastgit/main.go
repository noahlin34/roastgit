package main

import (
	"context"
	"errors"
	"flag"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
	"time"

	"roastgit/internal/analyze"
	"roastgit/internal/git"
	"roastgit/internal/model"
	"roastgit/internal/render"
	"roastgit/internal/roast"
	"roastgit/internal/util"
)

const (
	exitOK       = 0
	exitUsage    = 2
	exitNotRepo  = 3
	exitGitError = 4
)

func main() {
	cfg, err := parseFlags(os.Args[1:])
	if err != nil {
		if errors.Is(err, flag.ErrHelp) {
			return
		}
		exitWith(exitUsage, err.Error(), true)
	}

	if err := validateConfig(cfg); err != nil {
		exitWith(exitUsage, err.Error(), true)
	}

	if err := git.EnsureGit(); err != nil {
		exitWith(exitGitError, "git executable not found in PATH", false)
	}

	repoPath, err := resolveRepo(cfg.Path)
	if err != nil {
		if errors.Is(err, git.ErrNotRepo) {
			exitWith(exitNotRepo, "not a git repository", false)
		}
		exitWith(exitGitError, err.Error(), false)
	}

	ctx := context.Background()
	commits, head, repoName, err := loadRepo(ctx, repoPath, cfg)
	if err != nil {
		handleGitError(err)
	}

	branches := []string{}
	if bs, err := git.Branches(ctx, repoPath); err == nil {
		branches = bs
	}

	var spinner *util.Spinner
	if !cfg.JSON && len(commits) > 2000 {
		spinner = util.NewSpinner(os.Stderr, "Analyzing commits")
		spinner.Start()
	}
	sizes, sampled, err := loadSizes(ctx, repoPath, commits, cfg)
	if spinner != nil {
		spinner.Stop("Analysis complete")
	}
	if err != nil {
		handleGitError(err)
	}

	metrics, offenders := analyze.Analyze(commits, sizes, branches, analyze.AnalyzeConfig{TZ: cfg.TZ})
	metrics.Size.Sampled = sampled

	score := analyze.Score(metrics)
	seed := head + fmt.Sprintf("-%d-%t", len(commits), cfg.Wholesome)
	roasts := roast.GenerateRoasts(metrics, score, cfg.Intensity, cfg.Wholesome, cfg.Censor, seed)

	report := model.Report{
		Repo: model.RepoInfo{
			Path:        repoPath,
			Name:        repoName,
			Head:        head,
			CommitCount: len(commits),
		},
		Filters: model.Filters{
			Since:      cfg.Since,
			Until:      cfg.Until,
			Author:     cfg.Author,
			MaxCommits: cfg.MaxCommits,
			TZ:         cfg.TZ,
			Deep:       cfg.Deep,
		},
		Score:     score,
		Metrics:   metrics,
		Offenders: offenders,
		Roasts:    roasts,
	}

	if cfg.JSON {
		out, err := render.JSON(report)
		if err != nil {
			exitWith(exitGitError, err.Error(), false)
		}
		fmt.Fprintln(os.Stdout, out)
		return
	}

	output := render.Text(report, render.TextConfig{NoColor: cfg.NoColor, Explain: cfg.Explain})
	fmt.Fprintln(os.Stdout, output)
}

func parseFlags(args []string) (model.Config, error) {
	cfg := model.Config{}
	fs := flag.NewFlagSet("roastgit", flag.ContinueOnError)
	fs.SetOutput(io.Discard)
	fs.StringVar(&cfg.Path, "path", "", "path to repo (default: auto-detect from cwd)")
	fs.StringVar(&cfg.Since, "since", "", "since date YYYY-MM-DD")
	fs.StringVar(&cfg.Until, "until", "", "until date YYYY-MM-DD")
	fs.StringVar(&cfg.Author, "author", "", "author filter")
	fs.BoolVar(&cfg.JSON, "json", false, "output JSON")
	fs.BoolVar(&cfg.NoColor, "no-color", false, "disable ANSI colors")
	fs.IntVar(&cfg.Intensity, "intensity", 3, "roast intensity 0-5")
	fs.BoolVar(&cfg.Wholesome, "wholesome", false, "wholesome mode")
	fs.BoolVar(&cfg.Censor, "censor", false, "censor profanity")
	fs.BoolVar(&cfg.Deep, "deep", false, "analyze all commits for size metrics")
	fs.IntVar(&cfg.MaxCommits, "max-commits", 0, "limit commits analyzed")
	fs.StringVar(&cfg.TZ, "tz", "local", "time zone: local or commit")
	fs.BoolVar(&cfg.Explain, "explain", false, "include scoring explanation")
	fs.Usage = func() {
		fmt.Fprintln(os.Stderr, usageText())
	}
	if err := fs.Parse(args); err != nil {
		if errors.Is(err, flag.ErrHelp) {
			fs.Usage()
			return cfg, flag.ErrHelp
		}
		return cfg, fmt.Errorf("%w", err)
	}
	return cfg, nil
}

func validateConfig(cfg model.Config) error {
	if cfg.Intensity < 0 || cfg.Intensity > 5 {
		return fmt.Errorf("--intensity must be between 0 and 5")
	}
	if cfg.TZ != "local" && cfg.TZ != "commit" {
		return fmt.Errorf("--tz must be 'local' or 'commit'")
	}
	if cfg.Since != "" {
		if _, err := time.Parse("2006-01-02", cfg.Since); err != nil {
			return fmt.Errorf("--since must be YYYY-MM-DD")
		}
	}
	if cfg.Until != "" {
		if _, err := time.Parse("2006-01-02", cfg.Until); err != nil {
			return fmt.Errorf("--until must be YYYY-MM-DD")
		}
	}
	return nil
}

func resolveRepo(path string) (string, error) {
	if path == "" {
		cwd, err := os.Getwd()
		if err != nil {
			return "", err
		}
		return git.FindRepoRoot(cwd)
	}
	return git.FindRepoRoot(path)
}

func loadRepo(ctx context.Context, repoPath string, cfg model.Config) ([]model.Commit, string, string, error) {
	if !git.IsGitRepo(ctx, repoPath) {
		return nil, "", "", git.ErrNotRepo
	}
	head, err := git.HeadSHA(ctx, repoPath)
	if err != nil {
		return nil, "", "", err
	}
	repoName := filepath.Base(repoPath)
	commits, err := git.LogCommits(ctx, repoPath, git.LogOptions{
		Since:      cfg.Since,
		Until:      cfg.Until,
		Author:     cfg.Author,
		MaxCommits: cfg.MaxCommits,
	})
	if err != nil {
		return nil, "", "", err
	}
	return commits, head, repoName, nil
}

func loadSizes(ctx context.Context, repoPath string, commits []model.Commit, cfg model.Config) (map[string]model.CommitSize, bool, error) {
	if len(commits) == 0 {
		return map[string]model.CommitSize{}, false, nil
	}
	if cfg.Deep {
		sizes, err := git.NumstatForLog(ctx, repoPath, git.LogOptions{
			Since:      cfg.Since,
			Until:      cfg.Until,
			Author:     cfg.Author,
			MaxCommits: cfg.MaxCommits,
		})
		return sizes, false, err
	}
	count := len(commits)
	sampleSize := count
	if count > 500 {
		sampleSize = 500
	}
	indices := util.SampleIndices(count, sampleSize)
	shas := make([]string, 0, len(indices))
	for _, idx := range indices {
		shas = append(shas, commits[idx].SHA)
	}
	sizes, err := git.NumstatForCommits(ctx, repoPath, shas)
	return sizes, sampleSize < count, err
}

func handleGitError(err error) {
	msg := err.Error()
	if strings.Contains(strings.ToLower(msg), "not a git repository") {
		exitWith(exitNotRepo, "not a git repository", false)
	}
	exitWith(exitGitError, msg, false)
}

func exitWith(code int, message string, showUsage bool) {
	if message != "" {
		fmt.Fprintln(os.Stderr, message)
		if showUsage {
			fmt.Fprintln(os.Stderr, usageText())
		}
	}
	os.Exit(code)
}

func usageText() string {
	return `Usage: roastgit [flags]

Flags:
  --path string         path to repo (default: auto-detect from cwd)
  --since YYYY-MM-DD
  --until YYYY-MM-DD
  --author string
  --json               output JSON only
  --no-color           disable ANSI colors
  --intensity int      0-5 (default 3)
  --wholesome          wholesome mode
  --censor             censor profanity
  --deep               analyze all commits for numstat-based metrics
  --max-commits int    limit commits analyzed from newest backwards (default 0 = no limit)
  --tz string          "local" (default) or "commit"
  --explain            include scoring explanation in text output
  -h, --help
`
}
