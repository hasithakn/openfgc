# Build context: portal/frontend/ — builds the static bundle from source with pnpm (via
# corepack, no host Node/pnpm install needed), then serves it with nginx.
FROM node:22-alpine AS builder
WORKDIR /src
RUN corepack enable
COPY package.json pnpm-lock.yaml ./
RUN corepack prepare pnpm@10.6.5 --activate && pnpm install --frozen-lockfile
COPY . .
# Baked in at build time — the browser calls these directly, so they must be the demo's
# actual host-published values, not container-internal hostnames.
ENV VITE_API_BASE_URL=http://localhost:8081
ENV VITE_ORG_ID=insurance.org
ENV VITE_AUTH_LOGOUT_ALLOWED_ORIGINS=https://wso2is:9443
RUN pnpm run build

FROM nginx:1.27-alpine
COPY --from=builder /src/dist/ /usr/share/nginx/html/
COPY nginx.conf /etc/nginx/conf.d/default.conf
EXPOSE 80
