# ONLYOFFICE app for Miro

## Commands

```bash
git submodule update --init --recursive   # REQUIRED: document-formats and document-templates are embedded via go:embed

# backend (Go 1.24)
cd backend
docker compose -f deployments/docker-compose.yml up -d   # postgres + redis
cp config.example.yaml config.yaml                        # fill in (in .gitignore)
go run ./cmd/server          # :8080; or `make run` (air)
make test                    # go test -race ./... (smoke tests, no DB/Redis needed)
make check-style             # golangci-lint (config in v1 format)
go build ./... && go vet ./...

# frontend (Node 20, yarn)
cd frontend
yarn install && yarn dev     # :3000; Miro requires https - use a tunnel
yarn lint && yarn type-check && yarn build
```

## Structure

- `backend/initializer/`: DI (uber/fx) and **all routes** (`router.go`).
- `backend/pkg/controller/{auth,callback,editor,file,settings}`: HTTP handlers; `base/` holds shared helpers.
- `backend/pkg/middleware/authentication`: validation of `x-miro-signature` / `?token=`.
- `backend/pkg/service/{document,settings,oauth,storage/pg,cache}`: editor config, per-board settings, OAuth tokens (AES), Postgres, Redis.
- `backend/pkg/client/{miro,docserver,oauth}`: external APIs.
- `backend/deployments/migrations`: SQL migrations, applied automatically on startup.
- `backend/assets/`: embedded views (`editor.html`), translations, icons, templates (submodule).
- `frontend/src/{app/pages,features/*,stores,api}`: pages, features (file, manager, settings), zustand stores.

## Conventions

- Commits: Conventional Commits (`fix(settings): ...`, `chore(i18n): ...`). Branches `feature/*` - PR into `develop`; `master` is the release branch.
- Every new file starts with the Apache-2.0 header `(c) Copyright Ascensio System SIA` (copy it from a neighboring file).
- UI strings go through i18n only: frontend `frontend/src/i18n/locales/*.json`, backend `backend/assets/translations/*.json`. A new key goes into all 8 languages.
- Frontend: alias imports (`@features/...`, `@stores/...`), components as arrow functions (airbnb + prettier).
- Backend: config = YAML + env overrides (`config/*.go`, each with `loadEnv` and `Validate`). A new setting - both places + `config.example.yaml`.
- DB schema changes only via a new `NNNNNN_*.up.sql/.down.sql` pair; never edit existing migrations.

## Don'ts / be careful

- Never commit `backend/config.yaml`, `frontend/.env` or any secrets. Never log tokens or secrets.
- Do not change the document key formula (`pkg/service/document/generator.go`: md5(id+modifiedAt)); it would break co-editing.
- `OAUTH_CLIENT_SECRET` both validates the Miro idToken and signs the editor token; `OAUTH_ENCRYPTION_SECRET` encrypts data in the DB, so changing it loses all stored tokens and secrets.
- Docs addresses must be https (checked on both frontend and backend).
- The callback (`/api/callback`) must reply `{"error":"0"}` to everything except a real error; otherwise Docs shows a save error.
- `VITE_*` variables are baked in at build time, so each environment needs its own frontend image.
- Do not edit the contents of submodule directories (`backend/assets/templates`, `backend/pkg/service/document/formats`) by hand; only update the commit reference.
- Releases go through the `release` skill (`.claude/skills/release/SKILL.md`).
