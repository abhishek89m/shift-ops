FROM node:22-alpine

WORKDIR /app

COPY package.json pnpm-workspace.yaml pnpm-lock.yaml ./
COPY apps/web ./apps/web

RUN corepack enable && pnpm install --frozen-lockfile
RUN pnpm --filter web build

EXPOSE 4173

CMD ["pnpm", "--filter", "web", "preview", "--host", "0.0.0.0", "--port", "4173"]

