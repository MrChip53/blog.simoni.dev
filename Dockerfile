# Stage 1: Build CSS with Tailwind
FROM node:20-alpine AS css-builder
WORKDIR /app
COPY package.json package-lock.json* ./
RUN npm install
COPY tailwind.config.js postcss.config.js buildCss.js ./
COPY templates/ ./templates/
COPY css/ ./css/
RUN npm run build && npm run build:themes

# Stage 2: Build Go binary
FROM golang:1.26-alpine AS go-builder
WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -ldflags="-s -w" -o blog .

# Stage 3: Final image
FROM alpine:3.21
WORKDIR /app
RUN apk add --no-cache ca-certificates tzdata
COPY --from=go-builder /app/blog .
COPY --from=css-builder /app/css ./css
COPY js/ ./js/
EXPOSE 8080
CMD ["./blog"]