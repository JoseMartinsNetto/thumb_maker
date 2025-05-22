BIN=thmk
ENTRY_POINT=cmd/thumb_maker/*.go

clean:
	@rm -rf bin/**

build: clean
	@go build -o bin/$(BIN) $(ENTRY_POINT)

export: build
	@echo "Adding $(BIN) to temp PATH..."
	@export PATH="$(shell pwd)/bin:$(PATH)"

run:
	@go run $(ENTRY_POINT)

.PHONY: go clean build export run
