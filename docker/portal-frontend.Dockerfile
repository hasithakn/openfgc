# Build context: demo-artifacts/portal-frontend (populated by `demo.sh build` via
# `npm run build`, with VITE_API_BASE_URL baked in at build time) — just the static dist/.
FROM nginx:1.27-alpine
COPY dist/ /usr/share/nginx/html/
COPY nginx.conf /etc/nginx/conf.d/default.conf
EXPOSE 80
