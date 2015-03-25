package main

import(
	"fmt"
	"os/exec"
	"strconv"
	"strings"
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
	process_map := make(map[int]process);

	// Loop over all the outputs
	// todo: think of a better name for s, error checking for atoi
	var current_process process
	for i, s := range strings.Split(string(output), "\n") {
		if i == 0 { continue }  // skip line 1 if unskipped, gets headers
		if len(s) == 0 { continue } // skip last line, if unskipped causes index oob error
		ps_line := strings.Fields(s)
		pid, _ := strconv.Atoi(ps_line[0])
		current_process.pid = pid
		current_process.vsz, _ = strconv.Atoi(ps_line[1])
		current_process.uptime = ps_line[2]
		current_process.comm = ps_line[3]
		process_map[pid] = current_process
	}
	for k := range process_map {
		fmt.Printf("Process: %d  VSZ:%d  Time: %s Comm: %s \n", process_map[k].pid,  process_map[k].vsz, process_map[k].uptime, process_map[k].comm)
		}
}
