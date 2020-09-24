.PHONY: compile
compile:
	protoc -I service/rpc/ service/rpc/*.proto --go_out=plugins=grpc:service/rpc/ --go_opt=paths=source_relative
