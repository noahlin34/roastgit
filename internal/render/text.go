package render

import (
	"fmt"
	"math"
	"strings"

	"roastgit/internal/model"
	"roastgit/internal/util"
)

type TextConfig struct {
	NoColor bool
	Explain bool
}

// Text renders a human-friendly report.
func Text(report model.Report, cfg TextConfig) string {
	b := &strings.Builder{}
	color := func(s, code string) string {
		if cfg.NoColor {
			return s
		}
		return code + s + "\x1b[0m"
	}

	headerColor := "\x1b[36m"
	scoreColor := scoreColor(report.Score.Overall)

	fmt.Fprintf(b, "%s\n", color("Roastgit Report", headerColor))
	fmt.Fprintf(b, "Repo: %s (%s)\n", report.Repo.Name, report.Repo.Path)
	fmt.Fprintf(b, "HEAD: %s\n", util.ShortSHA(report.Repo.Head))
	fmt.Fprintf(b, "Commits analyzed: %d\n", report.Repo.CommitCount)
	if report.Filters.Since != "" || report.Filters.Until != "" {
		fmt.Fprintf(b, "Range: %s -> %s\n", emptyAsAll(report.Filters.Since), emptyAsAll(report.Filters.Until))
	}
	if report.Filters.Author != "" {
		fmt.Fprintf(b, "Author filter: %s\n", report.Filters.Author)
	}
	fmt.Fprintf(b, "\n%s %s\n", color(fmt.Sprintf("Overall Score: %d/100", report.Score.Overall), scoreColor), report.Roasts.Headline)
	if cfg.Explain && len(report.Score.Explain) > 0 {
		fmt.Fprintf(b, "Score breakdown: message %d/30, hygiene %d/30, cadence %d/20, size %d/20\n",
			report.Score.Breakdown.MessageQuality,
			report.Score.Breakdown.Hygiene,
			report.Score.Breakdown.Cadence,
			report.Score.Breakdown.SizeDiscipline,
		)
		fmt.Fprintf(b, "Explain: %s\n", report.Score.Explain["overall"])
	}

	writeSection(b, color("Commit Message Crimes", headerColor), messageBullets(report.Metrics), report.Roasts.Sections["commit_messages"])
	writeSection(b, color("Time & Cadence", headerColor), timeBullets(report.Metrics), report.Roasts.Sections["time_cadence"])
	writeSection(b, color("Repo Hygiene", headerColor), hygieneBullets(report.Metrics), report.Roasts.Sections["repo_hygiene"])
	writeSection(b, color("Chunkiness", headerColor), sizeBullets(report.Metrics), report.Roasts.Sections["chunkiness"])

	if len(report.Offenders) > 0 {
		fmt.Fprintf(b, "\n%s\n", color("Top Offenders", headerColor))
		for _, off := range report.Offenders {
			subject := truncate(off.Subject, 60)
			date := off.Date
			reason := strings.Join(off.Reasons, ", ")
			fmt.Fprintf(b, "- %s %s %s -- %s\n", util.ShortSHA(off.SHA), date[:10], subject, reason)
		}
	}

	if len(report.Roasts.Tips) > 0 {
		fmt.Fprintf(b, "\n%s\n", color("Tips", headerColor))
		for _, tip := range report.Roasts.Tips {
			fmt.Fprintf(b, "- %s\n", tip)
		}
	}

	fmt.Fprintf(b, "\nHints: try --deep for full size analysis, --wholesome for kinder output, or --json for machine use.\n")
	return b.String()
}

func writeSection(b *strings.Builder, title string, bullets []string, summary string) {
	fmt.Fprintf(b, "\n%s\n", title)
	for _, bullet := range bullets {
		fmt.Fprintf(b, "- %s\n", bullet)
	}
	if summary != "" {
		fmt.Fprintf(b, "%s\n", summary)
	}
}

func messageBullets(metrics model.Metrics) []string {
	bullets := []string{}
	if metrics.Message.Total == 0 {
		return []string{"No commits found."}
	}
	bullets = append(bullets, fmt.Sprintf("Generic messages: %d (%.0f%%)", metrics.Message.Generic, percent(metrics.Message.Generic, metrics.Message.Total)))
	bullets = append(bullets, fmt.Sprintf("Emoji-only: %d, too short: %d, too long: %d", metrics.Message.EmojiOnly, metrics.Message.TooShort, metrics.Message.TooLong))
	if metrics.Message.Lying > 0 {
		bullets = append(bullets, fmt.Sprintf("Lying messages: %d", metrics.Message.Lying))
	}
	if metrics.Message.Panic > 0 {
		bullets = append(bullets, fmt.Sprintf("Panic commits: %d", metrics.Message.Panic))
	}
	if len(metrics.Message.TopGenericWords) > 0 {
		bullets = append(bullets, fmt.Sprintf("Top generic words: %s", strings.Join(metrics.Message.TopGenericWords, ", ")))
	}
	return trimBullets(bullets, 6)
}

func timeBullets(metrics model.Metrics) []string {
	bullets := []string{}
	bullets = append(bullets, fmt.Sprintf("Commits/day avg: %.2f", metrics.Time.CommitsPerDayAvg))
	bullets = append(bullets, fmt.Sprintf("Commits/week avg: %.2f", metrics.Time.CommitsPerWeekAvg))
	bullets = append(bullets, fmt.Sprintf("Midnight commits: %.0f%%", metrics.Time.MidnightRatio*100))
	bullets = append(bullets, fmt.Sprintf("Deadline window commits: %.0f%%", metrics.Time.DeadlineRatio*100))
	bullets = append(bullets, fmt.Sprintf("Longest streak: %d days", metrics.Time.LongestStreakDays))
	return trimBullets(bullets, 6)
}

func hygieneBullets(metrics model.Metrics) []string {
	bullets := []string{}
	bullets = append(bullets, fmt.Sprintf("Merge ratio: %.0f%%", metrics.Hygiene.MergeRatio*100))
	bullets = append(bullets, fmt.Sprintf("Linear ratio: %.0f%%", metrics.Hygiene.LinearRatio*100))
	bullets = append(bullets, fmt.Sprintf("Branches: %d (bad: %d)", metrics.Hygiene.BranchCount, metrics.Hygiene.BadBranchCount))
	if len(metrics.Hygiene.BadBranches) > 0 {
		bullets = append(bullets, fmt.Sprintf("Bad branches: %s", strings.Join(trimStrings(metrics.Hygiene.BadBranches, 3), ", ")))
	}
	return trimBullets(bullets, 6)
}

func sizeBullets(metrics model.Metrics) []string {
	bullets := []string{}
	if metrics.Size.SampleSize == 0 {
		return []string{"No size data collected."}
	}
	sampleNote := ""
	if metrics.Size.Sampled {
		sampleNote = " (sampled)"
	}
	bullets = append(bullets, fmt.Sprintf("Large commits: %d/%d%s", metrics.Size.LargeCommitCount, metrics.Size.SampleSize, sampleNote))
	bullets = append(bullets, fmt.Sprintf("Binary commits: %d/%d%s", metrics.Size.BinaryCommitCount, metrics.Size.SampleSize, sampleNote))
	bullets = append(bullets, fmt.Sprintf("Average lines changed: %.0f", metrics.Size.AverageLines))
	bullets = append(bullets, fmt.Sprintf("Max lines changed: %d", metrics.Size.MaxLines))
	return trimBullets(bullets, 6)
}

func trimBullets(bullets []string, max int) []string {
	if len(bullets) <= max {
		return bullets
	}
	return bullets[:max]
}

func percent(a, b int) float64 {
	if b == 0 {
		return 0
	}
	return math.Round(float64(a)*1000/float64(b)) / 10
}

func scoreColor(score int) string {
	switch {
	case score >= 85:
		return "\x1b[32m"
	case score >= 70:
		return "\x1b[33m"
	case score >= 50:
		return "\x1b[35m"
	default:
		return "\x1b[31m"
	}
}

func truncate(s string, max int) string {
	runes := []rune(s)
	if len(runes) <= max {
		return s
	}
	return string(runes[:max-3]) + "..."
}

func emptyAsAll(s string) string {
	if s == "" {
		return "all time"
	}
	return s
}

func trimStrings(in []string, max int) []string {
	if len(in) <= max {
		return in
	}
	return in[:max]
}
