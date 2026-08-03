# Build context: event-framework/ (where webhook-listener.js lives) — a single dependency-free
# Node script (uses only core `http`/`url` modules), no build step needed.
FROM node:22-alpine
WORKDIR /app
COPY webhook-listener.js /app/webhook-listener.js
EXPOSE 9091
CMD ["node", "webhook-listener.js"]
