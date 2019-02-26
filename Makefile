# general make targets
all: protobuf

protobuf:
	@protoc --gogofaster_out=. network/transport/*.proto 
	@protoc --gogofaster_out=. frameworks/ethpool/protobuf/*.proto 
	