.PHONY: build zip clean

# Variables
BINARY_NAME = main
ZIP_NAME = function.zip
MAIN_FILE = ./main/main.go
BINARY_PATH = ./main/$(BINARY_NAME)  # Path to the binary inside the main/ directory
BOOTSTRAP_NAME = bootstrap

# Build for Linux
build:
	GOOS=linux GOARCH=amd64 go build -o $(BINARY_PATH) $(MAIN_FILE)

# Package into a zip file
zip: build
	@echo "Zipping $(BINARY_PATH) and $(BOOTSTRAP_NAME) into $(ZIP_NAME)"
	zip $(ZIP_NAME) $(BINARY_PATH) $(BOOTSTRAP_NAME)
	@echo "Zipping complete, now cleaning up"
	@rm -f $(BINARY_PATH)  # Remove the binary after zipping
	@echo "Cleanup complete"

# Clean up generated files
clean:
	@echo "Cleaning up generated files..."
	@rm -f $(ZIP_NAME)   # Remove zip file if it exists
	@if [ -f $(BINARY_PATH) ]; then rm -f $(BINARY_PATH); fi   # Remove binary if it's a file
	@if [ -d $(BINARY_PATH) ]; then rm -rf $(BINARY_PATH); fi  # Remove binary if it's a directory
	@echo "Cleanup complete"
