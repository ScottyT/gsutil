FROM golang:1.17-buster as builder

# Create and change to the app directory.
WORKDIR /app

# Retrieve application dependencies.
# This allows the container build to reuse cached dependencies.
# Expecting to copy go.mod and if present go.sum.
COPY go.* ./
RUN go mod download

# Copy local code to the container image.
COPY . ./
COPY ./env /app/

# Build the binary.
RUN go build -mod=readonly -v -o server

EXPOSE 8081
ENV HOST=0.0.0.0
ENV PORT=8081

# Use a gcloud image based on debian:buster-slim for a lean production container.
# https://docs.docker.com/develop/develop-images/multistage-build/#use-multi-stage-builds
FROM gcr.io/google.com/cloudsdktool/cloud-sdk:slim

WORKDIR /app

# Copy the binary to the production image from the builder stage.
COPY --from=builder /app/server /app/server
COPY *.sh /app/
RUN chmod +x /app/*.sh

# Run the web service on container startup.
CMD ["/app/server"]
