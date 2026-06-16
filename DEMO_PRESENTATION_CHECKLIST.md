# Checklist de preparación para demo - MendoCultura

## 1. Conservar `<aside class="card">`

- [x] Hecho.
- No se eliminó ningún bloque `<aside class="card">`.
- Se mantuvieron los paneles existentes de login, registro y validación.

## 2. Footer

- [x] Hecho.
- Se ajustó el layout global para que el footer quede abajo cuando hay poco contenido.
- `main` ahora ocupa el espacio disponible y el footer queda al final normal cuando la página tiene mucho contenido.
- Se agregó información institucional breve, enlaces rápidos, ubicación y año actual.

## 3. Validaciones del checkout

- [x] Hecho.
- Se validan titular, número de tarjeta, vencimiento, CVV, DNI y email.
- La tarjeta exige solo números, 16 dígitos y control Luhn.
- El vencimiento acepta `MM/AA` o `MM/YYYY`, valida mes y evita fechas vencidas.
- CVV exige 3 o 4 números.
- DNI y email tienen validaciones básicas.
- No se envían ni guardan número de tarjeta ni CVV.

## 4. Desborde visual del checkout

- [x] Hecho.
- Se ajustaron `overflow`, ancho, botones, mensajes, resumen y confirmación.
- La confirmación queda contenida dentro de la card sticky.
- Se evitó renderizar mensajes fuera de `#checkoutSlot`.

## 5. Imagen desacomodada en Eventos

- [x] Hecho.
- Las imágenes de cards de eventos usan proporción consistente con `aspect-ratio`.
- Se aplicó `object-fit: cover` y `object-position: center` sin deformar imágenes.

## 6. Filtros por lugar/departamento

- [x] Hecho.
- Se agregó normalización en frontend: lowercase, trim y remoción de diacríticos.
- El backend conserva y refuerza la normalización.
- La búsqueda compara nombre, descripción, categoría, departamento, venue y dirección.

## 7. Carga de imagen de evento desde archivo

- [x] Hecho.
- El panel organizador mantiene la URL de imagen.
- Se agregó selector de archivo local, preview y subida al guardar.
- El backend agregó endpoint multipart `POST /api/v1/organizer/events/upload-image`.
- Se guardan imágenes en `backend/uploads/events` y se sirven desde `/uploads/events/...`.
- Docker persiste uploads con volumen `backend_uploads`.
- `.gitignore` evita versionar uploads locales.

## 8. Tipografía

- [x] Hecho.
- Se registró la fuente local `Playfair Display` desde `frontend/src/assets/fonts`.
- Se aplicó globalmente con fallbacks locales.

## 9. Mejoras visuales menores

- [x] Hecho.
- Se mejoró el footer, checkout, estados de error, preview de imagen, botones internos y consistencia de cards.
- No se realizó una reescritura grande ni se cambió el stack.

## 10. Verificaciones

- [x] `go test ./...`
- [x] `npm install` (completó, pero reportó 1 vulnerabilidad moderada y 3 altas en dependencias)
- [x] `npm run build`
- [x] `docker compose config`
- [x] `docker compose up --build -d`
- [x] Healthcheck y frontend respondieron en `http://localhost:8080/health` y `http://localhost:4321`.
- [x] Upload de imagen probado con una imagen PNG mínima y respuesta pública `200`.
- [ ] `npm audit --json` quedó pendiente porque el endpoint de audit no respondió en este entorno.

## 11. Documentación

- [x] README actualizado.
- [x] ROADMAP actualizado con tareas DONE.
- [x] ARCHITECTURE actualizado.
- [x] PROJECT_CONTEXT actualizado.
- [x] RISK_ANALYSIS actualizado.

## 12. Estado final

- [x] La demo queda más presentable para una exposición académica.
- [x] Login, roles, eventos, compra, tickets, QR, validación, panel admin y panel organizador se preservaron.
- [x] La compra sigue siendo simulada y no guarda datos reales de tarjeta.
