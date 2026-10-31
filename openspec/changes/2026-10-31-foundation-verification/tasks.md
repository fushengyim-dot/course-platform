## 1. Shared page
- [x] 1.1 Extract and reuse the status component for / and /status/details, with explicit gateway routing.
- [x] 1.2 Verify shared rendering, state handling and unknown-route behavior with automated rendering tests and actual gateway HTTP requests.
- [ ] 1.3 Verify both pages in a browser, including narrow layout, keyboard navigation and retry interaction. Browser connection unavailable; see docs/verification.md.
## 2. Isolated contract
- [x] 2.1 Implement repository interface, isolated adapter, service and GET/POST handlers.
- [x] 2.2 Test normal, failure, boundary, ownership and concurrent idempotency scenarios.
## 3. Verification
- [x] 3.1 Extend architecture and mutation checks to the example layers.
- [x] 3.2 Run one-command checks and strict specification validation.
- [x] 3.3 Record reproducible verification commands and results in docs/verification.md.
