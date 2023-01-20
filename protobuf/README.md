# Benchmarks JSON vs Protobuf
```
protoc-gen-go v1.28
protoc v3.21.12
mailru/easyjson v0.7.7
golang/protobuf v1.5.2
json-iterator/go v1.1.12 (reflect-api, ConfigFastest)
```

# Useful commands

```
easyjson -no_std_marshalers protobuf/json.go

protoc \
    --go_out=./protobuf \
    --go-vtproto_out=./protobuf \
    --plugin protoc-gen-go-vtproto="$GOPATH/bin/protoc-gen-go-vtproto.exe" \
    --go-vtproto_opt=features=marshal+unmarshal+size \
    protobuf/service.proto

protoc \
    --go_out=./protobuf/search-v3 \
    --go-vtproto_out=./protobuf/search-v3 \
    --plugin protoc-gen-go-vtproto="$GOPATH/bin/protoc-gen-go-vtproto.exe" \
    --go-vtproto_opt=features=marshal+unmarshal+size \
    protobuf/search-v3/results.proto
```