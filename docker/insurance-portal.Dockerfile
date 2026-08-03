# Build context: demo-artifacts/insurance-portal (populated by `demo.sh build`, which already
# ran `npm ci --omit=dev` there) — source + node_modules are both prebuilt, so this image just
# copies and runs, no npm install at image-build time.
FROM node:22-alpine
WORKDIR /app
COPY . /app
EXPOSE 3020
CMD ["node", "server.js"]
