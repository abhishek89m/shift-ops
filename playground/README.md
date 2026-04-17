# Playground

Internal repo-local playground notes.

This repo keeps the split that worked well:

- `apps/web` = frontend playground + eventual product web app
- `services/api` = backend playground + eventual product API

Rule:

- do not recreate separate frontend/backend prototype copies outside these folders
- preview and test helpers should point at the real app/API here
- any temporary review shell should consume `apps/web` and `services/api`, not fork them

Use this as the mental model for the timed build:

- design and UX iteration land in `apps/web`
- API, schema, and validation land in `services/api`
- preview/test aids wrap the live outputs
