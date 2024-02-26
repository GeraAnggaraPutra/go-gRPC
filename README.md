# GO Unary gRPC

[Install Proto!](https://github.com/protocolbuffers/protobuf/releases)
[Documentation](https://grpc.io/docs/languages/go/quickstart/)

## Compile

```bash
protoc --proto_path=protos protos/*.proto --go_out=./ --go-grpc_out=./
```

## run

```bash
cp .env.example .env
```

```bash
go mod tidy
```

```bash
go run cmd/main.go
```
