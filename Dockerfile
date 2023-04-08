FROM golang:alpine as base

FROM base as builder

# Set destination for COPY
WORKDIR /app

# Download Go modules
COPY go.mod go.sum ./
RUN go mod download

# Copy the source code. Note the slash at the end, as explained in
# https://docs.docker.com/engine/reference/builder/#copy
COPY . ./

# Build
RUN CGO_ENABLED=0 GOOS=linux go build -o /docker-gs-ping

# https://docs.docker.com/language/golang/build-images/#tag-images
# if it is neccesary maybe use a debian base image, missin go test also
FROM scratch
COPY --from=builder /docker-gs-ping /docker-gs-ping

EXPOSE 8080

ENTRYPOINT ["/docker-gs-ping"]


