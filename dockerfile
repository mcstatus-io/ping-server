# Build stage
FROM golang:1.18 AS build

WORKDIR /app

# Install required dependencies
COPY go.mod go.sum ./
RUN go mod download

# Copy the source code
COPY . .

# Build the executable with CGO disabled
RUN CGO_ENABLED=0 go build -o bin/main src/*.go
RUN mv config.example.yml ./bin/config.yml


#######################
# Test stage
FROM build as test
RUN go test -v ./... 


#######################
# Runtime stage

FROM gcr.io/distroless/base

# Set the working directory
WORKDIR /app

# Copy the binary from the build stage
COPY --from=build /app/bin/main  /app/main
COPY --from=build /app/bin/config.yml  /app/config.yml

# Expose the port the app runs on (replace with your desired port)
EXPOSE 3001

# Run the app
CMD ["./main"]
