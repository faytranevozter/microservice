FROM node:20-alpine AS builder
WORKDIR /app
COPY package.json package-lock.json* ./
RUN npm ci || npm install
COPY . .
# If an app.env file exists in the frontend folder, copy it to .env.production
# so Vite/dotenv picks up variables at build time. This allows injecting
# VITE_... variables when building the image (e.g. VITE_API_BASE).
RUN if [ -f ./app.env ]; then cp ./app.env .env.production; fi

# Build with Vite; VITE_API_BASE (and other VITE_* vars) will be resolved at build time
RUN npm run build

FROM nginx:1.27-alpine
COPY ./nginx.conf /etc/nginx/conf.d/default.conf
COPY --from=builder /app/dist /usr/share/nginx/html
EXPOSE 80
CMD ["nginx", "-g", "daemon off;"]


