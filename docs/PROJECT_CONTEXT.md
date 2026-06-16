# Project Context - MendoCultura

## Project name

MendoCultura.

## Short description

MendoCultura is a web platform for selling, managing, and validating digital tickets for cultural, tourist, sports, gastronomic, and university events in Mendoza, Argentina.

## Problem solved

Small and medium event organizers in Mendoza often depend on expensive national ticketing platforms or informal sales through WhatsApp, Instagram, spreadsheets, or manual lists. That creates high commissions, weak control over capacity, duplicated tickets, poor traceability, and difficult door validation.

MendoCultura provides a local ticketing base where users can discover events, buy or reserve tickets, see their digital tickets, and validate them at the door using a unique signed code.

## Local Mendoza context

The system is designed for Mendoza events such as:

- Vendimia and harvest-related events.
- Winery and tourism experiences.
- Municipal festivals.
- University and cultural activities.
- Gastronomic and sports events.

The visual identity should feel local, cultural, touristic, modern, and clear without becoming hard to maintain.

## Current technical stack

- Backend: Go + Gin.
- Database: PostgreSQL.
- ORM: GORM.
- Frontend: Astro.
- Authentication: JWT HS256.
- Password hashing: bcrypt.
- Local infrastructure: Docker Compose.
- Version control target: GitHub.

Do not change this stack without a strong technical reason and explicit documentation.

## Main actors

- USER: final customer who explores events, buys tickets, and sees their own tickets.
- ORGANIZER: approved organizer who creates and manages their own events and views basic sales reports.
- VALIDATOR: door/check-in role prepared for ticket validation.
- ADMIN: platform administrator who can manage users, organizers, events, metrics, and audit logs.

## Current functionality

Implemented in the current codebase:

- Backend Go + Gin entrypoint.
- PostgreSQL connection through GORM.
- Automatic GORM migrations for the demo database.
- Docker Compose with PostgreSQL, backend, and frontend services.
- JWT authentication.
- bcrypt password hashing.
- RBAC middleware for roles.
- User registration and login.
- Organizer registration with pending approval state.
- Public event catalog.
- Public enriched event detail with image, gallery, map, important info, availability and checkout card.
- Multi-step simulated ticket purchase/reservation in the frontend.
- Transactional stock decrement using SERIALIZABLE isolation and row locking.
- Ticket generation with unique signed code.
- User ticket history with visual QR code.
- Ticket detail page with QR, signed code and validation shortcut.
- Online ticket validation.
- Prevention of double validation.
- Organizer panel with expanded event create/edit fields and sales cards.
- Organizer panel with main event image upload from a local file, preview before saving, and URL fallback.
- Administrator panel with metrics, users, organizers, events and audit views.
- Basic metrics.
- Basic audit log.
- Demo validator user for the validation flow.
- Astro frontend with user-visible text in Spanish and corrected UTF-8 encoding.
- Role-aware header navigation based on stored session and `/me` validation.
- Case-insensitive and accent-insensitive catalog filters.
- Global local Playfair Display typography and demo-focused visual polish for footer, checkout, event cards and forms.
- Mendoza demo seed events with images, categories, locations and richer descriptions.
- Clearer frontend error messages for network failures, invalid forms, protected pages, and expired/invalid sessions.
- README with verified Docker Compose and manual PowerShell instructions plus demo users.

## Future functionality

Planned or deferred features from the technical report:

- Real Mercado Pago sandbox integration.
- Payment webhooks.
- Confirmation emails.
- PDF ticket generation.
- Camera-based QR scanning.
- Offline validation mode with Service Worker.
- Advanced audit filters.
- Advanced technical metrics such as P95 latency, goroutines, DB connections, and error rate.
- OpenAPI documentation.
- CI/CD.
- gosec and golangci-lint.
- HTTPS/Nginx production deployment.
- PostgreSQL backups, WAL archiving, or replication.

Do not implement advanced features halfway. If a feature cannot be completed properly, keep it documented as DEFERRED in `docs/ROADMAP.md`.

## Current scope

Current scope is a functional demo base:

- Events can be listed and viewed.
- Users can register/login.
- A logged-in user can buy tickets through a simulated paid purchase.
- Ticket capacity is decremented safely at the database level.
- Tickets have a unique signed code.
- Users can see their own tickets.
- Organizers/admins can validate tickets online.
- A used ticket cannot be validated again.

Payment is simulated by design. The architecture is prepared for future real payment integration, but the current version must not claim that Mercado Pago, PDF, email, camera scanning, or offline mode are complete.

## Important project rules

- Visitors can see published events.
- Buying tickets requires login.
- Users can only see their own tickets.
- Organizers can only manage their own events unless the user is ADMIN.
- ADMIN can manage all platform data exposed by admin endpoints.
- Events have limited capacity.
- The system must not sell more tickets than available.
- Tickets have a unique signed code.
- Tickets have a clear status.
- A VALID ticket can be used once.
- A USED ticket cannot be validated again.
- A CANCELLED ticket must not validate.
- Critical purchase operations must protect against over-selling.
- Passwords must remain hashed.
- Protected routes must require JWT.
- Roles must be enforced centrally through middleware.

## Language conventions

- Code, folders, files, variables, functions, structs, JSON keys, components, and internal route concepts must be in English.
- User-visible frontend text must be in Spanish.
- Examples of visible Spanish text: `Iniciar sesión`, `Crear cuenta`, `Comprar entrada`, `Mis entradas`, `Validar entrada`, `Entrada válida`, `Entrada ya utilizada`, `Sin disponibilidad`, `Panel de organizador`, `Entradas vendidas`, `Cupos disponibles`, `Cerrar sesión`.

## Naming consistency

Use these names consistently:

- `ticket` for digital ticket/entry.
- `event` for event.
- `organizer` for organizer role/entity.
- `capacity` for total event capacity.
- `availableTickets` for remaining ticket availability.
- `purchase` for purchase transaction.

Avoid mixing ticket/entry/pass/reservation for the same concept unless a future design explicitly separates them.

## Things that must not be broken

- Backend must keep compiling with `go test ./...`.
- Frontend must keep building with `npm run build`.
- `docker-compose.yml` must keep PostgreSQL, backend, and frontend aligned.
- `backend/.env.example` and `frontend/.env.example` must remain enough to run locally.
- JWT-protected routes must continue to reject missing/invalid tokens.
- Role-protected routes must continue to enforce permissions.
- The ticket validation flow must continue to prevent double use.
- The purchase flow must continue to avoid over-selling.
- README must stay short and practical.
- Roadmap must stay updated after important changes.

## Decisions already taken

- Keep Go + Gin + PostgreSQL + Astro.
- Use GORM for the current demo persistence layer.
- Use automatic GORM migration for local/demo setup.
- Use JWT HS256 implemented locally instead of adding a JWT dependency.
- Use bcrypt for password hashing.
- Use `VALID` as the operative confirmed ticket state instead of duplicating semantics with `CONFIRMED`.
- Implement purchase as simulated `PAID` purchase for now.
- Prepare the model for future payments but do not pretend Mercado Pago is complete.
- Generate a signed ticket code with prefix `MC-` and render it as a visual QR in the frontend. PDF remains deferred.
- Use a basic audit log for important actions.
- Keep frontend static Astro pages that call the API from browser-side scripts, with reusable styling and page-local scripts.
- Keep CORS permissive for local development; revisit for production.

## Instructions for future AI work

Before making future changes:

1. Read these files first:
   - `docs/PROJECT_CONTEXT.md`
   - `docs/ROADMAP.md`
   - `docs/ARCHITECTURE.md`
   - `docs/RISK_ANALYSIS.md`
   - `README.md`
2. Inspect the real code before assuming architecture.
3. Do not change the stack without a documented justification.
4. Do not create giant files mixing responsibilities.
5. Keep architecture clear and easy for a student to understand.
6. Update `docs/ROADMAP.md` after every important change.
7. Mark completed tasks with `[x]` and state `DONE`.
8. Keep visible frontend text in Spanish.
9. Keep internal names in English.
10. Do not implement advanced features halfway.
11. Document every new environment variable in `.env.example` and README if it affects running the project.
12. Update README if the way to run the project changes.
13. If adding endpoints, document them in README or architecture docs as appropriate.
14. If adding seeds or demo users, document credentials clearly.
15. Avoid deleting existing files unless there is a clear reason.
