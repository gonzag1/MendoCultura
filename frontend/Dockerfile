FROM node:22-alpine AS build

WORKDIR /app
ENV ASTRO_TELEMETRY_DISABLED=1
COPY package.json package-lock.json* ./
RUN npm install
COPY . .
ARG PUBLIC_API_URL=http://localhost:8080/api/v1
ENV PUBLIC_API_URL=$PUBLIC_API_URL
RUN npm run build

FROM node:22-alpine
WORKDIR /app
ENV ASTRO_TELEMETRY_DISABLED=1
COPY --from=build /app/dist ./dist
COPY server.mjs ./
EXPOSE 4321
CMD ["node", "server.mjs"]
