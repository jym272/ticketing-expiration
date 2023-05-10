FROM cosmtrek/air:v1.43.0 as base

FROM base as builder

# Set destination for COPY
WORKDIR /app

# Download Go modules
COPY go.mod go.sum ./
RUN go mod download

# Copy the source code. Note the slash at the end, as explained in
# https://docs.docker.com/engine/reference/builder/#copy
# go files
COPY ./async ./async
COPY ./callbacks ./callbacks
COPY ./nats ./nats
COPY ./echo.go ./
COPY ./main.go ./

# .ait.toml [build] cmd
# CGO_ENABLED=0 -> it will force the rebuild with air CMD
RUN GOOS=linux go build -buildvcs=false -o ./tmp/main .

# config
COPY ./.air.toml ./

CMD ["air", "-c", ".air.toml"]

