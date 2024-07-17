DWAL - Distributed Write-Ahead Log Service
==========================================

A distributed write-ahead log (aka commit or transaction log) is a system designed to handle the storage, replication, and dissemination of records (log entries) across multiple servers or nodes within a network.

This is a simple implementation of distributed WAL server using GO, protobuffs and gRPC.

For example, Kafka can serve as a kind of external transaction log for a distributed system. The log helps replicate data between nodes and acts as a re-syncing mechanism for failed nodes to restore their data.

It's a fundamental component in distributed systems and databases, ensuring high availability, durability, and fault tolerance.

# Core Principles

* Immutability: Entries in the WAL are immutable once written. This simplifies replication, data recovery, and consistency checks, as logs can be replayed to rebuild state.
* Append-only: New records are always appended at the end of the log. This ensures efficient writes and straightforward replication across nodes.
* Sequential Access: Logs are designed for sequential access, making reads and writes efficient, especially for use cases that naturally fit a time-ordered sequence.

# Functionality

* Replication: The log is replicated across several nodes to ensure data availability even in the face of hardware failures or network partitions.

* Consistency: It helps in achieving consistency across distributed systems through consensus algorithms (like Raft or Paxos) that ensure all copies of the log agree on the current state.

# Implementation notes:
* directory structure: https://gist.github.com/ayoubzulfiqar/9f1a34049332711fddd4d4b2bfd46096
* missing modules: `go get github.com/bla/foomodule`

## gRPC
* follow installation instructions here: https://grpc.io/docs/languages/go/quickstart/

## Protobuf
* Install protobuf runtime, and compile protobuf with:
```
go get google.golang.org/protobuf/
protoc api/v1/*.proto --go_out=. --go_opt=paths=source_relative --proto_path=.
```


# Testing
To test the simple JSON WAL server, launch it: 
```
go run cmd/server/main.go
```

In other tab do POST and GET requests:
```
curl -X POST localhost:8080 -d '{"record": {"value": "BLA1234567FOOBAR"}}'
curl -X POST localhost:8080 -d '{"record": {"value": "FOO1234567FOOBAR"}}'
curl -X POST localhost:8080 -d '{"record": {"value": "XXX1234567FOOBAR"}}'

curl -X GET localhost:8080 -d '{"offset": 0}'
curl -X GET localhost:8080 -d '{"offset": 1}'
curl -X GET localhost:8080 -d '{"offset": 2}'
```