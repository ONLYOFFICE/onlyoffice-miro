# HANDOVER: ONLYOFFICE app for Miro

## 1. Integration type

**Standalone service + Miro Web SDK app.** The code is not embedded into Miro. Miro loads our frontend
in an iframe (the app panel), and the backend is our own HTTP service.

| Component | Stack | What it does |
|---|---|---|
| `frontend/` | React 19 + Vite 6 + TS 5.7, zustand, i18next, Miro Web SDK v2 (`miro.js` from Miro's CDN) | Two entry points: `index.html` (headless, registers `icon:click` - `openPanel`) and `application.html` (the panel: file manager, file creation, settings). Served by an nginx image. |
| `backend/` | Go 1.24, Echo v4, uber/fx (DI), pgx (PostgreSQL), go-redis, golang-migrate | REST API for the panel, Miro OAuth, the `/editor` page, the Docs callback `/api/callback`, PDF conversion. |
| PostgreSQL | - | `authentications` (Miro OAuth tokens, AES-encrypted), `settings` (Docs address/header/secret per board, secret encrypted), `demos` (demo start per team). Migrations are embedded in the binary and applied on startup (`deployments/migrations`). |
| Redis | - | Rate limiting (per IP) + settings cache (5 min). The backend does not start without Redis (`log.Fatalf`). |

**ONLYOFFICE integration is with ONLYOFFICE Docs (Document Server) only**, through the standard Docs API:
- `/editor` renders `assets/views/editor.html`, loads `<DocsAddress>/web-apps/apps/api/documents/api.js`; the config is signed with the Docs JWT secret.
- Docs downloads the file via a public (temporary) Miro link and saves through `POST /api/callback` (status 2 - the backend PATCHes the file in Miro using the `url` from the callback).
- PDF conversion: `GET /api/files/convert` - Docs ConvertService (async).
- Server check when settings are saved: the Docs `version` command; **Docs 8.2 minimum** (`validateDocServerVersion`, `backend/pkg/service/settings/settings.go`).
- ONLYOFFICE demo server: address/header/secret come from the backend config (`DEMO_SERVER_*`), duration is `DEMO_SERVER_DAYS` (the frontend displays `VITE_ASC_DEMO_EXPIRATION_DAYS`). The demo is tied to the Miro **team** (`demos.team_id`) and cannot be restarted.

**Miro:** REST API v2 (`api.miro.com/v2`: boards, members, docs/files) + `v1/oauth-token`, OAuth 2.0 code flow (`/api/oauth` is the redirect URI), Web SDK v2 on the frontend (`getIdToken`, `board.getInfo`, events, notifications).

**Authentication (important to understand):**
1. The panel gets `miro.board.getIdToken()`, a JWT signed with the **Miro app's client secret**, and sends it in the `x-miro-signature` header. The backend validates it with the same `OAUTH_CLIENT_SECRET`.
2. Using `team_id+user_id` from the token, the backend looks up the stored Miro OAuth access token (refreshing it when expired). No token - the panel shows the installation screen (`VITE_MIRO_INSTALLATION_URL`).
3. The editor opens in a new tab, where neither cookies nor the idToken can be passed. So `/api/authorize` issues a short-lived (5 min) JWT, which goes into `?token=` on `/editor`.
4. Only the **board owner** can view/change settings (checked via the Miro API in `pkg/controller/settings`). On the frontend, `admin` means "GET /api/settings did not return 403".

Docs settings are stored **per board** (`team_id + board_id`), not per team.

Supported versions: ONLYOFFICE Docs ≥ 8.2 (enforced in code). Miro is cloud-only with no versions; Web SDK v2, REST v2.

**Environments.** CI has two target registries: `test` (private, `secrets.DOCKER_TEST_REGISTRY`) and `prod` (`docker.io/onlyoffice/miro-frontend`, `miro-backend`). So there is a test environment and production, both running from Docker images.

## 2. Key features (user scenarios)

1. **Installation/authorization**: on first open without an OAuth token, the installation screen - Miro OAuth - `/api/oauth` stores the tokens - redirect back to Miro.
2. **Settings (board owner)**: Docs address (https only), JWT secret, JWT header; Docs availability and version are checked on save; on error, the previous values are restored.
3. **Demo server**: a "connect to demo" checkbox for N days; after expiry, "demo is over"; if own server credentials are entered, the demo is "detached" (`demo_detached`).
4. **Board file list**: office files from the board (`cursor` pagination, search, dedup), refreshed by board events and by the refresh button.
5. **File creation**: docx/xlsx/pptx from `document-templates` templates matching the board locale (fallback `default`), placed on the board at a random point, then the view navigates to it.
6. **Editing/viewing**: opens in a new tab, co-editing, saved back to Miro through the callback. Edit/view permissions come from `document-formats` (`onlyoffice-docs-formats.json`). Chat and comments are **disabled** in the config.
7. **Navigate to**: zooms to the file on the board.
8. **Download as PDF**: conversion through Docs.
9. **Delete**: removes the item from the board (frontend, via the SDK).
10. **Localization**: frontend and backend support en, de, es, fr, ja_JP, ko, pl, pt_BR (`frontend/src/i18n/locales`, `backend/assets/translations`).

## 3. Build

**Toolchain:** Go 1.24 (Docker: `golang:1.24` - `alpine:3`), Node 20 (Docker: `node:20-alpine` - `nginx:alpine`). Verified locally: the Go build and `tsc` pass on Go 1.26 / Node 23.

**Submodules (required!)**: the backend will not build without them (`go:embed`):
- `backend/pkg/service/document/formats` - `ONLYOFFICE/document-formats` (currently v2.1.0)
- `backend/assets/templates` - `ONLYOFFICE/document-templates`, branch `main/new`

```
git submodule update --init --recursive
```
Both are public. The code has no private dependencies.

**Backend:**
```
cd backend
docker compose -f deployments/docker-compose.yml up -d   # postgres + redis + pgadmin (local)
cp config.example.yaml config.yaml                        # fill in; config.yaml is in .gitignore
make run          # air hot-reload (requires air), or: go run ./cmd/server
make test         # go test -race ./...
make check-style  # golangci-lint
make build        # cross-compile into build/dist
```
Config: YAML (`CONFIG_PATH`, defaults to `./config.yaml`) + environment variable overrides (`DB_*`, `REDIS_*`, `OAUTH_*`, `MIRO_*`, `SERVER_DOMAIN`, `CALLBACK_URL`, `CORS_*`, `RATE_LIMIT_*`, `DEMO_SERVER_*`, `LOGGER_*`). The port is hard-coded to `:8080`.

**Frontend:**
```
cd frontend
yarn install      # the repo has yarn.lock, but the Dockerfile runs npm install
yarn dev          # vite, port 3000
yarn build        # tsc -b && vite build (both html entry points)
yarn lint / yarn type-check
```
`.env` variables (baked in at build time!): `VITE_MIRO_ONLYOFFICE_BACKEND`, `VITE_MIRO_INSTALLATION_URL`, `VITE_ASC_DEMO_EXPIRATION_DAYS`. Each environment needs its own image build.

**Local development with Miro:** Miro requires https, so use a tunnel. `vite.config.mts` hard-codes `allowedHosts: ['nsnz.ngrok.io']`, the author's personal ngrok host; replace it with your own. Use a separate dev app in Miro Developer with the App URL pointing to the tunnel.

**CI:** `.github/workflows/build.yaml` is a manual `workflow_dispatch` only. It builds and pushes the Docker images `miro-frontend:<tag>` and `miro-backend:<tag>` to the test registry (`secrets.DOCKER_TEST_*`) or to `docker.io/onlyoffice` (`secrets.DOCKER_HUB_*`). The backend address, install URL and demo duration are passed as inputs (masked). There are no tests or linting in CI.

**Git:** the main repository is `git.onlyoffice.com/ONLYOFFICE/onlyoffice-miro` (Gitea, origin). PRs into `develop` go through Gitea and are merged by Sergey Linnik. `github.com/ONLYOFFICE/onlyoffice-miro` is the public mirror (the README points there for issues; GitHub Actions also run there).

**Secrets: where they are stored (values are not recorded):**
- GitHub Actions secrets: `DOCKER_TEST_REGISTRY`, `DOCKER_TEST_USERNAME`, `DOCKER_TEST_ACCESS_TOKEN`, `DOCKER_HUB_USERNAME`, `DOCKER_HUB_ACCESS_TOKEN`.
- Backend runtime: `OAUTH_CLIENT_ID/SECRET` (Miro app), `OAUTH_ENCRYPTION_SECRET` (AES key for tokens and Docs secrets in the DB; **changing it loses all stored data**), `DEMO_SERVER_SECRET`, `DB_PASSWORD`, `REDIS_PASSWORD`.

## 4. Marketplace publishing

Marketplace: **Miro Marketplace** (the only one).

How it works:
- A Miro app is an entry in Miro Developer (App URL, redirect URI, scopes, icon) + a Marketplace listing. Miro never receives the code: a "release" means deploying our images.
- **A code update that does not change scopes, App URL or the listing does not need Miro review**; deploying the images is enough.
- Miro review is needed for the first submission and when scopes, the listing (description, screenshots, privacy/terms/support links) or key URLs change. Submission goes through the Marketplace submission form in Miro Developer; review usually takes from a few days to a couple of weeks.
- Typical rejection reasons: excessive scopes, broken installation/authorization flow, missing privacy policy or support link, UI not following Miro guidelines (mirotone), panel errors on an empty board, http addresses.
- Scopes required by the code: `boards:read`, `boards:write` (+ identity for `v1/oauth-token`). Verify against the production app.

## 5. Implicit knowledge

- **The Miro idToken is signed with the app's client secret.** That is why `OAUTH_CLIENT_SECRET` is used for OAuth, for validating `x-miro-signature`, and for signing the internal 5-minute editor token. If the secret is changed in Miro, change it on the backend too.
- **The editor authenticates with a short-lived token in the URL**, not a cookie: cookies in a new tab opened from the Miro iframe are unreliable (third-party cookies). The token is refreshed on every open. The `authCookieName` constant and `VITE_ACS_COOKIE_NAME` are leftovers of the old scheme.
- **Document key** = md5(`fileID + modifiedAt`) (`document/generator.go`). After a save, Miro changes `modifiedAt`, so the key changes - a new session. Do not change the formula without understanding co-editing.
- **The file name is passed in the callback URL query** (`filename=`); otherwise the original name was lost on save.
- **Callback**: only `status == 2` is processed; the token in the body is validated with the board's secret (or the demo secret). For non-save statuses the backend always replies `{"error":"0"}`.
- **A 404 on the backend - redirect to miro.com** (`router.go`). This is intentional.
- **Docs addresses must be https**: the Miro panel runs over https and the browser would block mixed content.
- **Demo**: tied to the team, `started` is never reset, so the demo cannot be obtained twice. `demo_detached` exists so that entering own credentials disables the demo without losing its start date.
- **Retries**: the frontend uses exponential retries (files, settings); `editor.html` makes 3 reconnection attempts 5 s apart; `destroyEditor` is not called on cleanup.
- `experimental:items:update` is an experimental Web SDK event and may disappear.
- `mirotone@2.0.0` is loaded from unpkg in `application.html`, while `package.json` has `mirotone@5`.
- `@mirohq/websdk-types: latest` is not pinned.

## 6. Known issues and tech debt

- There were no tests before the handover. Smoke tests were added (`backend/**/smoke_test.go`, see below); the frontend has no tests.
- `GET /api/settings` returns the **decrypted Docs JWT secret** to the owner (needed for the form; only the board owner has access). A compromise, left as is.
- The auth middleware logs tokens at Info level (`authentication/base.go`).
- PDF conversion: `ToDocumentType(path.Ext(title))` receives the extension with a leading dot, so `filetype` is always `docx` (`pkg/controller/file/conversion.go`). Docs apparently detects the type from the content, but "Download as PDF" should be checked for xlsx/pptx (it is in the manual checklist).
- `callback.getSecretFromSettings` and `conversion` dereference `Demo.Started` without a nil check (editor/base do check).
- `settingsService.logEvent` is an empty switch, so the settings service effectively writes no logs.
- Duplicated code in `NewMapFormatManager`.
- `ExecuteWithTimeout` derives its context from `context.Background()` rather than the request, so client cancellation is not propagated.
- Frontend Dockerfile: `npm install` with only a `yarn.lock` present - non-reproducible builds.
- The working copy contains untracked artifacts: `frontend/console.log`, `backend/tmp`, `build/dist`, `frontend/dist` (ignored).
- `master` contains only the Initial commit; all work is in `develop` (150 commits), with no tags. Releases are effectively marked only by the image tag (`inputs.tag` in `build.yaml`). Recommendation: on the next release, merge `develop` - `master` and add a git tag matching the image tag.
- Unfinished branches: origin (Gitea) has only `master` and `develop`; the feature branches (`feature/changed-keys`, `feature/translations`, `feature/update-readme`) were merged and deleted. There are no unfinished branches. Apart from the files added for this handover, there are no uncommitted changes.

## 7. Plans and commitments

Sensible next steps (from the tech debt, not commitments):
- fix `filetype` in conversion and the `Demo.Started` nil checks;
- remove tokens from Info logs;
- pin `@mirohq/websdk-types`, switch the Dockerfile to `yarn install --frozen-lockfile`;
- add `go test` and `yarn lint`/`type-check` to CI;
- multi-arch backend image;
- PDF form creation (`new.pdf` templates already exist in the submodule but are not exposed in the UI).

## Smoke tests

Automated (no DB/Redis/Miro needed): `cd backend && go test ./...`; see `*smoke_test.go`.
Covered: templates and assets are embedded (submodules), locale - template mapping, editor config permissions and signature, document key, the Docs 8.2 version threshold, callback responses, choosing the demo vs. own server, file name escaping in the callback URL.

Not automatable (needs live Miro + Docs): a **manual checklist** on the test environment, using a board you own and a second member account:
1. [ ] New board, app not authorized - installation screen - OAuth - back in Miro, the panel opens.
2. [ ] Owner with no settings - Settings opens. Non-owner - file manager without the gear icon.
3. [ ] Settings: an http address is rejected; wrong secret - error, fields revert; valid Docs ≥ 8.2 - saved; Docs < 8.2 - version error.
4. [ ] Demo: enable the checkbox - the editor opens on the demo server; enter own server - own server is used.
5. [ ] Create docx, xlsx, pptx - the file appears on the board and the panel navigates to it; with a de/ja board locale the template is localized.
6. [ ] Open a file - editor in a new tab (tab icon matches the type), edit - close - after ~10 s the file on the board is updated and its name is preserved.
7. [ ] Two users co-edit the same file.
8. [ ] A file uploaded to the board manually (docx and doc/odt) - visible after refresh; doc opens view-only.
9. [ ] Navigate to - zooms to the file. Delete - the file is removed from the board and the list.
10. [ ] Download as PDF for docx, xlsx, pptx (check that xlsx/pptx convert correctly).
11. [ ] List search, pagination (>50 files), refresh.
12. [ ] Miro locale de/fr/ja - panel texts and editor errors are translated.
13. [ ] Docs unavailable - the editor shows the error page and retries automatically.
14. [ ] `GET /health` - `{"status":"UP"}`; unknown path - redirect to miro.com.
