package roast

import (
	"roastgit/internal/model"
	"roastgit/internal/util"
)

// GenerateRoasts produces headline, section roasts, and tips.
func GenerateRoasts(metrics model.Metrics, score model.Score, intensity int, wholesome bool, censor bool, seed string) model.RoastOutput {
	level := util.ClampInt(intensity, 0, 5)
	headline := chooseHeadline(score.Overall, level, wholesome, seed)
	sections := map[string]string{
		"commit_messages": sectionMessage(metrics, level, wholesome),
		"time_cadence":    sectionTime(metrics, level, wholesome),
		"repo_hygiene":    sectionHygiene(metrics, level, wholesome),
		"chunkiness":      sectionChunkiness(metrics, level, wholesome),
	}
	tips := buildTips(metrics, wholesome)
	if censor {
		headline = Censor(headline)
		for k, v := range sections {
			sections[k] = Censor(v)
		}
		for i := range tips {
			tips[i] = Censor(tips[i])
		}
	}
	return model.RoastOutput{
		Headline: headline,
		Sections: sections,
		Tips:     tips,
	}
}

func chooseHeadline(overall, intensity int, wholesome bool, seed string) string {
	tier := "mid"
	switch {
	case overall >= 90:
		tier = "gold"
	case overall >= 75:
		tier = "green"
	case overall >= 60:
		tier = "yellow"
	case overall >= 40:
		tier = "orange"
	default:
		tier = "red"
	}
	if wholesome {
		return util.DeterministicChoice(seed+"wholesome"+tier, wholesomeHeadlines(tier))
	}
	if intensity == 0 {
		return util.DeterministicChoice(seed+"factual"+tier, factualHeadlines(tier))
	}
	if intensity <= 2 {
		return util.DeterministicChoice(seed+"light"+tier, lightHeadlines(tier))
	}
	if intensity <= 4 {
		return util.DeterministicChoice(seed+"savage"+tier, savageHeadlines(tier))
	}
	return util.DeterministicChoice(seed+"nuclear"+tier, nuclearHeadlines(tier))
}

func factualHeadlines(tier string) []string {
	switch tier {
	case "gold":
		return []string{"Pristine Git habits detected."}
	case "green":
		return []string{"Mostly clean Git behavior with a few quirks."}
	case "yellow":
		return []string{"Mixed Git habits with noticeable rough edges."}
	case "orange":
		return []string{"Significant Git turbulence detected."}
	default:
		return []string{"Git habits need serious attention."}
	}
}

func wholesomeHeadlines(tier string) []string {
	switch tier {
	case "gold":
		return []string{"You are a Git role model. Keep it up."}
	case "green":
		return []string{"Solid work. A few tweaks and you are spotless."}
	case "yellow":
		return []string{"You are close. A bit more care will shine."}
	case "orange":
		return []string{"You can turn this around with a few habits."}
	default:
		return []string{"No shame: every repo starts messy. You can improve fast."}
	}
}

func lightHeadlines(tier string) []string {
	switch tier {
	case "gold":
		return []string{"Your Git is so clean it squeaks."}
	case "green":
		return []string{"Mostly tidy, with a few crumbs under the sofa."}
	case "yellow":
		return []string{"Your Git is fine, but it keeps forgetting to floss."}
	case "orange":
		return []string{"Your Git is doing parkour. Mostly into walls."}
	default:
		return []string{"Your Git habits are auditioning for chaos."}
	}
}

func savageHeadlines(tier string) []string {
	switch tier {
	case "gold":
		return []string{"Your Git is so pristine it judges other repos."}
	case "green":
		return []string{"Decent, but you still commit like a caffeinated raccoon."}
	case "yellow":
		return []string{"Your commit history looks like a crime scene."}
	case "orange":
		return []string{"This repo is held together by duct tape and vibes."}
	default:
		return []string{"This history is what linters see in their nightmares."}
	}
}

func nuclearHeadlines(tier string) []string {
	switch tier {
	case "gold":
		return []string{"Even the bots are jealous of this history."}
	case "green":
		return []string{"Nice work, but the chaos is peeking through."}
	case "yellow":
		return []string{"Your Git log screams " + "\"" + "I ship on vibes" + "\"" + "."}
	case "orange":
		return []string{"Somehow both over-committed and under-explained."}
	default:
		return []string{"The log is a cursed anthology of " + "\"" + "fix" + "\"" + " and regret."}
	}
}

func sectionMessage(metrics model.Metrics, intensity int, wholesome bool) string {
	badRatio := ratio(metrics.Message.LowQuality, metrics.Message.Total)
	if wholesome {
		if badRatio < 0.2 {
			return "Commit messages are mostly clear and useful."
		}
		return "Commit messages could be more descriptive, but that is easy to fix."
	}
	if badRatio > 0.4 {
		return pickByIntensity(intensity, "Your commit messages are on a first-name basis with "+"\""+"fix"+"\""+".", "These messages are a fog machine for future you.")
	}
	if badRatio > 0.2 {
		return pickByIntensity(intensity, "A few messages read like placeholder text.", "Your messages flirt with ambiguity.")
	}
	return pickByIntensity(intensity, "Messages are mostly fine, with minor misdemeanors.", "Messages are crisp enough to survive code review.")
}

func sectionTime(metrics model.Metrics, intensity int, wholesome bool) string {
	if wholesome {
		if metrics.Time.MidnightRatio > 0.2 {
			return "You ship late nights sometimes. Remember sleep is a feature."
		}
		return "Cadence looks steady. Keep the rhythm."
	}
	if metrics.Time.MidnightRatio > 0.25 {
		return pickByIntensity(intensity, "Midnight commits detected. The gremlins approve.", "You and 2am are in a committed relationship.")
	}
	if metrics.Time.DeadlineRatio > 0.2 {
		return pickByIntensity(intensity, "Deadline scrambles spotted.", "Friday afternoon commits: bold choice.")
	}
	return pickByIntensity(intensity, "Cadence is reasonably sane.", "Your commit schedule is suspiciously normal.")
}

func sectionHygiene(metrics model.Metrics, intensity int, wholesome bool) string {
	badBranchRatio := ratio(metrics.Hygiene.BadBranchCount, metrics.Hygiene.BranchCount)
	if wholesome {
		if badBranchRatio > 0.2 {
			return "Branch names could be clearer, but the structure is there."
		}
		return "History structure looks healthy."
	}
	if badBranchRatio > 0.3 {
		return pickByIntensity(intensity, "Branch names look like secret passwords.", "Your branches are named like test files in a hurry.")
	}
	if metrics.Hygiene.MergeRatio > 0.6 {
		return pickByIntensity(intensity, "Merge commits everywhere.", "Your history is a bowl of spaghetti merges.")
	}
	return pickByIntensity(intensity, "Hygiene is mostly clean.", "History is tidy enough to eat off of.")
}

func sectionChunkiness(metrics model.Metrics, intensity int, wholesome bool) string {
	largeRatio := ratio(metrics.Size.LargeCommitCount, metrics.Size.SampleSize)
	if wholesome {
		if largeRatio > 0.2 {
			return "Some commits are big. Breaking them up could help reviews."
		}
		return "Commit sizes look manageable."
	}
	if largeRatio > 0.25 {
		return pickByIntensity(intensity, "Some commits are the size of small planets.", "Your commits need a loading screen.")
	}
	if metrics.Size.BinaryCommitCount > 0 {
		return pickByIntensity(intensity, "Binary blobs are lurking in history.", "Binary files in git: bold, but painful.")
	}
	return pickByIntensity(intensity, "Chunk sizes are reasonable.", "Commit sizes are not horrifying. Nice.")
}

func buildTips(metrics model.Metrics, wholesome bool) []string {
	tips := []string{}
	if metrics.Message.Generic > 0 {
		tips = append(tips, "Replace generic messages with intent and impact.")
	}
	if metrics.Message.TooShort > 0 {
		tips = append(tips, "Add a few words of context to short commits.")
	}
	if metrics.Time.MidnightRatio > 0.2 {
		tips = append(tips, "Consider earlier commits to reduce late-night risk.")
	}
	if metrics.Hygiene.BadBranchCount > 0 {
		tips = append(tips, "Use descriptive branch names so future you can find work.")
	}
	if metrics.Size.LargeCommitCount > 0 {
		tips = append(tips, "Split large commits into focused chunks for reviewability.")
	}
	if metrics.Size.BinaryCommitCount > 0 {
		tips = append(tips, "Avoid committing large binaries; use git-lfs or artifacts.")
	}
	if len(tips) < 3 {
		if wholesome {
			tips = append(tips, "Keep leaning into consistency and clarity.")
		} else {
			tips = append(tips, "Consistency beats cleverness in commit history.")
		}
	}
	if len(tips) > 6 {
		tips = tips[:6]
	}
	return tips
}

func pickByIntensity(intensity int, light, savage string) string {
	if intensity <= 2 {
		return light
	}
	return savage
}

func ratio(a, b int) float64 {
	if b == 0 {
		return 0
	}
	return float64(a) / float64(b)
}
