# TEAM INFORMATION:

- Thua Duc Nguyen, 03752081
- Duc Trung Nguyen, 03783118
- Team name: Gossip-7
- Implementation: Gossip Protocol


# Programming Language and Operating System:

In this project, we decide to use Go as the programming language. We also use Protocol Buffers for data serialization and gRPC to design APIs. Using Go, Protocol Buffers, and gRPC offers numerous benefits:
1. Efficiency in Development:
- GoLang: Go's simplicity, readability, and concurrency support make it ideal for developing networked applications like VoidPhone. Its compiled nature ensures fast execution and efficient resource utilization, which is critical for a P2P application where performance is paramount.

2. Protocol Buffers for Data Serialization:
- Efficient Serialization: Protocol Buffers shine in efficient data serialization. Unlike JSON's text-based approach, they leverage a compact binary format. This translates to substantial performance boosts (1)
 Benchmarks show Protocol Buffers can be up to six times faster than JSON. Furthermore, messages are 34% smaller, and delivery to JavaScript code sees a 21% speedup (2). This efficiency makes Protocol Buffers ideal for applications like Voidphone, where data transfer speed and size are critical.

3. gRPC for Communication:
- Bidirectional Streaming: gRPC supports bidirectional streaming (3), allowing efficient real-time communication between peers.
- Strong Typing and Code Generation: gRPC leverages Protocol Buffers to define service contracts, enabling strong typing and automatic code generation for client and server. This reduces the likelihood of errors and streamlines the development process.

4. Security and Reliability:
- TLS Support in gRPC: gRPC supports Transport Layer Security (TLS) encryption out of the box (3), ensuring secure communication between peers in the VoidPhone network. This helps mitigate potential security threats, safeguarding user privacy and data integrity.

(1) Karandikar, Sagar, et al. "A hardware accelerator for protocol buffers." MICRO-54: 54th Annual IEEE/ACM International Symposium on Microarchitecture. 2021.
(2) https://auth0.com/blog/beating-json-performance-with-protobuf/
(3) https://grpc.io/docs/what-is-grpc/core-concepts/


# Build System :

Go Modules [1] has been Go projects' official dependency management system since Go 1.11. Using Go Modules with some additional tooling is a great choice for building our VoidPhone project:

* Simplicity: The Go build system is already integrated into the Go toolchain, so we won't need to install or learn a separate build system. This keeps our development environment streamlined.

* Native Support: Go natively supports compiling Go source code and linking dependencies. Additionally, there are official packages for Protobuf generation (protoc-gen-go [2]) and gRPC server/client generation (grpc-gateway [3]). These integrate seamlessly with the Go build system.

* Flexibility: While the Go build system is lightweight, it allows us to extend functionality through custom build commands and integration with other tools. This is useful for tasks like generating Protobuf code from your .proto files and gRPC interfaces.

[1] https://go.dev/blog/using-go-modules
[2] https://pkg.go.dev/google.golang.org/protobuf/cmd/protoc-gen-go
[3] https://github.com/grpc-ecosystem/grpc-gateway 

# Quality Assurance measures:

Intended measures to guarantee the quality of your software
+ How do you write test cases: Unit-test for business logic, E2E test for API
+ Linter: built-in glint
+ Security scanning tool: https://securego.io/
+ Static analysis: Valgrind
+ Logging: https://github.com/uber-go/zap 	