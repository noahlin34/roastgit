package analyze

import (
	"fmt"
	"sort"
	"strings"
	"time"

	"roastgit/internal/model"
)

type AnalyzeConfig struct {
	TZ string
}

type commitFlags struct {
	msgInfo    MessageInfo
	lowQuality bool
	lying      bool
	panic      bool
	large      bool
	binary     bool
	midnight   bool
	deadline   bool
}

// Analyze computes metrics and offenders for a set of commits.
func Analyze(commits []model.Commit, sizes map[string]model.CommitSize, branches []string, cfg AnalyzeConfig) (model.Metrics, []model.Offender) {
	metrics := model.Metrics{}
	if len(commits) == 0 {
		return metrics, nil
	}
	flags := make([]commitFlags, len(commits))
	genericCounts := map[string]int{}
	msgLenTotal := 0
	msgQualityTotal := 0
	lowQualityCount := 0
	lyingCount := 0

	// Attach size info and message metrics.
	sizeMetrics := model.SizeMetrics{}
	var sizeLineTotal int
	for i := range commits {
		c := &commits[i]
		if size, ok := sizes[c.SHA]; ok {
			c.Size = &size
		}
		info := AnalyzeMessage(c.Subject)
		flags[i].msgInfo = info
		if info.Generic {
			genericCounts[info.GenericKey]++
		}
		msgLenTotal += len([]rune(c.Subject))
		msgQualityTotal += info.Score
		low := info.Generic || info.EmojiOnly || info.TooLong || info.TooShort
		flags[i].lowQuality = low
		if low {
			lowQualityCount++
		}
		if c.Size != nil {
			lines := c.Size.Added + c.Size.Deleted
			sizeLineTotal += lines
			if lines > sizeMetrics.MaxLines {
				sizeMetrics.MaxLines = lines
			}
			if lines >= 800 || c.Size.Files >= 20 {
				sizeMetrics.LargeCommitCount++
				flags[i].large = true
			}
			if c.Size.BinaryFiles > 0 {
				sizeMetrics.BinaryCommitCount++
				flags[i].binary = true
			}
			sizeMetrics.SampleSize++
			if IsLyingMessage(c.Subject, lines, c.Size.Files) {
				flags[i].lying = true
				lyingCount++
			}
		}
	}
	metrics.Message.Total = len(commits)
	metrics.Message.Generic = countGeneric(flags)
	metrics.Message.EmojiOnly = countEmoji(flags)
	metrics.Message.TooLong = countTooLong(flags)
	metrics.Message.TooShort = countTooShort(flags)
	metrics.Message.Lying = lyingCount
	metrics.Message.LowQuality = lowQualityCount
	metrics.Message.AverageLength = float64(msgLenTotal) / float64(len(commits))
	metrics.Message.AverageQuality = float64(msgQualityTotal) / float64(len(commits))
	metrics.Message.TopGenericWords = topGenericWords(genericCounts, 3)

	metrics.Size = sizeMetrics
	if sizeMetrics.SampleSize > 0 {
		metrics.Size.AverageLines = float64(sizeLineTotal) / float64(sizeMetrics.SampleSize)
	}

	// Time metrics and panic detection.
	applyTZ := func(t time.Time) time.Time {
		if cfg.TZ == "commit" {
			return t
		}
		return t.In(time.Local)
	}
	dayCounts := map[string]int{}
	weekCounts := map[string]int{}
	midnightCount := 0
	deadlineCount := 0
	timesAsc, idxAsc := orderTimesAsc(commits, applyTZ)
	for i, idx := range idxAsc {
		local := timesAsc[i]
		dayKey := local.Format("2006-01-02")
		dayCounts[dayKey]++
		year, week := local.ISOWeek()
		weekKey := fmt.Sprintf("%04d-W%02d", year, week)
		weekCounts[weekKey]++
		hour := local.Hour()
		if hour >= 0 && hour < 5 {
			midnightCount++
			flags[idx].midnight = true
		}
		if isDeadlineTime(local) {
			deadlineCount++
			flags[idx].deadline = true
		}
	}
	metrics.Time.UniqueDays = len(dayCounts)
	metrics.Time.UniqueWeeks = len(weekCounts)
	metrics.Time.CommitsPerDayAvg = float64(len(commits)) / float64(max(1, metrics.Time.UniqueDays))
	metrics.Time.CommitsPerWeekAvg = float64(len(commits)) / float64(max(1, metrics.Time.UniqueWeeks))
	metrics.Time.MidnightRatio = float64(midnightCount) / float64(len(commits))
	metrics.Time.DeadlineRatio = float64(deadlineCount) / float64(len(commits))
	metrics.Time.LongestStreakDays = longestStreak(dayCounts)

	panicFlags := detectPanic(timesAsc, flags, idxAsc)
	panicCount := 0
	for i, isPanic := range panicFlags {
		if isPanic {
			flags[i].panic = true
			panicCount++
		}
	}
	metrics.Message.Panic = panicCount

	// Hygiene metrics.
	mergeCount := 0
	for _, c := range commits {
		if len(c.Parents) > 1 {
			mergeCount++
		}
	}
	metrics.Hygiene.MergeRatio = float64(mergeCount) / float64(len(commits))
	metrics.Hygiene.LinearRatio = 1 - metrics.Hygiene.MergeRatio
	metrics.Hygiene.BranchCount = len(branches)
	badBranches := badBranchNames(branches)
	metrics.Hygiene.BadBranchCount = len(badBranches)
	metrics.Hygiene.BadBranches = badBranches

	// Offenders
	offenders := buildOffenders(commits, flags)
	return metrics, offenders
}

func countGeneric(flags []commitFlags) int {
	count := 0
	for _, f := range flags {
		if f.msgInfo.Generic {
			count++
		}
	}
	return count
}

func countEmoji(flags []commitFlags) int {
	count := 0
	for _, f := range flags {
		if f.msgInfo.EmojiOnly {
			count++
		}
	}
	return count
}

func countTooLong(flags []commitFlags) int {
	count := 0
	for _, f := range flags {
		if f.msgInfo.TooLong {
			count++
		}
	}
	return count
}

func countTooShort(flags []commitFlags) int {
	count := 0
	for _, f := range flags {
		if f.msgInfo.TooShort {
			count++
		}
	}
	return count
}

func topGenericWords(counts map[string]int, limit int) []string {
	type pair struct {
		Word  string
		Count int
	}
	pairs := []pair{}
	for k, v := range counts {
		pairs = append(pairs, pair{Word: k, Count: v})
	}
	sort.Slice(pairs, func(i, j int) bool {
		if pairs[i].Count == pairs[j].Count {
			return pairs[i].Word < pairs[j].Word
		}
		return pairs[i].Count > pairs[j].Count
	})
	out := []string{}
	for i := 0; i < len(pairs) && i < limit; i++ {
		out = append(out, pairs[i].Word)
	}
	return out
}

func orderTimesAsc(commits []model.Commit, apply func(time.Time) time.Time) ([]time.Time, []int) {
	n := len(commits)
	times := make([]time.Time, n)
	idx := make([]int, n)
	for i := 0; i < n; i++ {
		orig := n - 1 - i
		idx[i] = orig
		times[i] = apply(commits[orig].Date)
	}
	return times, idx
}

func isDeadlineTime(t time.Time) bool {
	weekday := t.Weekday()
	hour := t.Hour()
	if weekday == time.Monday {
		return hour >= 8 && hour <= 11
	}
	if weekday == time.Friday {
		return hour >= 15 && hour <= 19
	}
	return false
}

func longestStreak(dayCounts map[string]int) int {
	if len(dayCounts) == 0 {
		return 0
	}
	dates := make([]time.Time, 0, len(dayCounts))
	for day := range dayCounts {
		t, err := time.Parse("2006-01-02", day)
		if err != nil {
			continue
		}
		dates = append(dates, t)
	}
	sort.Slice(dates, func(i, j int) bool { return dates[i].Before(dates[j]) })
	longest := 1
	current := 1
	for i := 1; i < len(dates); i++ {
		if dates[i].Sub(dates[i-1]) == 24*time.Hour {
			current++
			if current > longest {
				longest = current
			}
		} else {
			current = 1
		}
	}
	return longest
}

func detectPanic(timesAsc []time.Time, flags []commitFlags, idxAsc []int) []bool {
	n := len(timesAsc)
	panicFlags := make([]bool, n)
	if n == 0 {
		return panicFlags
	}
	lowPrefix := make([]int, n+1)
	for i := 0; i < n; i++ {
		origIdx := idxAsc[i]
		lowPrefix[i+1] = lowPrefix[i]
		if flags[origIdx].lowQuality {
			lowPrefix[i+1]++
		}
	}
	intervals := [][2]int{}
	start := 0
	for end := 0; end < n; end++ {
		for timesAsc[end].Sub(timesAsc[start]) > time.Hour {
			start++
		}
		windowSize := end - start + 1
		lowCount := lowPrefix[end+1] - lowPrefix[start]
		if windowSize >= 4 && lowCount >= 3 {
			intervals = append(intervals, [2]int{start, end})
		}
	}
	if len(intervals) == 0 {
		return panicFlags
	}
	// Merge intervals
	sort.Slice(intervals, func(i, j int) bool { return intervals[i][0] < intervals[j][0] })
	merged := [][2]int{intervals[0]}
	for _, in := range intervals[1:] {
		last := &merged[len(merged)-1]
		if in[0] <= last[1]+1 {
			if in[1] > last[1] {
				last[1] = in[1]
			}
		} else {
			merged = append(merged, in)
		}
	}
	for _, in := range merged {
		for i := in[0]; i <= in[1]; i++ {
			panicFlags[idxAsc[i]] = true
		}
	}
	return panicFlags
}

func badBranchNames(branches []string) []string {
	bad := []string{}
	badSet := map[string]struct{}{"test": {}, "new": {}, "asdf": {}, "temp": {}, "final": {}, "final2": {}, "wip": {}, "tmp": {}}
	for _, b := range branches {
		name := strings.ToLower(b)
		if _, ok := badSet[name]; ok {
			bad = append(bad, b)
			continue
		}
		if len(name) <= 3 {
			bad = append(bad, b)
		}
	}
	return bad
}

func buildOffenders(commits []model.Commit, flags []commitFlags) []model.Offender {
	type candidate struct {
		idx     int
		score   int
		reasons []string
	}
	cands := []candidate{}
	for i := range commits {
		reasons := []string{}
		score := 0
		f := flags[i]
		if f.msgInfo.Generic {
			reasons = append(reasons, "generic message")
			score += 8
		}
		if f.msgInfo.EmojiOnly {
			reasons = append(reasons, "emoji-only message")
			score += 7
		}
		if f.msgInfo.TooLong {
			reasons = append(reasons, "too long")
			score += 3
		}
		if f.msgInfo.TooShort {
			reasons = append(reasons, "too short")
			score += 3
		}
		if f.lying {
			reasons = append(reasons, "lying message")
			score += 9
		}
		if f.panic {
			reasons = append(reasons, "panic streak")
			score += 6
		}
		if f.large {
			reasons = append(reasons, "huge commit")
			score += 7
		}
		if f.binary {
			reasons = append(reasons, "binary blobs")
			score += 5
		}
		if f.midnight {
			reasons = append(reasons, "midnight gremlin")
			score += 2
		}
		if f.deadline {
			reasons = append(reasons, "deadline scramble")
			score += 2
		}
		if len(reasons) == 0 {
			continue
		}
		cands = append(cands, candidate{idx: i, score: score, reasons: reasons})
	}
	sort.Slice(cands, func(i, j int) bool {
		if cands[i].score == cands[j].score {
			return commits[cands[i].idx].Date.After(commits[cands[j].idx].Date)
		}
		return cands[i].score > cands[j].score
	})
	limit := 5
	if len(cands) < limit {
		limit = len(cands)
	}
	offenders := make([]model.Offender, 0, limit)
	for i := 0; i < limit; i++ {
		c := commits[cands[i].idx]
		offenders = append(offenders, model.Offender{
			SHA:     c.SHA,
			Subject: c.Subject,
			Date:    c.Date.Format(time.RFC3339),
			Reasons: cands[i].reasons,
			Score:   cands[i].score,
		})
	}
	return offenders
}

func max(a, b int) int {
	if a > b {
		return a
	}
	return b
}
