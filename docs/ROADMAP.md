# Roadmap - MendoCultura

This roadmap is the main tracking document for the project. Keep it updated after every important change.

Status values:

- TODO: not started.
- IN_PROGRESS: started but not complete.
- DONE: implemented and checked.
- BLOCKED: cannot progress due to an external blocker.
- DEFERRED: intentionally postponed.

Priority values:

- P0: critical / needed for the system to work.
- P1: important for the delivery.
- P2: useful improvement.
- P3: future / post-demo.

## P0 - Crítico

- [x] DONE - Backend Go + Gin base created.
- [x] DONE - PostgreSQL integration through GORM.
- [x] DONE - Automatic database migration for demo/local usage.
- [x] DONE - Docker Compose file with PostgreSQL, backend, and frontend services.
- [x] DONE - JWT HS256 authentication.
- [x] DONE - bcrypt password hashing.
- [x] DONE - Roles USER, ORGANIZER, VALIDATOR, and ADMIN.
- [x] DONE - User registration and login.
- [x] DONE - Organizer registration with pending approval data model.
- [x] DONE - Approved organizer check before organizer actions.
- [x] DONE - Public event catalog.
- [x] DONE - Public event detail.
- [x] DONE - Simulated ticket purchase.
- [x] DONE - Overselling protection with SERIALIZABLE transaction and row lock.
- [x] DONE - Tickets with unique signed code.
- [x] DONE - User ticket list / Mis entradas.
- [x] DONE - Online ticket validation.
- [x] DONE - Double validation blocking.
- [x] DONE - Basic organizer panel.
- [x] DONE - Basic admin panel.
- [x] DONE - Basic metrics endpoint.
- [x] DONE - Basic audit log.
- [x] DONE - Astro frontend with visible Spanish text.
- [x] DONE - Initial README created.
- [x] DONE - Internal documentation files created.
- [x] DONE - Docker Compose config validated with `docker compose config`.
- [x] DONE - Frontend build validated after error-handling improvements.
- [x] DONE - npm audit currently reports 0 vulnerabilities.
- [x] DONE - Verify full startup with `docker compose up --build` while Docker Desktop is running.
- [x] DONE - Verify full end-to-end user flow: login -> event detail -> purchase -> my tickets -> validation -> double validation blocked.
- [x] DONE - Verify role permissions for USER, ORGANIZER, VALIDATOR, and ADMIN in the demo flow.
- [x] DONE - Review that `backend/.env.example` and `frontend/.env.example` are sufficient on a fresh machine.
- [x] DONE - Improve clear visual error handling in frontend forms and protected pages.
- [x] DONE - Correct UTF-8 encoding in Spanish visible text and documentation.
- [x] DONE - Restore broken JavaScript ternaries and backend SQL placeholders after encoding cleanup.

## P1 - Importante

- [x] DONE - Rediseño visual general del frontend con estática mendocina moderna.
- [x] DONE - Navegación por rol en header y protección por backend/frontend de rutas sensibles.
- [x] DONE - Detalle de evento enriquecido con imagen, galería, ubicación, mapa, cupos, estado e información útil.
- [x] DONE - Checkout simulado mejorado con pasos de selección, resumen, pago ficticio y confirmación.
- [x] DONE - Campos ampliados de evento para organizadores: imágenes, galería, dirección, mapa, descripción completa, tipo de entrada e información importante.
- [x] DONE - Improve catalog filters with clearer UI for category and department.
- [x] DONE - Filtros case-insensitive/accent-insensitive por nombre, categoría, departamento, venue y dirección.
- [x] DONE - QR visual en Mis entradas generado desde el código firmado del ticket.
- [x] DONE - Página de detalle de entrada con QR, código, estado y acceso a validación.
- [x] DONE - Add more Mendoza demo seed events and categories.
- [x] DONE - Improve complete event editing from the organizer panel.
- [x] DONE - Improve organizer dashboard with clearer sales cards and event status controls.
- [x] DONE - Improve admin dashboard with clearer user, organizer, event, metric and audit views.
- [x] DONE - Improve frontend form validation before sending API requests.
- [x] DONE - Add clearer handling for expired sessions in the frontend.
- [ ] TODO - Add a clear screen for approving or rejecting organizers.
- [ ] TODO - Document admin and organizer workflows with short examples.
- [ ] TODO - Review CORS configuration before any non-local deployment.

## P2 - Mejoras

- [x] DONE - Improve responsive UI details on small screens.
- [ ] TODO - Add pagination to admin lists, audit logs, and public event catalog.
- [ ] TODO - Add endpoint documentation with request/response examples.
- [ ] TODO - Add backend unit tests for security helpers and ticket validation rules.
- [ ] TODO - Add backend handler/service tests for critical flows.
- [ ] TODO - Add a simple frontend smoke test or documented manual QA checklist.
- [ ] TODO - Add more precise audit log filters.
- [ ] TODO - Add safer production-ready configuration notes.
- [ ] DEFERRED - Generate PDF ticket files.
- [ ] DEFERRED - Send confirmation emails.

## P3 - Futuro / post-demo

- [ ] DEFERRED - Real Mercado Pago sandbox integration.
- [ ] DEFERRED - Payment webhooks and pending-ticket TTL release flow.
- [ ] DEFERRED - Camera-based QR validation.
- [ ] DEFERRED - Offline validation mode with Service Worker.
- [ ] DEFERRED - Conflict resolution strategy for offline validation.
- [ ] DEFERRED - OpenAPI generation.
- [ ] DEFERRED - CI/CD pipeline.
- [ ] DEFERRED - gosec and golangci-lint integration.
- [ ] DEFERRED - ESLint or frontend linting pipeline.
- [ ] DEFERRED - HTTPS/Nginx production deployment.
- [ ] DEFERRED - PostgreSQL backup, WAL archiving, or replication plan.
- [ ] DEFERRED - Advanced technical metrics: P95 latency, goroutines, DB connections, error rate.
- [ ] DEFERRED - Rate limiting and additional OWASP hardening.
