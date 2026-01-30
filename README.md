# 😈 roastgit

**Roast your Git habits. Learn something. Laugh a little.**

`roastgit` is a fast, offline CLI that analyzes your local Git history and delivers a humorous roast report: commit message crimes, midnight gremlin stats, deadline scrambles, chunky commits, and more. It also ships JSON output for automation and a wholesome mode for kinder feedback.

> Disclaimer: it’s a joke tool. Don’t be mean to coworkers.

---

## ✨ What You Get
- **Commit message analysis**: generic, too-short/long, emoji-only, and “lying” messages.
- **Cadence insights**: commits/day, midnight gremlin score, deadline spikes, streaks.
- **Repo hygiene**: merge ratio, branch name quality, linearity.
- **Chunkiness**: large commits and binary blobs.
- **Configurable tone**: intensity 0–5, wholesome mode, and profanity censor.
- **Offline-only**: uses local git data — no network calls.

---

## 🚀 Quick Start

### Install / Build (Go 1.22+ required)
```bash
# from the repo root
go build ./cmd/roastgit
```
This produces a `roastgit` binary in the repo root.

### Run
```bash
./roastgit
```

---

## 🧰 Usage Examples
```bash
# analyze current repo
./roastgit

# date range filtering
./roastgit --since 2024-01-01 --until 2024-12-31

# filter by author
./roastgit --author "alex@example.com"

# crank the heat
./roastgit --intensity 5

# wholesome + censor
./roastgit --wholesome --censor

# full size analysis (slower)
./roastgit --deep

# JSON output
./roastgit --json > report.json
```

---

## 🧩 Flags
```
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
```

---

## 📝 Sample Output
```
Roastgit Report
Repo: example-repo (/path/to/example-repo)
HEAD: 1a2b3c4
Commits analyzed: 231
Range: 2024-01-01 -> 2024-12-31

Overall Score: 62/100 Your commit history looks like a crime scene.

Commit Message Crimes
- Generic messages: 54 (23%)
- Emoji-only: 3, too short: 12, too long: 7
- Lying messages: 4
A few messages read like placeholder text.

Time & Cadence
- Commits/day avg: 1.14
- Commits/week avg: 6.20
- Midnight commits: 18%
- Deadline window commits: 9%
- Longest streak: 9 days
Midnight commits detected. The gremlins approve.

Repo Hygiene
- Merge ratio: 22%
- Linear ratio: 78%
- Branches: 14 (bad: 2)
History is tidy enough to eat off of.

Chunkiness
- Large commits: 12/231
- Binary commits: 4/231
- Average lines changed: 212
- Max lines changed: 1250
Some commits are the size of small planets.

Top Offenders
- 7b3e9ad 2024-06-11 update -- generic message, huge commit
- 2afc010 2024-02-03 :fire: :fire: -- emoji-only message, midnight gremlin

Tips
- Replace generic messages with intent and impact.
- Split large commits into focused chunks for reviewability.
- Avoid committing large binaries; use git-lfs or artifacts.

Hints: try --deep for full size analysis, --wholesome for kinder output, or --json for machine use.
```

---

## ⚡ Performance Notes
- Uses streaming parsing of `git log` for speed and low memory.
- For large repos, size analysis is sampled by default (up to 500 commits).
- Use `--deep` for complete size stats (slower, but accurate).

---

## ✅ Testing
```bash
go test ./...
```

---

## 🛡️ Requirements
- **Go 1.22+**
- **Git executable** available on your PATH
- Works on **macOS, Linux, and Windows**

---

## 🤝 Contributing
See `AGENTS.md` for project guidelines and development notes.
