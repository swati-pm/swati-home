FROM golang:1.23-alpine AS build

WORKDIR /app

# Copy and download deps first (cache layer)
COPY server/go.mod server/go.sum* ./
RUN go mod download || true

# Copy server source
COPY server/*.go ./

# Download deps now that all source exists
RUN go mod tidy && go mod download

# Build
RUN CGO_ENABLED=0 go build -o /server .

# --- Runtime ---
FROM alpine:3.20
COPY --from=build /server /server
COPY server/data/ /app/data/
COPY server/templates/ /app/templates/
EXPOSE 8080
CMD ["/server"]
