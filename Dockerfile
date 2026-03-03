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

# Install all system dependencies (Python, Chromium, and libzbar)
RUN apt-get update && apt-get install -y \
    python3 \
    python3-pip \
    libzbar0 \
    libgl1 \
    chromium \
    chromium-driver \
    && apt-get clean

# Install Python dependencies globally (no venv needed in container)
RUN pip3 install --break-system-packages -r requirements.txt

# Expose the app port (if necessary)
EXPOSE 8080

# Build the application
RUN go build -o main .

# Set the entry point for the application
ENTRYPOINT ["./main"]
