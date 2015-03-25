package main

import(
	"fmt"
	"os/exec"
	"strconv"
	"strings"
	"process.proto"
	"github.com/golang/protobuff/proto"
)

type process struct {
	pid, vsz int
	uptime, comm string	
}

// This currently organizes all relevant information into a map

// todo: instead of putting into map, put directly into protobuf thing. Need to research protobuf more

func main(){
	// Call ps for process info, -e gives all, -o gives specified output
	ps := exec.Command("ps", "-e", "-o pid,vsz,time,comm") 
	// todo: error checking
	output, _ := ps.Output()
	processMap := make(map[int]process);
	
	// Loop over all the outputs
	// todo: think of a better name for s, error checking for atoi
	var currentProcess process
	for i, s := range strings.Split(string(output), "\n") {
		if i == 0 { continue }  // skip line 1 if unskipped, gets headers
		if len(s) == 0 { continue } // skip last line, if unskipped causes index oob error
		ps_line := strings.Fields(s)
		pid, _ := strconv.Atoi(ps_line[0])
		currentProcess.pid = pid
		currentProcess.vsz, _ = strconv.Atoi(ps_line[1])
		currentProcess.uptime = ps_line[2]
		currentProcess.comm = ps_line[3]
		processMap[pid] = currentProcess
	}
	for k := range processMap {
		fmt.Printf("Process: %d  VSZ:%d  Time: %s Comm: %s \n", processMap[k].pid,  processMap[k].vsz, processMap[k].uptime, processMap[k].comm)
	}	
}
