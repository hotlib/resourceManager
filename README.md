# Resource manager

## TODOs

* !! graphql.schema - expand and support CRUD for resoruceType, pools, resources
* graphql.schema - add support for directives to better control entity field values
* !! pool - support custom pool type with WASM (on graphql and ent)
* !! pool - support labels
* ent - research and implement policies/privacy for entities based on user role (RBAC)
* docker-compose - create docker compose for graphql server + mysql
* security ?
* logging
* actions triggers ?

## Glossary

* Telemetry - tracing data streaming
* Policy - RBAC control over entity CRUD operations. Defined as part of ent.go schema
* Hook - Custom code invoked when interacting ent.go entity. Not used in RM
* Features -
* Directive - Custom extension to graphql schema for graphqlgen framework. Needs to be defined in the schema and implemented as go code. Example: IntRange restriction directive.
* Actions -
* Triggers -

## Features
List of important features of resource manager

### Building on FBC inventory
Lots of components and parts of the DB schema are reused from the inventory project.

### Model driven DB
Database schema is derived/generated from ent.go schema definition. Ent.go hides/handles all DB interactions. Ent schema can be found at ent/schema

### Model driven graphql server
GraphQL server is derived/generated from graphql.schema. Code which ties graphql and ent together is written manually.
Schema needs to be kept in sync with ent.go DB schema, they are not connected in any automated way.

### APIs
Northbound APIs:

#### HTTP
Exposes grahpql API

#### gRPC
Exposes tenant control

#### webSockets
???

### Multitenancy
Multitenancy is supported throughout the stack.
In DB, each tenant has their own database.
GraphQL server switches to appropriate tenant context using TenantHandler baked into the HTTP API.

### RBAC
???
Policy

### Logging
???

### Telementry
Support for tracing (distributed tracing). Streams data into a collector such as Jaeger.
Default is Nop.
See main parameters or telementry/config.go for further details to enable jaeger tracing

### Health
Basic health info of the app (also checks if mysql connection is healthy)

```
# server can serve requests
http://localhost:8884/healthz/liveness
# server works fine
http://localhost:8884/healthz/readiness 
```

#### Metrics
Prometheus style metrics are exposed at:

```
http://localhost:8884/metrics
```

### Security
???

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

### ent.go

Generate ent.go entities:
```
go generate ./eng/...
```

### graphqlgen

Generate graphql resolvers:
```
go generate ./graph/graphql/...
```