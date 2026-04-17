FROM node:22-alpine

WORKDIR /app

COPY package.json pnpm-workspace.yaml pnpm-lock.yaml ./
COPY playground ./playground
COPY services/api/schema.sql ./services/api/schema.sql

ENV VITE_API_BASE_URL=http://127.0.0.1:8080
ENV VITE_APP_BASE_URL=http://127.0.0.1:4174/

RUN corepack enable && pnpm install --frozen-lockfile
RUN pnpm --filter playground build

EXPOSE 4175

CMD ["pnpm", "--filter", "playground", "preview", "--host", "0.0.0.0", "--port", "4175"]
