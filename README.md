# Gossip

- Thua Duc Nguyen
- Duc Trung Nguyen

## Init proto

Inside internal folder:

```bash
protoc
--go_out=gossip \
--go_opt=paths=source_relative \
--go-grpc_out=gossip \
--go-grpc_opt=paths=source_relative \
gossip.proto 
```

## Usage
<https://github.com/grpc/grpc-go/blob/master/Documentation/server-reflection-tutorial.md>

### Run server
```bash
go run cmd/gossipServer.go 
```

### List and describe method
```bash
grpcurl -plaintext localhost:9001 list 

grpcurl -plaintext localhost:9001 describe gossip.GossipService.Announce                                                      
```

uint32 size = 1;
uint32 gossip_announce = 2;
uint32 ttl = 3;
uint32 reserved = 4;
uint32 data_type = 5;
bytes data = 6;

### Run a method
```bash
grpcurl -plaintext \
-d '{"size": 1, "gossip_announce": 2, "ttl": 3, "reserved": 4, "data_type": 5, "data": "data"}' \
localhost:9001 gossip.GossipService.Announce
```

## Go Directories

source: https://github.com/golang-standards/project-layout/blob/master/README.md?plain=1

### `/cmd`

Main applications for this project.

The directory name for each application should match the name of the executable you want to have (e.g., `/cmd/myapp`).

Don't put a lot of code in the application directory. If you think the code can be imported and used in other projects, then it should live in the `/pkg` directory. If the code is not reusable or if you don't want others to reuse it, put that code in the `/internal` directory. You'll be surprised what others will do, so be explicit about your intentions!

It's common to have a small `main` function that imports and invokes the code from the `/internal` and `/pkg` directories and nothing else.

### `/internal`

Private application and library code. This is the code you don't want others importing in their applications or libraries. Note that this layout pattern is enforced by the Go compiler itself. See the Go 1.4 [`release notes`](https://golang.org/doc/go1.4#internalpackages) for more details. Note that you are not limited to the top level `internal` directory. You can have more than one `internal` directory at any level of your project tree.

You can optionally add a bit of extra structure to your internal packages to separate your shared and non-shared internal code. It's not required (especially for smaller projects), but it's nice to have visual clues showing the intended package use. Your actual application code can go in the `/internal/app` directory (e.g., `/internal/app/myapp`) and the code shared by those apps in the `/internal/pkg` directory (e.g., `/internal/pkg/myprivlib`).

### `/api`

OpenAPI/Swagger specs, JSON schema files, protocol definition files.

### `/configs`

Configuration file templates or default configs.

Put your `confd` or `consul-template` template files here.

### `/test`

Additional external test apps and test data. Feel free to structure the `/test` directory anyway you want. For bigger projects it makes sense to have a data subdirectory. For example, you can have `/test/data` or `/test/testdata` if you need Go to ignore what's in that directory. Note that Go will also ignore directories or files that begin with "." or "_", so you have more flexibility in terms of how you name your test data directory.