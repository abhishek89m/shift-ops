FROM golang:1.26.1-alpine AS build

WORKDIR /src

COPY services/api ./services/api

WORKDIR /src/services/api

RUN go build -o /out/shift-ops-api .

FROM alpine:3.20

WORKDIR /app

COPY --from=build /out/shift-ops-api ./shift-ops-api

EXPOSE 8080

CMD ["./shift-ops-api"]

