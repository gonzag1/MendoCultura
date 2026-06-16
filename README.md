# MendoCultura

MendoCultura es una plataforma web para venta, gestión y validación de entradas digitales para eventos culturales, turísticos, deportivos, gastronómicos y universitarios de Mendoza.

## Inicio rápido

La forma recomendada para la demo es Docker Compose, porque levanta PostgreSQL, backend y frontend juntos.

### Forma 1 - Con Docker Compose

1. Abrí **Docker Desktop** desde Windows.
2. Esperá a que indique que Docker está corriendo.
3. Abrí PowerShell y verificá Docker:

```powershell
docker info
```

Si ese comando falla, Docker Desktop no está listo o no está corriendo.

4. Parate en la carpeta del proyecto:

```powershell
cd "C:\Users\ASUS\OneDrive\Documentos\MendoCultura"
```

5. Levantá todo:

```powershell
docker compose up --build
```

6. Abrí estas URLs:

- Frontend: http://localhost:4321
- Backend: http://localhost:8080
- Healthcheck: http://localhost:8080/health
- API base: http://localhost:8080/api/v1

7. Para detener el proyecto, en la terminal donde está corriendo presioná `Ctrl + C`. Después podés limpiar contenedores con:

```powershell
docker compose down
```

### Forma 2 - Manual si Docker falla

Esta forma requiere tener PostgreSQL disponible. Si Docker Desktop funciona parcialmente, podés levantar solo PostgreSQL con:

```powershell
cd "C:\Users\ASUS\OneDrive\Documentos\MendoCultura"
docker compose up postgres
```

Si Docker no funciona, necesitás PostgreSQL instalado localmente con estos datos:

- Host: `localhost`
- Puerto: `5432`
- Base de datos: `mendocultura`
- Usuario: `mendocultura`
- Contraseña: `mendocultura`

#### Backend manual

```powershell
cd "C:\Users\ASUS\OneDrive\Documentos\MendoCultura\backend"
Copy-Item .env.example .env -ErrorAction SilentlyContinue
$env:PORT="8080"
$env:DATABASE_URL="postgres://mendocultura:mendocultura@localhost:5432/mendocultura?sslmode=disable"
$env:JWT_SECRET="local-development-secret"
$env:SEED_DATA="true"
go run ./cmd
```

Backend esperado:

- http://localhost:8080
- http://localhost:8080/health

#### Frontend manual

En otra terminal PowerShell:

```powershell
cd "C:\Users\ASUS\OneDrive\Documentos\MendoCultura\frontend"
Copy-Item .env.example .env -ErrorAction SilentlyContinue
$env:PUBLIC_API_URL="http://localhost:8080/api/v1"
$env:ASTRO_TELEMETRY_DISABLED="1"
npm install
npm run dev
```

Frontend esperado:

- http://localhost:4321

## Usuarios de prueba

Con `SEED_DATA=true`, el backend crea datos iniciales si la base está vacía.

| Rol | Email | Contraseña |
|---|---|---|
| Usuario | `usuario@mendocultura.local` | `usuario123` |
| Organizador | `organizador@mendocultura.local` | `organizador123` |
| Validador | `validador@mendocultura.local` | `validador123` |
| Administrador | `admin@mendocultura.local` | `admin123` |

## Flujo demo recomendado

1. Entrar a http://localhost:4321.
2. Ir a **Eventos**.
3. Probar filtros por nombre, categoría o departamento, incluso sin tildes.
4. Entrar al detalle de un evento.
5. Iniciar sesión como usuario demo.
6. Completar la compra simulada con selector, resumen y formulario de pago ficticio.
7. Ir a **Mis entradas**.
8. Ver el QR visual o abrir el detalle de la entrada.
9. Iniciar sesión como organizador, validador o admin para validar.
10. Validar la entrada una vez.
11. Intentar validarla de nuevo para comprobar el bloqueo de doble validación.

## Stack

- Backend: Go + Gin + GORM.
- Frontend: Astro.
- Base de datos: PostgreSQL.
- Autenticación: JWT HS256 + bcrypt.
- Infraestructura local: Docker Compose.
- QR visual: librería frontend `qrcode` renderizando el código firmado del ticket.

## Estructura rápida

```text
backend/              Backend Go
  cmd/main.go         Punto de entrada
  config/             Variables de entorno y configuración
  database/           Conexión, AutoMigrate y datos iniciales
  models/             Modelos de dominio y estados
  handlers/           Handlers HTTP por recurso
  middleware/         Autenticación, roles, organizador aprobado y CORS
  routes/             Armado del router y endpoints
  services/           Helpers de seguridad, JWT, bcrypt y códigos de ticket
  utils/              Respuestas JSON compartidas
frontend/             Frontend Astro
  src/pages/          Páginas públicas, paneles y tickets
  src/layouts/        Layout base con navegación por rol
  src/styles/         Estilos globales
frontend/server.mjs   Servidor estático para Docker
docs/                 Contexto, roadmap, arquitectura y riesgos
docker-compose.yml    PostgreSQL + backend + frontend
```

## Endpoints principales

Públicos:

- `POST /api/v1/auth/register`
- `POST /api/v1/auth/register-organizer`
- `POST /api/v1/auth/login`
- `GET /api/v1/events`
- `GET /api/v1/events/:id`

Con JWT:

- `GET /api/v1/me`
- `POST /api/v1/tickets/reserve`
- `GET /api/v1/tickets`
- `GET /api/v1/tickets/:id`
- `POST /api/v1/checkin`

Organizador:

- `GET /api/v1/organizer/events`
- `POST /api/v1/organizer/events`
- `PUT /api/v1/organizer/events/:id`
- `POST /api/v1/organizer/events/upload-image`
- `GET /api/v1/organizer/reports`

Admin:

- `GET /api/v1/admin/users`
- `GET /api/v1/admin/organizers`
- `PUT /api/v1/admin/organizers/:id/status`
- `GET /api/v1/admin/events`
- `PUT /api/v1/admin/events/:id/status`
- `GET /api/v1/admin/metrics`
- `GET /api/v1/admin/audit-logs`

## Documentación interna

Antes de seguir desarrollando, leer:

- `docs/PROJECT_CONTEXT.md`
- `docs/ROADMAP.md`
- `docs/ARCHITECTURE.md`
- `docs/RISK_ANALYSIS.md`

Después de cambios importantes, actualizar especialmente `docs/ROADMAP.md`.

## Verificaciones útiles

Backend:

```powershell
cd "C:\Users\ASUS\OneDrive\Documentos\MendoCultura\backend"
go test ./...
```

No configures `GOCACHE` dentro del repositorio. Las carpetas `.gocache*`,
`frontend/node_modules`, `frontend/dist` y `frontend/.astro` son generadas y
deben quedar fuera del árbol versionable.

Frontend:

```powershell
cd "C:\Users\ASUS\OneDrive\Documentos\MendoCultura\frontend"
$env:ASTRO_TELEMETRY_DISABLED="1"
npm run build
```

Docker Compose config:

```powershell
cd "C:\Users\ASUS\OneDrive\Documentos\MendoCultura"
docker compose config
```

## Alcance actual

La versión actual es una demo funcional con compra simulada mejorada. Incluye catálogo visual, navegación por rol, detalle enriquecido de evento, campos ampliados para organizadores, filtros case-insensitive/accent-insensitive, carga de imagen principal desde archivo para eventos y QR visual generado desde el código firmado del ticket.

No incluye todavía Mercado Pago real, webhooks, PDF, email, cámara QR ni modo offline. Esas mejoras siguen documentadas como futuras en `docs/ROADMAP.md`.
