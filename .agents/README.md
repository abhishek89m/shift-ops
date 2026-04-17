# AI Workspace

Base AI workspace files live here.

Current layout:

- `skills/`
- `subagents/`
- `context/`

`context/` keeps repo-local working memory that should travel with the output repo:

- decision log
- AI audit log

Rule:

- keep these docs lean and refreshed
- mirror only what helps future agents or reviewers inside this repo
- avoid depending on the larger root prep tree for repo-local context
