.PHONY: proto test clean

PROTO_DIR := protos
GEN_DIR := protos/gen

proto:
	@mkdir -p $(GEN_DIR)
	protoc \
		--proto_path=$(PROTO_DIR) \
		--go_out=$(GEN_DIR) \
		--go_opt=paths=source_relative \
		--go-grpc_out=$(GEN_DIR) \
		--go-grpc_opt=paths=source_relative \
		$(PROTO_DIR)/v1/*.proto

test:
	go test ./...

clean:
	@rm -rf $(GEN_DIR)