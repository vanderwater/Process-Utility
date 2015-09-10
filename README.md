# Process Utility

Records statistics from currently running processes, storing them as Protobufs in binary files (currently .txt files)

Running main.go records all process information as a protobuf in changes.txt and events.txt

Running with the -unmarshal flag decodes the information in changes.txt and events.txt

If necessary, .pb.go files can be remade with protoc --go_out=. process.proto

Program runs until interrupted by user

TODO: Refactor unmarshal and related functions
      Add Testing
      Change unmarshal output to be similar to ps
