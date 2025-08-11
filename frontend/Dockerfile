FROM node:20-alpine AS builder
WORKDIR /app
COPY package.json package-lock.json* ./
RUN npm ci || npm install
COPY . .
# Build with Vite; VITE_API_BASE resolved at build time if provided, otherwise defaults in code
RUN npm run build

FROM nginx:1.27-alpine
COPY ./nginx.conf /etc/nginx/conf.d/default.conf
COPY --from=builder /app/dist /usr/share/nginx/html
EXPOSE 80
CMD ["nginx", "-g", "daemon off;"]


