# Architecture - MendoCultura

This document describes the architecture that exists in the current codebase. It should not describe ideal future architecture unless explicitly marked as future-ready or deferred.

## Overview

MendoCultura currently has three main parts:

```text
User browser
  -> Frontend Astro
  -> HTTP/JSON API
  -> Backend Go + Gin
  -> GORM
  -> PostgreSQL
```

Docker Compose defines local services for PostgreSQL, backend, and frontend.

## Runtime services

- `postgres`: PostgreSQL 16 Alpine, database `mendocultura`.
- `backend`: Go binary built from `backend/cmd/main.go`, exposed on port `8080`.
- `frontend`: Astro static build served by `frontend/server.mjs`, a small Node static server, exposed on port `4321`.

## Backend modules

Current backend structure:

```text
backend/
|-- cmd/main.go
|-- config/
|-- database/
|-- handlers/
|-- middleware/
|-- models/
|-- routes/
|-- services/
`-- utils/
```

### `cmd/main.go`

Application entrypoint:

1. Loads configuration.
2. Opens PostgreSQL connection.
3. Runs GORM migrations.
4. Runs seed data when `SEED_DATA=true`.
5. Starts the Gin router.

### `config`

Reads environment variables:

- `PORT`
- `DATABASE_URL`
- `JWT_SECRET`
- `SEED_DATA`

### `database`

Opens the GORM PostgreSQL connection and runs AutoMigrate for the main domain models.
It also contains the demo seed routine used when `SEED_DATA=true`.

### `models`

Defines models, roles, and states:

- `User`
- `OrganizerProfile`
- `Event`
- `Purchase`
- `Ticket`
- `AuditLog`

### `handlers`

Contains:

- HTTP handlers.
- Consistent JSON handler responses through shared helpers.

### `middleware`

Contains:

- Auth middleware.
- Role middleware.
- Approved organizer middleware.
- CORS middleware.

### `routes`

Builds the Gin router, health endpoints and `/api/v1` route groups without
changing public URLs.

### `services`

Contains:

- bcrypt password hashing and checking.
- JWT HS256 generation and parsing.
- Signed ticket code generation and validation.

### `utils`

Contains shared HTTP response helpers used by handlers and middleware.

## Frontend pages

Current Astro pages:

```text
frontend/src/pages/
|-- index.astro
|-- events.astro
|-- event-detail.astro
|-- login.astro
|-- register.astro
|-- tickets.astro
|-- ticket-detail.astro
|-- validate.astro
|-- organizer.astro
`-- admin.astro
```

Current shared frontend files:

- `frontend/src/layouts/BaseLayout.astro`
- `frontend/src/styles/global.css`
- `frontend/server.mjs` for Docker static serving

The frontend uses a role-aware header. It reads the stored token/user, refreshes `/me` when possible, hides links that do not match the current role, and still relies on backend RBAC as the authoritative access control.

The frontend is mostly static Astro with browser-side scripts calling the backend API. Protected pages now show clear Spanish notices when there is no token, the role is not sufficient, the session is rejected, or the backend cannot be reached.

## Public routes

No JWT required:

- `GET /`
- `GET /health`
- `POST /api/v1/auth/register`
- `POST /api/v1/auth/register-organizer`
- `POST /api/v1/auth/login`
- `GET /api/v1/events`
- `GET /api/v1/events/:id`

## JWT-protected routes

Require header:

```text
Authorization: Bearer <token>
```

Routes:

- `GET /api/v1/me`
- `POST /api/v1/tickets/reserve`
- `GET /api/v1/tickets`
- `GET /api/v1/tickets/:id`
- `POST /api/v1/checkin`
- `GET /api/v1/organizer/events`
- `POST /api/v1/organizer/events`
- `PUT /api/v1/organizer/events/:id`
- `POST /api/v1/organizer/events/upload-image`
- `GET /api/v1/organizer/reports`
- `GET /api/v1/admin/users`
- `GET /api/v1/admin/organizers`
- `PUT /api/v1/admin/organizers/:id/status`
- `GET /api/v1/admin/events`
- `PUT /api/v1/admin/events/:id/status`
- `GET /api/v1/admin/metrics`
- `GET /api/v1/admin/audit-logs`

## Roles and permissions

### USER

- Can login.
- Can buy/reserve tickets.
- Can see only their own tickets.

### ORGANIZER

- Can access organizer routes only if approved.
- Can manage own events.
- Can see own basic sales reports.
- Can validate tickets when approved.

### VALIDATOR

- Prepared for check-in/validation flow.
- Can validate tickets through `/api/v1/checkin`.

### ADMIN

- Can access admin routes.
- Can see users, organizers, events, metrics, and audit logs.
- Can change organizer status.
- Can change event status.

## Authentication flow

```text
User submits email/password
  -> POST /api/v1/auth/login
  -> Backend finds user by email
  -> bcrypt checks password
  -> Backend verifies user status is ACTIVE
  -> Backend signs JWT with HS256
  -> Frontend stores token in localStorage
  -> Future requests send Authorization: Bearer <token>
```

The auth middleware parses the token and loads the user from the database. This means a suspended or inactive user can be blocked on the next protected request instead of waiting for token expiry.

## Purchase simulation flow

```text
Authenticated user completes the frontend checkout simulation
  -> selects ticket type and quantity
  -> reviews summary
  -> fills fake payment form without persisting card data
  -> confirms
  -> POST /api/v1/tickets/reserve
  -> Backend starts SERIALIZABLE DB transaction
  -> Backend locks event row with SELECT FOR UPDATE semantics
  -> Backend checks event is PUBLISHED and in the future
  -> Backend checks availableTickets >= quantity
  -> Backend decrements availableTickets
  -> Backend creates Purchase with status PAID
  -> Backend creates Ticket rows with status VALID
  -> Backend generates signed codes and saves them
  -> Backend writes audit log
```

This is intentionally a simulated paid purchase. Mercado Pago and webhooks are not implemented yet.

## Ticket generation flow

For every created ticket:

1. A placeholder ticket row is created to obtain the database ID.
2. `services.NewTicketCode` creates a code.
3. Code format begins with `MC-`.
4. Payload includes ticket ID and random bytes.
5. Payload is signed with HMAC-SHA256 using `JWT_SECRET`.
6. The signed code is saved in the ticket row.

The ticket code is a signed unique code. The frontend renders that code as a visual QR using the `qrcode` npm package. PDF generation remains deferred.

## Ticket validation flow

```text
Validator submits eventId + code
  -> POST /api/v1/checkin
  -> Backend validates code prefix and HMAC signature
  -> Backend extracts ticket ID
  -> Backend locks ticket row
  -> Backend checks eventId and code match
  -> Backend rejects USED, CANCELLED, or non-VALID tickets
  -> Backend checks event is PUBLISHED
  -> Backend sets ticket status to USED and usedAt timestamp
  -> Backend writes audit log
```

If the same ticket is validated again, the backend returns a conflict and does not mark it valid again.

## Main data model

### User

Represents platform users and roles. Important fields:

- `id`
- `name`
- `email`
- `passwordHash`
- `dni`
- `role`
- `status`

### OrganizerProfile

Stores organizer fiscal/business data and approval state.

- `userId`
- `businessName`
- `taxId`
- `locality`
- `status`

### Event

Represents events that can be sold.

- `organizerId`
- `title`
- `description`
- `category`
- `department`
- `venue`
- `startAt`
- `priceCents`
- `capacity`
- `availableTickets`
- `status`
- `imageUrl`
- `galleryImages`
- `mapUrl`
- `latitude` / `longitude`
- `extendedDescription`
- `importantInfo`
- `recommendations`
- `ticketType`

### Purchase

Represents a simulated purchase.

- `userId`
- `eventId`
- `quantity`
- `totalCents`
- `status`

### Ticket

Represents the digital ticket.

- `userId`
- `eventId`
- `purchaseId`
- `code`
- `status`
- `usedAt`

### AuditLog

Stores simple traceability events.

- `userId`
- `action`
- `entity`
- `entityId`
- `detail`

## States

### Event

- `DRAFT`
- `PUBLISHED`
- `PAUSED`
- `CANCELLED`
- `FINISHED`

### Purchase

- `PENDING`
- `PAID`
- `CANCELLED`
- `REJECTED`

### Ticket

- `PENDING`
- `VALID`
- `USED`
- `CANCELLED`
- `EXPIRED`

`VALID` is used as the operative confirmed ticket state for the current demo.

## Error handling

Backend error responses use a consistent JSON shape:

```json
{
  "error": "ERROR_CODE",
  "message": "Spanish user-facing message"
}
```

Common helpers exist in `utils/responses.go`:

- `badRequest`
- `conflict`
- `forbidden`
- `notFound`
- `unauthorized`
- `serverError`

Frontend pages generally read `data.message` and show it to the user in Spanish. The main forms also validate required fields before sending requests and show network/backend availability errors when fetch fails.

## Environment variables

Backend:

- `PORT`: backend port, default `8080`.
- `DATABASE_URL`: PostgreSQL connection string.
- `JWT_SECRET`: secret used for JWT and ticket-code HMAC.
- `SEED_DATA`: when `true`, creates demo data if the DB is empty.

Frontend:

- `PUBLIC_API_URL`: backend API base URL, default expected `http://localhost:8080/api/v1`.

If any environment variable changes, update `.env.example` and README.

## Overselling prevention

The critical purchase endpoint uses:

- A database transaction.
- Serializable isolation.
- Row lock through GORM `clause.Locking{Strength: "UPDATE"}`.
- Availability check before decrement.
- Atomic decrement inside the same transaction.

This protects the current simulated purchase flow from selling more tickets than available when PostgreSQL is used.

## Current technical decisions

- Use GORM AutoMigrate for this demo phase instead of manual migrations.
- Use local JWT implementation to avoid adding an extra dependency.
- Use `JWT_SECRET` for both JWT signing and ticket-code HMAC in the current version.
- Keep frontend lightweight and mostly static.
- Keep user-visible frontend text in Spanish.
- Keep internal code names in English.
- Keep payment simulated until real Mercado Pago can be implemented properly.

## Prepared for future growth

The current design leaves room for:

- Replacing simulated purchase with payment provider abstraction.
- Adding Mercado Pago webhooks.
- Adding pending purchase TTL and stock release.
- Rendering QR images from the signed ticket code.
- Generating PDF tickets.
- Sending emails.
- Adding OpenAPI documentation.
- Adding real migrations.
- Adding stricter production CORS and rate limiting.
- Adding more tests around auth, permissions, purchase, and validation.

## Demo UX implementation notes

- Public events can be filtered by search, category and department. The backend normalizes text by lowercasing, trimming spaces and folding common Spanish accents so `lujan` matches `Luján de Cuyo`.
- Event detail pages show hero image, gallery, key facts, important information, recommendations, generated Google Maps iframe and a sticky checkout card.
- Checkout remains simulated: the frontend shows fake payment fields but sends only `eventId` and `quantity` to the backend. Card number and CVV are never stored or sent.
- `tickets.astro` and `ticket-detail.astro` render a QR from `ticket.code`. Validation continues to use the signed code and marks tickets as `USED` exactly once.
- Organizer event forms expose the enriched `Event` fields while preserving GORM AutoMigrate compatibility.
- Organizer event forms keep URL-based images and also support one main image upload through multipart/form-data. Uploaded demo files are saved under `backend/uploads/events`, served publicly from `/uploads/events/...`, ignored by Git, and persisted in Docker with the `backend_uploads` volume.
