set protoc=%GoPath%\bin\protoc.exe
set protoc-gen-go=%GoPath%\bin\protoc-gen-go.exe

%protoc% --plugin=protoc-gen-go=%protoc-gen-go% --go_out .. --proto_path . option.proto -I=.;google/protobuf=.