// Use gofmt to clean code structure. For vim I use fatih/vim-go.
package main

import (
	"fmt"
	"os"
	"os/exec"
	"time"
	// Yes I was going to suggest using "flag".
//	"flag"		todo: implement -h
	"strconv"
	"strings"
	"github.com/golang/protobuf/proto"
	"processUtility"
)


// This currently organizes all relevant information into a map

// todo: instead of putting into map, put directly into protobuf thing. Need to research protobuf more

// This function could accept a updatePreiod time.Duration and return a (<-chan processUtility.Process)
// Doing so would alow the output to be processed concurrently somewhere else in the code. The go
// routine could then be moved into the function.
func GetProcessInfo() ([]byte, error){
	
	processSet := new(processUtility.ProcessSet)	
	
	// Move this into its own call.
	// Call ps for process info, -e gives all, -o gives specified output
	ps := exec.Command("ps", "-e", "-o pid,vsz,time,comm") 
	// todo: error checking
	output, _ := ps.Output()
	
	// Another way to write this that would handle errors could be to have
	// a function:
	func tryParse(outputLine string) (processUtility.Process, bool)
	// Then all your error handling would be internal to that function and
	// the outside loop could be something like:
	for _, outputLine := range output {
		if processInfo, valid := tryParse(line); valid {
			out <- processInfo
		}
	}
	// Loop over all the outputs
	// todo: think of a better name for s, error checking for atoi
	for i, s := range strings.Split(string(output), "\n") {
		if i == 0 { continue }  // skip line 1 if unskipped, gets headers
		if len(s) == 0 { continue } // skip last line, if unskipped causes index oob error
		currentProcess := new(processUtility.Process)
		ps_line := strings.Fields(s)
		//fmt.Println(ps_line)  For testing
		pid, _ := strconv.Atoi(ps_line[0])
		currentProcess.Pid = proto.Int32(int32(pid))
		vsz, _ := strconv.Atoi(ps_line[1])
		currentProcess.Vsz = proto.Int32(int32(vsz))
		currentProcess.Time = &ps_line[2]
		currentProcess.Comm = &ps_line[3]
		processSet.Processes = append(processSet.Processes, currentProcess)
	}

	//fmt.Println(processSet) For testing
	return proto.Marshal(processSet)
}


func main(){
	if len(os.Args) != 2 {
		// This does not error. Maybe use panic or replace with "flag".
		fmt.Errorf("Invalid number of arguments")
		return
	}
	
	waitTime, _ := strconv.Atoi(os.Args[1])
	
	// Use time.Ticker() instead. No need to Stop().
	clock := time.NewTicker(time.Duration(waitTime) * time.Second)
	
	fileo, _ := os.Create("output.txt")	// todo: implement err checking
	
	defer fileo.Close()
	
	go func() {
		for now := range clock.C {
			processInfo, _ := GetProcessInfo()
			fileo.Write(processInfo)
			fmt.Println("Logged at", now)
		}
	}()
	// Run until killed with Ctrl-C
	// Currently just runs 2 iterations
	time.Sleep(time.Second * 11)
	clock.Stop()
	fmt.Println("Done")

}
