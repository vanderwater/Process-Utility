package main

import (
	"fmt"
	"os"
	"os/exec"
	"time"
//	"flag"		todo: implement -h
	"strconv"
	"strings"
	"github.com/golang/protobuf/proto"
	"processUtility"
)


// This currently organizes all relevant information into a map

// todo: instead of putting into map, put directly into protobuf thing. Need to research protobuf more

func GetProcessInfo() ([]byte, error){
	
	processSet := new(processUtility.ProcessSet)	
	
	// Call ps for process info, -e gives all, -o gives specified output
	ps := exec.Command("ps", "-e", "-o pid,vsz,time,comm") 
	// todo: error checking
	output, _ := ps.Output()
	
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
		fmt.Errorf("Invalid number of arguments")
		return
	}
	
	waitTime, _ := strconv.Atoi(os.Args[1])
	
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
	// Currently just runs 2 iterations
	time.Sleep(time.Second * 11)
	clock.Stop()
	fmt.Println("Done")

}
