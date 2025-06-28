# syntax=docker/dockerfile:1

# Base stage for shared configurations
FROM golang:1.23.3-alpine AS base
WORKDIR /app
RUN apk add --no-cache gcc musl-dev make

# Development stage
FROM base AS development
RUN apk add --no-cache git
COPY . .
RUN make build
# Install CompileDaemon
RUN go install github.com/githubnemo/CompileDaemon@latest

# Create build directory and set permissions
RUN mkdir -p /app/build && \
    chown -R 1000:1000 /app && \
    chmod -R 777 /app


# Production stage
FROM alpine:latest AS production
RUN apk --no-cache add make

WORKDIR /app
COPY --from=development /app/build/toy-redis .
COPY --from=development /app/build/toy-redis .
COPY --from=development /app/makefile .

ENTRYPOINT ["./toy-redis"]
