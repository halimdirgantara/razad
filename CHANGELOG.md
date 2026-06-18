# Changelog

All notable changes to Razad will be documented in this file.

## Unreleased

### Added
- Live log streaming on the app detail page, including an in-page live log panel with connection state.
- App lifecycle hooks that attach and detach the log streamer when apps are started or removed.
- WebSocket routing safeguards so log upgrades remain compatible with request logging middleware.
- Regression coverage for WebSocket-capable routing and app lifecycle integration.

### Changed
- App detail page now connects to the backend WebSocket origin during local development so live logs work from Vite dev servers.
- Backend startup wires the observability log streamer into the app service.

### Verified
- `go test github.com/razad/razad/... -count=1 -short`
- `go test github.com/razad/razad/internal/api -count=1`
- `npm run check`
- `npm run build`
- Browser smoke test: login → open app detail → deploy → live logs appear
