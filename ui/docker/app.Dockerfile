FROM node:24-alpine

WORKDIR /app/internal

COPY internal/package.json internal/package-lock.json ./
RUN npm ci

COPY internal ./
