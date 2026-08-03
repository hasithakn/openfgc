# Build context: demo-artifacts/webhook-listener (populated by `demo.sh build`) — a single
# dependency-free Node script (uses only core `http`/`url` modules).
FROM node:22-alpine
WORKDIR /app
COPY webhook-listener.js /app/webhook-listener.js
EXPOSE 9091
CMD ["node", "webhook-listener.js"]
