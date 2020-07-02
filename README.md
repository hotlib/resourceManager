# Resource manager

## Development

Prerequisites:
* go

### Wire

Install:
```
go get github.com/google/wire/cmd/wire
```

Generate wiring code:
```
wire ./graph/...
```

### Grpc

Install tools:
```
go get -u github.com/golang/protobuf/{proto,protoc-gen-go}
```

Generate grpc schema:
```
go generate ./graph/graphgrpc/schema/
```

### graphqlgen

Generate graphql resolvers:
```
go generate ./graph/graphql/...
```