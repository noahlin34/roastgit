package analyze

import (
	"fmt"
	"math"

	"roastgit/internal/model"
	"roastgit/internal/util"
)

// Score computes the overall score and breakdown.
func Score(metrics model.Metrics) model.Score {
	msg := scoreMessageCategory(metrics)
	hyg := scoreHygiene(metrics)
	cad := scoreCadence(metrics)
	size := scoreSize(metrics)
	overall := msg + hyg + cad + size
	return model.Score{
		Overall: overall,
		Breakdown: model.ScoreBreakdown{
			MessageQuality: msg,
			Hygiene:        hyg,
			Cadence:        cad,
			SizeDiscipline: size,
		},
		Explain: explainScore(metrics, msg, hyg, cad, size, overall),
	}
}

func scoreMessageCategory(metrics model.Metrics) int {
	if metrics.Message.Total == 0 {
		return 30
	}
	total := float64(metrics.Message.Total)
	genericRatio := float64(metrics.Message.Generic) / total
	emojiRatio := float64(metrics.Message.EmojiOnly) / total
	shortRatio := float64(metrics.Message.TooShort) / total
	longRatio := float64(metrics.Message.TooLong) / total
	lyingRatio := float64(metrics.Message.Lying) / total
	panicRatio := float64(metrics.Message.Panic) / total
	penalty := genericRatio*15 + emojiRatio*5 + shortRatio*5 + longRatio*3 + lyingRatio*7 + panicRatio*5
	score := 30 - int(math.Round(penalty))
	return util.ClampInt(score, 0, 30)
}

func scoreHygiene(metrics model.Metrics) int {
	penalty := 0.0
	if metrics.Hygiene.MergeRatio > 0.4 {
		extra := (metrics.Hygiene.MergeRatio - 0.4) / 0.6
		penalty += extra * 6
	}
	if metrics.Hygiene.BranchCount > 0 {
		badRatio := float64(metrics.Hygiene.BadBranchCount) / float64(metrics.Hygiene.BranchCount)
		penalty += badRatio * 10
	}
	score := 30 - int(math.Round(penalty))
	return util.ClampInt(score, 0, 30)
}

func scoreCadence(metrics model.Metrics) int {
	penalty := metrics.Time.MidnightRatio*8 + metrics.Time.DeadlineRatio*5
	if metrics.Time.LongestStreakDays > 14 {
		over := float64(metrics.Time.LongestStreakDays - 14)
		penalty += math.Min(5, over/7*5)
	}
	score := 20 - int(math.Round(penalty))
	return util.ClampInt(score, 0, 20)
}

func scoreSize(metrics model.Metrics) int {
	penalty := 0.0
	if metrics.Size.SampleSize > 0 {
		largeRatio := float64(metrics.Size.LargeCommitCount) / float64(metrics.Size.SampleSize)
		binaryRatio := float64(metrics.Size.BinaryCommitCount) / float64(metrics.Size.SampleSize)
		penalty += largeRatio * 10
		penalty += binaryRatio * 5
		if metrics.Size.AverageLines > 400 {
			over := metrics.Size.AverageLines - 400
			penalty += math.Min(5, over/400*5)
		}
	}
	score := 20 - int(math.Round(penalty))
	return util.ClampInt(score, 0, 20)
}

func explainScore(metrics model.Metrics, msg, hyg, cad, size, overall int) map[string]string {
	explain := map[string]string{}
	explain["message_quality"] = fmt.Sprintf("30 - round(generic%%*15 + emoji%%*5 + short%%*5 + long%%*3 + lying%%*7 + panic%%*5) = %d", msg)
	explain["hygiene"] = fmt.Sprintf("30 - round(mergePenalty + badBranchPenalty) = %d", hyg)
	explain["cadence"] = fmt.Sprintf("20 - round(midnight%%*8 + deadline%%*5 + streakPenalty) = %d", cad)
	explain["size_discipline"] = fmt.Sprintf("20 - round(large%%*10 + binary%%*5 + avgLinesPenalty) = %d", size)
	explain["overall"] = fmt.Sprintf("message + hygiene + cadence + size = %d", overall)
	return explain
}
