# WitWiz Server

## How to generate go files from proto

```
protoc --go_out=. --go_opt=paths=source_relative --go-grpc_out=. --go-grpc_opt=paths=source_relative ./proto/witwiz.proto
```