FROM node:20-alpine AS builder
WORKDIR /web
COPY package*.json ./
COPY .env ./
RUN npm install
COPY . .
RUN npm run build

FROM nginx:alpine
COPY --from=builder /web/dist /usr/share/nginx/html
EXPOSE 80