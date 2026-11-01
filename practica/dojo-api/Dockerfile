# Imagen multi-stage del dojo-api.
#
# Muchos equipos compilan el binario FUERA (en el pipeline de CI) y la imagen
# solo lo copia. Aquí compilamos dentro, en una etapa "builder", para que veas el
# ciclo completo. La imagen final no trae Go ni el código: solo el binario.

# ---------- Etapa 1: compilar ----------
FROM golang:1.24-bookworm AS builder
WORKDIR /src

# Primero solo go.mod/go.sum: esta capa se cachea y no se vuelve a bajar todo
# cada vez que cambias un .go.
COPY go.mod go.sum ./
RUN go mod download

# swag genera docs/, que no se versiona .
RUN go install github.com/swaggo/swag/cmd/swag@v1.16.3

COPY . .
RUN swag init -g main.go \
 && CGO_ENABLED=0 go build -o /out/dojo-api .

# ---------- Etapa 2: ejecutar ----------
FROM debian:12-slim
RUN apt-get update \
 && apt-get install -y --no-install-recommends ca-certificates \
 && rm -rf /var/lib/apt/lists/*
WORKDIR /app
COPY --from=builder /out/dojo-api /app/dojo-api

# La imagen NO trae config.json: lo monta quien la corre
# (docker-compose) y se indica con la variable CONFIG.
ENV CONFIG=/app/config.json
EXPOSE 1099
CMD ["/app/dojo-api"]
