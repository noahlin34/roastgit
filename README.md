# roastgit

A CLI that analyzes your local Git history and prints a humorous roast report of your commit habits. It uses **only local Git data** (no network calls) and works offline.

Disclaimer: it's a joke tool. Don't be mean to coworkers.

## Features
- Commit message analysis (generic, emoji-only, too long, "lying", panic streaks)
- Cadence and timing analysis (midnight commits, deadline spikes, streaks)
- Repo hygiene signals (merge ratio, branch names)
- Chunkiness stats (large commits, binary blobs)
- JSON output for automation
- Roast intensity, wholesome mode, and profanity censor

## Install / Build
Requires Go 1.22+ and a local Git executable.

```bash
go build ./cmd/roastgit
```

This produces a `roastgit` binary in the repo root.

## Usage
```bash
./roastgit
./roastgit --since 2024-01-01 --until 2024-12-31
./roastgit --author "alex@example.com" --intensity 5
./roastgit --wholesome --censor
./roastgit --deep
./roastgit --json > report.json
```

## Example Output
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

## Performance Notes
- Uses streaming parsing of `git log` output for speed and low memory.
- For large repos, size analysis is sampled by default (up to 500 commits).
- Use `--deep` to analyze all commit sizes (slower but complete).

## Testing
```bash
go test ./...
```
