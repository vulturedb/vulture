.PHONY: compile
compile: ## Compile the proto file.
	protoc service/rpc/*.proto --go_out=plugins=grpc:. --go_opt=paths=source_relative
