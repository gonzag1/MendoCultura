# Risk Analysis - MendoCultura

## Risk table

| ID | Risk | Probability | Impact | Priority | Mitigation | Status |
|---|---|---|---|---|---|---|
| R-01 | Docker Desktop is not running or fails on Windows. | Alta | Alto | P0 | Docker Compose startup was verified successfully; README documents Docker Desktop checks and manual fallback. | Mitigado |
| R-02 | Overselling tickets under concurrent purchases. | Media | Crítico | P0 | Current purchase flow uses PostgreSQL transaction with SERIALIZABLE isolation and row lock before decrementing availability. Add concurrency tests later. | Mitigado |
| R-03 | Double validation of tickets. | Media | Alto | P0 | Current check-in locks the ticket, rejects USED/CANCELLED/non-VALID states, then updates to USED. Add tests later. | Mitigado |
| R-04 | JWT remains valid after user is suspended. | Media | Alto | P0 | Auth middleware loads the user from DB and checks ACTIVE status on protected requests. | Mitigado |
| R-05 | Environment variables are misconfigured. | Media | Alto | P0 | `.env.example` files were reviewed and README includes Docker and manual PowerShell setup. Recheck on a fresh machine later. | Mitigado |
| R-06 | Moderate npm vulnerability detected by `npm install`. | Media | Medio | P1 | `npm audit --audit-level=moderate` now reports 0 vulnerabilities. | Cerrado |
| R-07 | Difference between current demo and full technical report creates false expectations. | Alta | Alto | P1 | Clearly document simulated purchase and deferred advanced features in README, PROJECT_CONTEXT, and ROADMAP. | Mitigado |
| R-08 | Mercado Pago and webhooks are more complex than expected. | Alta | Alto | P3 | Keep payments simulated now; later add provider abstraction, webhook signature validation, idempotency, and pending-ticket TTL. | Aceptado |
| R-09 | Frontend validation and error messages are not clear enough. | Media | Medio | P1 | Main frontend pages show clearer Spanish messages for missing forms, sessions, rejected roles, backend errors, network failures and empty states. | Mitigado |
| R-10 | Not enough automated tests for critical flows. | Alta | Alto | P1 | Add tests for security helpers, auth middleware, purchase transaction, role permissions, and check-in. | Pendiente |
| R-11 | OneDrive/Windows path or permission issues affect Go, npm, Docker, or Git. | Media | Medio | P1 | Use local Go cache when required; document Windows PowerShell commands. | Abierto |
| R-12 | Future inconsistencies if ROADMAP.md is not updated after changes. | Media | Medio | P1 | Roadmap was updated after the demo UX improvement pass; future work must keep it current. | Mitigado |
| R-13 | CORS is permissive for local development. | Media | Medio | P2 | Keep permissive CORS for demo only; restrict origins before production deployment. | Aceptado |
| R-14 | GORM AutoMigrate may be insufficient for controlled production schema changes. | Media | Medio | P2 | Use AutoMigrate for demo; introduce explicit migrations before production. | Aceptado |
| R-15 | Ticket code is shown but QR is not visible. | Alta | Medio | P1 | Frontend now renders a QR with `qrcode` in Mis entradas and ticket detail, using the same signed code accepted by validation. | Mitigado |
| R-16 | Offline validation can accept duplicates across multiple disconnected devices. | Media | Crítico | P3 | Defer offline mode until conflict resolution, token sync, revocation, and storage security are designed. | Aceptado |
| R-17 | UTF-8 encoding breaks Spanish visible text and docs. | Alta | Alto | P1 | Spanish text and documentation were corrected to UTF-8 with tildes, ñ and special characters. | Mitigado |
| R-18 | Catalog filters miss results due to uppercase/lowercase or accents. | Media | Medio | P1 | Backend normalizes search, category and department by lowercasing and folding accents before comparison. | Mitigado |
| R-19 | Simulated checkout looks too unrealistic for demo. | Media | Medio | P1 | Event detail now uses a multi-step simulated checkout without storing card data. | Mitigado |
| R-20 | Navigation shows links to roles that should not access them. | Media | Alto | P1 | Header visibility is driven by `/me`/localStorage role and backend RBAC remains authoritative on protected endpoints. | Mitigado |
| R-21 | UI is too empty or unclear for demo stakeholders. | Media | Medio | P1 | Frontend was redesigned with modern cards, hero, responsive layout, empty states, badges and clearer dashboards. | Mitigado |

## Current critical risks

1. **Automated tests are still limited**: backend compiles and frontend builds, but critical business flows still need automated tests.
2. **OneDrive/Windows tooling can still be slow**: `go test ./...` may require a local `GOCACHE` path for reliable execution.
3. **Demo vs full report scope must stay explicit**: Mercado Pago, PDF, email, camera QR, and offline mode are still deferred.
4. **CORS remains permissive for local development**: restrict it before production deployment.

## Risk handling rules for future work

- If a risk is mitigated by a code change, update this file and `docs/ROADMAP.md`.
- If a risk is accepted or deferred, explain why.
- Do not mark a risk closed unless it has been verified.
- Keep demo limitations explicit; do not present deferred features as complete.
- Prefer small, verifiable changes over large rewrites.
