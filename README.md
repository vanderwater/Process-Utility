# Process Utility

Records statistics from all processes, placing them in a textfile according to the .proto

Place process.proto in GOPATH and use protoc --go_out=. process.proto

Program currently runs for 11 seconds writing to output.txt then terminates
Almost no error checking is done yet
