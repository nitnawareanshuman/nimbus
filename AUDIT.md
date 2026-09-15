# Nimbus audit — September 15, 2026

Repository reviewed: `nitnawareanshuman/nimbus`, branch `nimbus`.

## Fixed in this bundle

1. **Redis no longer takes down the API.** PostgreSQL is the source of truth, but the old startup path exited when `REDIS_URL` was missing, malformed, or temporarily unavailable. `main.go` now treats Redis as an optional cache.
2. **Fresh production databases are initialized.** Local Docker runs the migration container, but the Render web service does not. `main.go` now creates the `codes` table if it is missing.
3. **Redirect server errors are no longer reported as 404.** `handler/redirect.go` now returns 404 only for a missing short code and 500 for database/service failures.
4. **Local short links no longer point at production by default.** `.env.example` now uses `BASE_URL=http://localhost:8080`.
5. **Right-click failures are visible instead of appearing to do nothing.** Chrome extension service-worker fetches can be terminated when a response takes more than 30 seconds, while a sleeping free Render service can take long enough to hit that limit. `extension/background.js` now aborts at 25 seconds and shows a specific retry message, while preserving badge/notification feedback.
6. **Extension version is bumped to 1.4.1** because background runtime code changed.
7. **Documentation/checklist updated** for the production environment, optional Redis behavior, and extension version.

## Reviewed and not changed

- `handler/shorten.go`: validates JSON, non-empty input, HTTP/HTTPS scheme, and host before storage.
- `service/code.go`: uses cryptographically secure random six-character codes, handles collisions, PostgreSQL as source of truth, and Redis only as cache during request processing.
- `migrations/000001_create_codes.*.sql`: schema matches the runtime table definition.
- `docker-compose.yml`: correctly starts migrations before the API and provides local PostgreSQL/Redis health checks.
- `Dockerfile`: multi-stage static Go build is suitable for Render.
- `extension/popup.html`, `popup.css`, `popup.js`: no blocking bug found in the current branch.
- `extension/manifest.json`: Manifest V3 permissions are appropriately scoped to the current functionality; only the version was changed here.
- `chrome-web-store/privacy-policy.html` and `store-description.md`: disclosures are consistent with the extension behavior.
- Icon/store binary assets were checked for presence in the repository; their visual design was not modified by this bug-fix bundle.

## Validation performed

- `gofmt` parsing/format check passed for the modified Go files.
- `node --check` passed for the modified background service worker.
- `manifest.json` parses as valid JSON and its description is within Chrome Web Store length guidance.

A complete dependency build was not run in the isolated execution environment because the repository dependencies cannot be downloaded there. The repository's current `go.mod`/`go.sum` were left unchanged.
