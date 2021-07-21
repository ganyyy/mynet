PROTOC = protoc
PROTO_FILE_PATH = ./proto/define
PB_FILE_PATH = ../


.PHONY: proto

proto:
	@$(PROTOC) -I $(PROTO_FILE_PATH) --go_out=plugins=grpc:$(PB_FILE_PATH) $(PROTO_FILE_PATH)/*.proto
