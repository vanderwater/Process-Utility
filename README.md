# Process Utility

Records statistics from currently running processes, storing them as Protobufs in binary files (currently .txt files)

Running main.go records all process information into changes
Next run extract.go and all process info will be displayed in unmarshal.txt

If necessary, .pb.go files can be remade with protoc --go_out=. process.proto

Program runs until interrupted by user


