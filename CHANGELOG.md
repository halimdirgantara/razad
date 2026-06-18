# Changelog

All notable changes to Razad will be documented in this file.

## Unreleased

### Added
- Live log streaming on the app detail page, including an in-page live log panel with connection state.
- App lifecycle hooks that attach and detach the log streamer when apps are started or removed.
- WebSocket routing safeguards so log upgrades remain compatible with request logging middleware.
- Regression coverage for WebSocket-capable routing and app lifecycle integration.
- AI orchestration scaffold with a protected `/api/v1/ai` endpoint, safe action registry, and audit logging.
- Database management API and UI for listing and provisioning database instances, including persisted connection details.

### Changed
- App detail page now connects to the backend WebSocket origin during local development so live logs work from Vite dev servers.
- Backend startup wires the observability log streamer into the app service.
- Backend startup now also wires the AI and database handlers into the main router.

### Verified
- `go test github.com/razad/razad/... -count=1 -short`
- `npm run check`
- `npm run build`
- Browser smoke test: login → open AI page / databases page → submit safe action / provision database
