package analyze

import (
	"testing"

	"roastgit/internal/model"
)

func TestScoreMath(t *testing.T) {
	metrics := model.Metrics{
		Message: model.MessageMetrics{
			Total:     10,
			Generic:   5,
			EmojiOnly: 0,
			TooShort:  0,
			TooLong:   0,
			Lying:     0,
			Panic:     0,
		},
		Time: model.TimeMetrics{
			MidnightRatio:     0.1,
			DeadlineRatio:     0.0,
			LongestStreakDays: 5,
		},
		Hygiene: model.HygieneMetrics{
			MergeRatio:     0.3,
			BranchCount:    4,
			BadBranchCount: 1,
		},
		Size: model.SizeMetrics{
			SampleSize:        10,
			LargeCommitCount:  2,
			BinaryCommitCount: 1,
			AverageLines:      300,
		},
	}
	score := Score(metrics)
	if score.Breakdown.MessageQuality != 22 {
		t.Fatalf("expected message score 22, got %d", score.Breakdown.MessageQuality)
	}
	if score.Breakdown.Hygiene != 27 {
		t.Fatalf("expected hygiene score 27, got %d", score.Breakdown.Hygiene)
	}
	if score.Breakdown.Cadence == 0 || score.Breakdown.SizeDiscipline == 0 {
		t.Fatalf("expected non-zero cadence and size scores")
	}
	if score.Overall <= 0 {
		t.Fatalf("expected overall score")
	}
}
