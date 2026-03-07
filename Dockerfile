# codlocker-assets/Dockerfile
FROM golang:1.24-alpine AS build
WORKDIR /app
COPY go.mod ./
RUN go mod download
COPY . .
RUN go mod tidy
RUN CGO_ENABLED=0 go build -o app main.go

FROM gcr.io/distroless/base-debian12
WORKDIR /app
COPY --from=build /app/app .
COPY assets/ /app/assets/
EXPOSE 8080
USER nonroot:nonroot
ENTRYPOINT ["./app"]

# === add these below ===
# values passed from Jenkins/Kaniko
ARG ARTIFACT_ID
ARG VERSION
ARG COMMIT_SHA

# Unify + OCI labels
LABEL com.cloudbees.unify.artifact_id="${ARTIFACT_ID}" \
      com.cloudbees.unify.version="${VERSION}" \
      com.cloudbees.unify.commit_sha="${COMMIT_SHA}" \
      org.opencontainers.image.title="${ARTIFACT_ID}" \
      org.opencontainers.image.version="${VERSION}" \
      org.opencontainers.image.revision="${COMMIT_SHA}"
