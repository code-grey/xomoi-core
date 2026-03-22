# Xomoi-Core Makefile

.PHONY: all proto-go proto-sdk build clean

# Tools
PROTOC = protoc
NANOPB_GEN = python -m nanopb_generator

# Paths
PROTO_DIR = proto/v1
INTERNAL_DIR = internal/core/proto
SDK_DIR = sdk/proto

all: proto-go proto-sdk build

# Generate Go Protobuf code
proto-go:
	mkdir -p $(INTERNAL_DIR)
	$(PROTOC) --proto_path=$(PROTO_DIR) --go_out=$(INTERNAL_DIR) --go_opt=paths=source_relative $(PROTO_DIR)/*.proto

# Generate NanoPB C code for the SDK
proto-sdk:
	mkdir -p $(SDK_DIR)
	# Assumes nanopb is in the python path or provided via the sdk directory
	$(PROTOC) --proto_path=$(PROTO_DIR) --nanopb_out=$(SDK_DIR) $(PROTO_DIR)/*.proto

# Build the Go binary
build:
	go build -o build/xomoi cmd/xomoi/main.go

clean:
	rm -rf $(INTERNAL_DIR)/*.go
	rm -rf $(SDK_DIR)/*.h $(SDK_DIR)/*.c
	rm -f build/xomoi
