# general make targets
all: test

test:
	@go test -v ./...

protobuf:
	@protoc --gogofaster_out=. network/transport/*.proto 
	@protoc --gogofaster_out=. frameworks/ethpool/protobuf/*.proto 
	