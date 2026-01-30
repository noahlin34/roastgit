package model

import "time"

// Config controls analysis and rendering behavior.
type Config struct {
	Path       string
	Since      string
	Until      string
	Author     string
	JSON       bool
	NoColor    bool
	Intensity  int
	Wholesome  bool
	Censor     bool
	Deep       bool
	MaxCommits int
	TZ         string
	Explain    bool
}

// RepoInfo describes the repository under analysis.
type RepoInfo struct {
	Path        string `json:"path"`
	Name        string `json:"name"`
	Head        string `json:"head"`
	CommitCount int    `json:"commit_count"`
}

// Filters capture applied filters.
type Filters struct {
	Since      string `json:"since,omitempty"`
	Until      string `json:"until,omitempty"`
	Author     string `json:"author,omitempty"`
	MaxCommits int    `json:"max_commits,omitempty"`
	TZ         string `json:"tz"`
	Deep       bool   `json:"deep"`
}

// Commit represents a single commit.
type Commit struct {
	SHA         string
	AuthorName  string
	AuthorEmail string
	Date        time.Time
	Subject     string
	Parents     []string
	Size        *CommitSize
}

// CommitSize summarizes file/line changes.
type CommitSize struct {
	Files       int
	Added       int
	Deleted     int
	BinaryFiles int
}

// Metrics groups computed metrics.
type Metrics struct {
	Message MessageMetrics `json:"message"`
	Time    TimeMetrics    `json:"time"`
	Hygiene HygieneMetrics `json:"hygiene"`
	Size    SizeMetrics    `json:"size"`
}

// MessageMetrics captures commit message stats.
type MessageMetrics struct {
	Total           int      `json:"total"`
	Generic         int      `json:"generic"`
	EmojiOnly       int      `json:"emoji_only"`
	TooLong         int      `json:"too_long"`
	TooShort        int      `json:"too_short"`
	Lying           int      `json:"lying"`
	Panic           int      `json:"panic"`
	LowQuality      int      `json:"low_quality"`
	AverageLength   float64  `json:"average_length"`
	AverageQuality  float64  `json:"average_quality"`
	TopGenericWords []string `json:"top_generic_words,omitempty"`
}

// TimeMetrics captures cadence patterns.
type TimeMetrics struct {
	CommitsPerDayAvg  float64 `json:"commits_per_day_avg"`
	CommitsPerWeekAvg float64 `json:"commits_per_week_avg"`
	UniqueDays        int     `json:"unique_days"`
	UniqueWeeks       int     `json:"unique_weeks"`
	MidnightRatio     float64 `json:"midnight_ratio"`
	DeadlineRatio     float64 `json:"deadline_ratio"`
	LongestStreakDays int     `json:"longest_streak_days"`
}

// HygieneMetrics captures repo hygiene signals.
type HygieneMetrics struct {
	MergeRatio     float64  `json:"merge_ratio"`
	LinearRatio    float64  `json:"linear_ratio"`
	BranchCount    int      `json:"branch_count"`
	BadBranchCount int      `json:"bad_branch_count"`
	BadBranches    []string `json:"bad_branches,omitempty"`
}

// SizeMetrics captures commit size behavior.
type SizeMetrics struct {
	Sampled           bool    `json:"sampled"`
	SampleSize        int     `json:"sample_size"`
	LargeCommitCount  int     `json:"large_commit_count"`
	BinaryCommitCount int     `json:"binary_commit_count"`
	AverageLines      float64 `json:"average_lines"`
	MaxLines          int     `json:"max_lines"`
}

// ScoreBreakdown holds category scores.
type ScoreBreakdown struct {
	MessageQuality int `json:"message_quality"`
	Hygiene        int `json:"hygiene"`
	Cadence        int `json:"cadence"`
	SizeDiscipline int `json:"size_discipline"`
}

// Score summarizes scoring.
type Score struct {
	Overall   int               `json:"overall"`
	Breakdown ScoreBreakdown    `json:"breakdown"`
	Explain   map[string]string `json:"explain,omitempty"`
}

// Offender marks roastable commits.
type Offender struct {
	SHA     string   `json:"sha"`
	Subject string   `json:"subject"`
	Date    string   `json:"date"`
	Reasons []string `json:"reasons"`
	Score   int      `json:"-"`
}

// RoastOutput captures generated roast text.
type RoastOutput struct {
	Headline string            `json:"headline"`
	Sections map[string]string `json:"sections"`
	Tips     []string          `json:"tips"`
}

// Report is the full analysis output.
type Report struct {
	Repo      RepoInfo    `json:"repo"`
	Filters   Filters     `json:"filters"`
	Score     Score       `json:"score"`
	Metrics   Metrics     `json:"metrics"`
	Offenders []Offender  `json:"offenders"`
	Roasts    RoastOutput `json:"roasts"`
}
