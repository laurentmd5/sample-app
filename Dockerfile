# ==============================
# 🔨 STAGE 1 — Build
# ==============================
FROM golang:1.22-alpine AS builder

# Labels pour traçabilité
LABEL maintainer="Laurent MAVOUNGOU <dev@yourdomain.com>"
LABEL description="Go Dev Dashboard - outil de monitoring et sécurité pour développeurs freelances"

WORKDIR /app

# Copie des fichiers sources et initialisation du module
COPY go.mod go.sum* . 2>/dev/null || true
RUN go mod init go-dev-dashboard || true
RUN go mod tidy

COPY . .

# Compilation statique (sécurisée et portable)
RUN CGO_ENABLED=0 GOOS=linux go build -a -o /go-dev-dashboard .

# ==============================
# 🧊 STAGE 2 — Final image
# ==============================
FROM gcr.io/distroless/base-debian11

WORKDIR /

# Copie du binaire depuis le builder
COPY --from=builder /go-dev-dashboard /go-dev-dashboard

# Port d’écoute de l’application
ENV PORT=8090

# Utilisateur non-root pour sécurité
USER nonroot:nonroot

# Exposition du port (facilite debug Docker)
EXPOSE 8090

# Commande de démarrage
ENTRYPOINT ["/go-dev-dashboard"]
