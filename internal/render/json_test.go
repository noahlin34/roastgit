package render

import (
	"encoding/json"
	"testing"

	"roastgit/internal/model"
)

func TestJSONRender(t *testing.T) {
	report := model.Report{
		Repo:      model.RepoInfo{Path: "/tmp/repo", Name: "repo", Head: "abc", CommitCount: 2},
		Filters:   model.Filters{TZ: "local"},
		Score:     model.Score{Overall: 50, Breakdown: model.ScoreBreakdown{MessageQuality: 10, Hygiene: 10, Cadence: 10, SizeDiscipline: 20}},
		Metrics:   model.Metrics{},
		Offenders: []model.Offender{},
		Roasts:    model.RoastOutput{Headline: "hi", Sections: map[string]string{}, Tips: []string{}},
	}
	out, err := JSON(report)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	var decoded map[string]any
	if err := json.Unmarshal([]byte(out), &decoded); err != nil {
		t.Fatalf("invalid json: %v", err)
	}
	if decoded["repo"] == nil {
		t.Fatalf("expected repo in json")
	}
}
