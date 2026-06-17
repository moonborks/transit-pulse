# Proto Files

The proto files included in this project were retrieved from the following locations:

- [gtfs-realtime.proto](https://gtfs.org/documentation/realtime/proto/)
- [gtfs-realtime-NYCT.proto](https://www.mta.info/developers)

## Commands

Requirements:
- [protoc](https://github.com/protocolbuffers/protobuf/releases)
- [protoc-gen-go](https://pkg.go.dev/google.golang.org/protobuf/cmd/protoc-gen-go)

```
go install google.golang.org/protobuf/cmd/protoc-gen-go@latest
```

Afterwards, you can compile the `.proto` files into `.pb.go` files.

```
protoc --plugin=protoc-gen-go=<path_to_protoc-gen-go> \
--go_out=. \
--go_opt=paths=source_relative \
-I github.com/moonborks/transit-pulse/internal/transit/mta/gtfs/proto \
github.com/moonborks/transit-pulse/internal/transit/mta/gtfs/proto/<protofile_1.proto> \
github.com/moonborks/transit-pulse/internal/transit/mta/gtfs/proto/<protofile_2.proto>
```


## Additional Information

More information about the structure of the proto files can be found here:

- [api.mta.info/GTFS.pdf](https://api.mta.info/GTFS.pdf)
- [github.com/google/transit](https://github.com/google/transit/blob/master/gtfs-realtime/proto/gtfs-realtime.proto)
- [gtfs.org/documentation/realtime/proto](https://gtfs.org/documentation/realtime/proto/)
