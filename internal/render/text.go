package render

import (
	"fmt"
	"math"
	"os"
	"strconv"
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

	palette := pickPalette()
	headerColor := palette.Header
	scoreColor := scoreColor(report.Score.Overall, palette)
	label := func(s string) string { return color(s, palette.Label) }
	muted := func(s string) string { return color(s, palette.Muted) }
	accent := func(s string) string { return color(s, palette.Accent) }
	body := func(s string) string { return color(s, palette.Body) }
	bulletPrefix := color("- ", palette.Bullet)
	headline := report.Roasts.Headline
	if headline != "" {
		headline = accent(headline)
	}

	fmt.Fprintf(b, "%s\n", color("Roastgit Report", headerColor))
	fmt.Fprintf(b, "%s %s (%s)\n", label("Repo:"), report.Repo.Name, muted(report.Repo.Path))
	fmt.Fprintf(b, "%s %s\n", label("HEAD:"), accent(util.ShortSHA(report.Repo.Head)))
	fmt.Fprintf(b, "%s %s\n", label("Commits analyzed:"), body(fmt.Sprintf("%d", report.Repo.CommitCount)))
	if report.Filters.Since != "" || report.Filters.Until != "" {
		fmt.Fprintf(b, "%s %s -> %s\n", label("Range:"), body(emptyAsAll(report.Filters.Since)), body(emptyAsAll(report.Filters.Until)))
	}
	if report.Filters.Author != "" {
		fmt.Fprintf(b, "%s %s\n", label("Author filter:"), body(report.Filters.Author))
	}
	fmt.Fprintf(b, "\n%s %s\n", color(fmt.Sprintf("Overall Score: %d/100", report.Score.Overall), scoreColor), headline)
	if cfg.Explain && len(report.Score.Explain) > 0 {
		fmt.Fprintf(b, "%s %s\n", muted("Score breakdown:"), body(fmt.Sprintf("message %d/30, hygiene %d/30, cadence %d/20, size %d/20",
			report.Score.Breakdown.MessageQuality,
			report.Score.Breakdown.Hygiene,
			report.Score.Breakdown.Cadence,
			report.Score.Breakdown.SizeDiscipline,
		)))
		fmt.Fprintf(b, "%s %s\n", muted("Explain:"), body(report.Score.Explain["overall"]))
	}

	writeSection(b, color("Commit Message Crimes", headerColor), messageBullets(report.Metrics), report.Roasts.Sections["commit_messages"], color, bulletPrefix, palette.Body, palette.Accent)
	writeSection(b, color("Time & Cadence", headerColor), timeBullets(report.Metrics), report.Roasts.Sections["time_cadence"], color, bulletPrefix, palette.Body, palette.Accent)
	writeSection(b, color("Repo Hygiene", headerColor), hygieneBullets(report.Metrics), report.Roasts.Sections["repo_hygiene"], color, bulletPrefix, palette.Body, palette.Accent)
	writeSection(b, color("Chunkiness", headerColor), sizeBullets(report.Metrics), report.Roasts.Sections["chunkiness"], color, bulletPrefix, palette.Body, palette.Accent)

	if len(report.Offenders) > 0 {
		fmt.Fprintf(b, "\n%s\n", color("Top Offenders", headerColor))
		for _, off := range report.Offenders {
			subject := truncate(off.Subject, 60)
			date := off.Date[:10]
			reason := strings.Join(off.Reasons, ", ")
			fmt.Fprintf(b, "%s%s %s %s -- %s\n",
				bulletPrefix,
				accent(util.ShortSHA(off.SHA)),
				muted(date),
				body(subject),
				label(reason),
			)
		}
	}

	if len(report.Roasts.Tips) > 0 {
		fmt.Fprintf(b, "\n%s\n", color("Tips", headerColor))
		for _, tip := range report.Roasts.Tips {
			fmt.Fprintf(b, "%s%s\n", bulletPrefix, accent(tip))
		}
	}

	hint := fmt.Sprintf("try %s for full size analysis, %s for kinder output, or %s for machine use.",
		accent("--deep"),
		accent("--wholesome"),
		accent("--json"),
	)
	fmt.Fprintf(b, "\n%s %s\n", label("Hints:"), hint)
	return b.String()
}

func writeSection(b *strings.Builder, title string, bullets []string, summary string, colorize func(string, string) string, bulletPrefix string, bulletColor string, summaryColor string) {
	fmt.Fprintf(b, "\n%s\n", title)
	for _, bullet := range bullets {
		fmt.Fprintf(b, "%s%s\n", bulletPrefix, colorize(bullet, bulletColor))
	}
	if summary != "" {
		fmt.Fprintf(b, "%s\n", colorize(summary, summaryColor))
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

type colorPalette struct {
	Header    string
	Label     string
	Bullet    string
	Muted     string
	Accent    string
	Body      string
	ScoreGood string
	ScoreOk   string
	ScoreMid  string
	ScoreBad  string
}

func scoreColor(score int, palette colorPalette) string {
	switch {
	case score >= 85:
		return palette.ScoreGood
	case score >= 70:
		return palette.ScoreOk
	case score >= 50:
		return palette.ScoreMid
	default:
		return palette.ScoreBad
	}
}

func pickPalette() colorPalette {
	switch terminalBackground() {
	case "light":
		return colorPalette{
			Header:    ansiRGB(29, 78, 216),
			Label:     ansiRGB(14, 116, 144),
			Bullet:    ansiRGB(21, 128, 61),
			Muted:     ansiRGB(90, 98, 110),
			Accent:    ansiRGB(180, 83, 9),
			Body:      ansiRGB(30, 41, 59),
			ScoreGood: ansiRGB(21, 128, 61),
			ScoreOk:   ansiRGB(180, 83, 9),
			ScoreMid:  ansiRGB(109, 40, 217),
			ScoreBad:  ansiRGB(185, 28, 28),
		}
	default:
		return colorPalette{
			Header:    ansiRGB(138, 180, 255),
			Label:     ansiRGB(125, 211, 252),
			Bullet:    ansiRGB(163, 214, 102),
			Muted:     ansiRGB(148, 163, 184),
			Accent:    ansiRGB(255, 205, 120),
			Body:      ansiRGB(226, 232, 240),
			ScoreGood: ansiRGB(102, 204, 134),
			ScoreOk:   ansiRGB(245, 201, 102),
			ScoreMid:  ansiRGB(187, 154, 247),
			ScoreBad:  ansiRGB(247, 118, 142),
		}
	}
}

func terminalBackground() string {
	// Heuristic based on COLORFGBG; default to dark when unavailable.
	value := os.Getenv("COLORFGBG")
	if value == "" {
		return "dark"
	}
	parts := strings.Split(value, ";")
	bgStr := parts[len(parts)-1]
	bg, err := strconv.Atoi(bgStr)
	if err != nil || bg < 0 {
		return "dark"
	}
	r, g, b, ok := ansiColorToRGB(bg)
	if !ok {
		return "dark"
	}
	luminance := (0.2126*float64(r) + 0.7152*float64(g) + 0.0722*float64(b)) / 255.0
	if luminance >= 0.6 {
		return "light"
	}
	return "dark"
}

func ansiColorToRGB(code int) (int, int, int, bool) {
	if code >= 0 && code <= 15 {
		base := [16][3]int{
			{0, 0, 0},
			{205, 0, 0},
			{0, 205, 0},
			{205, 205, 0},
			{0, 0, 238},
			{205, 0, 205},
			{0, 205, 205},
			{229, 229, 229},
			{127, 127, 127},
			{255, 0, 0},
			{0, 255, 0},
			{255, 255, 0},
			{92, 92, 255},
			{255, 0, 255},
			{0, 255, 255},
			{255, 255, 255},
		}
		rgb := base[code]
		return rgb[0], rgb[1], rgb[2], true
	}
	if code >= 16 && code <= 231 {
		c := code - 16
		r := c / 36
		g := (c % 36) / 6
		b := c % 6
		scale := []int{0, 95, 135, 175, 215, 255}
		return scale[r], scale[g], scale[b], true
	}
	if code >= 232 && code <= 255 {
		gray := 8 + (code-232)*10
		return gray, gray, gray, true
	}
	return 0, 0, 0, false
}

func ansiRGB(r, g, b int) string {
	return fmt.Sprintf("\x1b[38;2;%d;%d;%dm", r, g, b)
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
