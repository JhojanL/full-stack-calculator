# syntax=docker/dockerfile:1

# Keep Node and pnpm aligned with ../mise.toml.
FROM node:24.19.0-bookworm-slim AS build
WORKDIR /app
RUN npm install --global pnpm@12.5.1

COPY package.json pnpm-lock.yaml ./
RUN pnpm install --frozen-lockfile
COPY . .
RUN pnpm build

FROM nginx:1.30.5-alpine AS runtime
COPY nginx.conf /etc/nginx/nginx.conf
COPY --from=build /app/dist /usr/share/nginx/html
USER nginx
EXPOSE 8080
# The complete configuration is baked in; no entrypoint templating is needed.
ENTRYPOINT ["nginx"]
CMD ["-g", "daemon off;"]
