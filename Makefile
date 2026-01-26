.PHONY: build run-server run-web clean install

# Variables
SERVER_DIR=server
WEB_DIR=web

# Default target
all: install build

# Install dependencies
install:
	cd $(SERVER_DIR) && go mod download
	cd $(WEB_DIR) && npm install

# Build the project
build:
	cd $(SERVER_DIR) && go build -o bin/server main.go
	cd $(WEB_DIR) && npm run build

# Run the backend server
run-server:
	cd $(SERVER_DIR) && go run main.go

# Run the frontend web app
run-web:
	cd $(WEB_DIR) && npm run dev

# Run both (using & for parallel execution in background)
dev:
	make run-server & make run-web

# Clean up build artifacts and temporary files
clean:
	rm -rf $(SERVER_DIR)/bin
	rm -rf $(SERVER_DIR)/videos/*
	rm -rf $(SERVER_DIR)/temp_chunks/*
	rm -rf $(WEB_DIR)/dist
