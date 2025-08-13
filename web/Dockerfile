FROM node:20-alpine AS builder
WORKDIR /web

# ✅ Declare the build argument so Docker can receive it
ARG VITE_BASE_URL

# ✅ Set it as an environment variable for the rest of the build stage
ENV VITE_BASE_URL=$VITE_BASE_URL

COPY package*.json ./
RUN npm install
COPY . .

# The `npm run build` command will now correctly find and use VITE_BASE_URL
RUN npm run build

# --- final stage ---
FROM nginx:alpine
COPY --from=builder /web/dist /usr/share/nginx/html
COPY nginx.conf /etc/nginx/conf.d/default.conf
EXPOSE 80