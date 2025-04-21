BIN=thumb_maker
ENTRY_POINT=main.go

clean:
	@rm -rf bin/$(BIN)

build: clean
	@go build -o bin/$(BIN) $(ENTRY_POINT)

export: build
	@echo "Adding $(BIN) to temp PATH..."
	@export PATH="$(shell pwd)/bin:$(PATH)"

run:
	@go run $(ENTRY_POINT)

.PHONY: go clean build export run
