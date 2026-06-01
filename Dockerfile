FROM golang:1.26.0
LABEL authors="layden"

# Set the working directory
WORKDIR /app

# Copy go.mod and go.sum files first to cache dependencies
COPY go.mod go.sum ./

# Download and install module dependencies
RUN go mod download

# Copy the source code into the container
COPY . .

# Install system dependencies
RUN apt-get update && apt-get install -y \
    chromium \
    chromium-driver \
    && apt-get clean

# Build the application
RUN go build -o main .

# Set the entry point for the application
ENTRYPOINT ["./main"]
