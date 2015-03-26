package main

import (
	"fmt"
	"os"
	"time"
	"io"
//	"flag"		todo: implement -h
	"strconv"
	"strings"
	"github.com/golang/protobuff/proto"
)

type process struct {
	pid, vsz int
	uptime, comm string	
}

// This currently organizes all relevant information into a map

// todo: instead of putting into map, put directly into protobuf thing. Need to research protobuf more

func GetProcessInfo(){
	
	processSet := new(processUtility.ProcessSet)	
	
	currentProcess := new(processUtility.Process)
	
	// Call ps for process info, -e gives all, -o gives specified output
	ps := exec.Command("ps", "-e", "-o pid,vsz,time,comm") 
	// todo: error checking
	output, _ := ps.Output()
	
	// Loop over all the outputs
	// todo: think of a better name for s, error checking for atoi
	for i, s := range strings.Split(string(output), "\n") {
		if i == 0 { continue }  // skip line 1 if unskipped, gets headers
		if len(s) == 0 { continue } // skip last line, if unskipped causes index oob error
		ps_line := strings.Fields(s)
		currentProcess.pid, _ := strconv.Atoi(ps_line[0])
		currentProcess.vsz, _ = strconv.Atoi(ps_line[1])
		currentProcess.uptime = ps_line[2]
		currentProcess.comm = ps_line[3]
		processSet.processes = append(processSet.processes, currentProcess)
	}

	return proto.Marshal(processSet)
}


func main(){
	if len(os.Args) != 2 {
		fmt.Errorf("Invalid number of arguments")
		return
	}
	
	waitTime := os.Args[1]
	
	clock := time.Tick(waitTime * time.Second)
	
	fileo, _ := os.Open("output.txt")	// todo: implement err checking
	
    defer func() {
	    if err := fileo.Close(); err != nil {
			 panic(err)
		}
    }()

	for now := range clock {
		fmt.Fprint(fileo, processUtility.GetProcessInfo())
	}
}
