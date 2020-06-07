package proto

//go:generate protoc -I. common/common.proto --go_out=paths=source_relative,plugins=grpc:.
//go:generate protoc -I. common/events.proto --go_out=paths=source_relative,plugins=grpc:.
//go:generate protoc -I. svc/api.proto --go_out=paths=source_relative,plugins=grpc:.
