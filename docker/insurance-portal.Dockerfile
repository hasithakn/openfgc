# Build context: dpdp-insuarance-portal/ — installs dependencies from source with npm, no
# host Node/npm install needed (a plain Node/Express app, no compile step).
FROM node:22-alpine
WORKDIR /app
COPY package.json package-lock.json ./
RUN npm ci --omit=dev
COPY . .
EXPOSE 3020
CMD ["node", "server.js"]
