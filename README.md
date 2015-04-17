# Process Utility

Records statistics from currently running processes, storing them as Protobufs in binary files (currently .txt files)

To use, first run main.go INT with INT being however many seconds to wait between collecting process info

Next run extract.go and all process info will be displayed in unmarshal.txt

If necessary, .pb.go files can be remade with protoc --go_out=. process.proto

Program runs until interrupted by user


