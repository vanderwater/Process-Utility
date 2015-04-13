package main

import (
	"fmt"
	"math"
	"os"
	"os/exec"
	"time"
	//	"flag"		todo: implement -h
	"github.com/golang/protobuf/proto"
	"processUtility"
	"strconv"
	"strings"
)

/* To-Do:
*	Rearrange functions
*	Create a vector(maybe map) of processes and processinfo struct
*	Compare slice with incoming processes
*	Only marshal when necessary
*
 */

type ProcessInfo struct {
	ProcessID, VirtualSize int32
	Command, TimeStarted   string
	CPUUsage               float64
}

//Calls ps and returns the output
func GetCurrentProcesses() []byte {

	ps := exec.Command("ps", "-e", "-o pid,vsz,time,comm,pcpu")
	output, err := ps.Output()
	if err != nil {
		panic(err)
	}
	return output
}

// Helper Function for SendProcessInfo
// Transforms ProcessInfo struct into Protobuf Process
func PBProcess(src ProcessInfo) *processUtility.Process {

	result := new(processUtility.Process)
	result.ProcessID = proto.Int32(src.ProcessID)
	result.VirtualSize = proto.Int32(src.VirtualSize)
	// Is it possible that addresses won't be cleaned up due to this?
	result.TimeStarted = &src.TimeStarted
	result.Command = &src.Command

	return result
}

func Int32PercentDifference(first int32, second int32) float64 {
	var result float64
	result = math.Abs(float64(first - second))
	result = result / (float64(first+second) / 2)
	result = result * 100
	return result
}

func Float64PercentDifference(first float64, second float64) float64 {
	var result float64
	result = math.Abs(first - second)
	result = result / ((first + second) / 2)
	result = result * 100
	return result
}

// Compares two Processes and return true if they are different enough
func HasProcessChanged(proc1 ProcessInfo, proc2 ProcessInfo) bool {

	// Check CPU

	// Check Virtual Size
	if Int32PercentDifference(proc1.VirtualSize, proc2.VirtualSize) >= 10 {
		return true
	}
	return false
}

// Marshals started and terminated process Info
func MarshalEventInfo(start []ProcessInfo, finished []ProcessInfo) ([]byte, error) {

	processSet := new(processUtility.ProcessSet)

	for _, curr := range start {
		currentProcess := PBProcess(curr)
		currentProcess.WasOpened = proto.Bool(true)
		processSet.Processes = append(processSet.Processes, currentProcess)
	}

	for _, curr := range finished {
		currentProcess := PBProcess(curr)
		currentProcess.WasClosed = proto.Bool(true)
		processSet.Processes = append(processSet.Processes, currentProcess)
	}

	return proto.Marshal(processSet)

}

// Marshals updated process info
func MarshalUpdateInfo(updates []ProcessInfo) ([]byte, error) {

	processSet := new(processUtility.ProcessSet)

	for _, curr := range updates {
		currentProcess := PBProcess(curr)
		processSet.Processes = append(processSet.Processes, currentProcess)
	}

	return proto.Marshal(processSet)
}

func TryParse(outputLine string) (ProcessInfo, bool) {

	outputFields := strings.Fields(outputLine)

	var result ProcessInfo
	nilProcess := ProcessInfo{ProcessID: 0, VirtualSize: 0, TimeStarted: "", Command: ""}
	// checks for first iteration and last iteration of outputFields
	// First iteration is top line of ps created by output
	if len(outputFields) != 4 {
		// To-do: Getting error on returning nil, figure out what to return
		// instead of nilProcess
		return nilProcess, false
	}

	// Turns ID and size strings into int before casting them into int32. Was having
	// issues doing it all in one step
	pid, err := strconv.Atoi(outputFields[0])
	if err != nil {
		return nilProcess, false
	}
	vsz, err := strconv.Atoi(outputFields[1])
	if err != nil {
		return nilProcess, false
	}
	// Puts outputline data into processinfo struct and returns it
	result.ProcessID = int32(pid)
	result.VirtualSize = int32(vsz)
	result.TimeStarted = outputFields[2]
	result.Command = outputFields[3]

	return result, true
}

func MapDifference(primary map[int32]ProcessInfo, secondary map[int32]ProcessInfo) []ProcessInfo {

	result := make([]ProcessInfo, 100)

	for id, process := range primary {
		if _, ok := secondary[id]; !ok {
			result = append(result, process)
		}
	}

	return result

}

func GetProcessInfo(updatePeriod time.Duration) /* ([]byte, error)*/ {

	oldProcessMap := make(map[int32]ProcessInfo)

	fileo, _ := os.Create("output.txt") // todo: implement err checking
	defer fileo.Close()
	filee, _ := os.Create("events.txt")
	defer filee.Close()

	clock := time.NewTicker(updatePeriod * time.Second)

	// removed go func{ }() around this for loop
	for now := range clock.C {

		processMap := make(map[int32]ProcessInfo)

		output := GetCurrentProcesses()
		outputLines := strings.Split(string(output), "\n")
		// Loop over all the outputs, check against current slice of processes
		for _, line := range outputLines {
			// To-do: Something if its invalid
			if currentProcess, valid := TryParse(line); valid {
				processMap[currentProcess.ProcessID] = currentProcess

				// We want to compare new process to old processes

			}

			// To-do: current Process against old process to see if I should update events slice
		}

		// Temporary until I add events.txt
		processesUpdated := make([]ProcessInfo, 0)

		// Gets processes started and finished
		processesStarted := MapDifference(processMap, oldProcessMap)
		processesFinished := MapDifference(oldProcessMap, processMap)

		// Marshals Processes
		events, _ := MarshalEventInfo(processesStarted, processesFinished)
		filee.Write(events)

		updates, _ := MarshalOutputInfo(processesUpdated)

		fileo.Write(updates)

		fmt.Println("Logged at", now)

		// set oldProcessMap to current
		oldProcessMap = processMap

	}
}

func main() {

	// Couldn't one line it all, what's best practice?
	updateAmount, _ := strconv.Atoi(os.Args[1]) // If an error is thrown, an invalid amount was entered
	updatePeriod := time.Duration(updateAmount)

	GetProcessInfo(updatePeriod)

}
