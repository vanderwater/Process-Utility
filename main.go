package main

import (
	"fmt"
	"github.com/golang/protobuf/proto"
	"math"
	"os"
	"os/exec"
	"processUtility"
	"strconv"
	"strings"
	"time"
	//	"flag"		todo: implement -h
)

/* To-Do:
*	Something is wrong with the encoding of or decoding of TimeStarted
 */

// In order to add a new field to Protobuf
// Add field to Process Info, update TryParse and PBProcess

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
	result.TimeStarted = &src.TimeStarted
	result.Command = &src.Command
	result.CPUUsage = proto.Float64(src.CPUUsage)

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

	// Check CPU greater than 50 percent
	if Float64PercentDifference(proc1.CPUUsage, proc2.CPUUsage) >= 50 {
		return true
	}

	// Check Virtual Size greater than 10 percent
	if Int32PercentDifference(proc1.VirtualSize, proc2.VirtualSize) >= 10 {
		return true
	}
	return false
}

// It could be possible to condense these two Marshal functions by passing in an integer
// to the function. 0 for updated, 1 for started, 2 for exited. Then only one function would
// need to be called. Concatenating two slices would have to happen in the GetProcessInfo function then

// Marshals started and terminated process Info
func MarshalEventInfo(start []ProcessInfo, finished []ProcessInfo) []byte {

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

	result, err := proto.Marshal(processSet)
	if err != nil {
		panic(err)
	}
	return result
}

// Marshals updated process info
func MarshalUpdateInfo(updates []ProcessInfo) []byte {

	processSet := new(processUtility.ProcessSet)

	for _, curr := range updates {
		currentProcess := PBProcess(curr)
		processSet.Processes = append(processSet.Processes, currentProcess)
	}

	result, err := proto.Marshal(processSet)
	if err != nil {
		panic(err)
	}
	return result
}

func TryParse(outputLine string) (ProcessInfo, bool) {

	outputFields := strings.Fields(outputLine)

	var result ProcessInfo
	nilProcess := ProcessInfo{ProcessID: 0, VirtualSize: 0, TimeStarted: "", Command: "", CPUUsage: 0}
	// checks for first iteration and last iteration of outputFields
	// First iteration is top line of ps created by output
	if len(outputFields) != 5 {
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
	cpuu, err := strconv.ParseFloat(outputFields[4], 64)
	if err != nil {
		return nilProcess, false
	}

	// Puts outputline data into processinfo struct and returns it
	result.ProcessID = int32(pid)
	result.VirtualSize = int32(vsz)
	result.TimeStarted = outputFields[2]
	result.Command = outputFields[3]
	result.CPUUsage = cpuu

	return result, true
}

// Finds all processes that are in the primary map but not in the secondary map
func MapDifference(primary map[int32]ProcessInfo, secondary map[int32]ProcessInfo) []ProcessInfo {

	result := make([]ProcessInfo, 0)

	for id, process := range primary {
		if _, ok := secondary[id]; !ok {
			result = append(result, process)
		}
	}

	return result

}

// Finds all updated processes by comparing them with the HasProcessChanged function
func getProcessesChanged(currentProcessMap map[int32]ProcessInfo, oldProcessMap map[int32]ProcessInfo) []ProcessInfo {

	processesUpdated := make([]ProcessInfo, 0)

	for _, currentProcess := range currentProcessMap {
		if oldProcess, ok := oldProcessMap[currentProcess.ProcessID]; ok {
			if HasProcessChanged(currentProcess, oldProcess) {
				processesUpdated = append(processesUpdated, currentProcess)
			}
		}
	}

	return processesUpdated

}

// Writes protobufs to given file and ensures they're delimited
// by 8 byte size so they can be decoded
func WriteProcessInfo(processData []byte, outputFile *os.File) {

	processBuffer := proto.NewBuffer(nil)
	length := len(processData)
	err := processBuffer.EncodeVarint(uint64(length))
	if err != nil {
		panic(err)
	}

	outputFile.Write(processBuffer.Bytes())
	// To-do: Make sure len of eventBuffer.Bytes() is not > 8

	// Ensures I write 8 bytes to the file
	blank := make([]byte, 8-len(processBuffer.Bytes()))
	outputFile.Write(blank)
	outputFile.Write(processData)

}

func GetProcessInfo(updatePeriod time.Duration) /* ([]byte, error)*/ {

	oldProcessMap := make(map[int32]ProcessInfo)

	updateFile, _ := os.Create("output.txt") // todo: implement err checking
	defer updateFile.Close()
	eventsFile, _ := os.Create("events.txt")
	defer eventsFile.Close()

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
			}
		}

		processesUpdated := getProcessesChanged(processMap, oldProcessMap)
		// Gets processes started and finished
		processesStarted := MapDifference(processMap, oldProcessMap)
		processesFinished := MapDifference(oldProcessMap, processMap)

		// Events are when Processes Started and Finished
		events := MarshalEventInfo(processesStarted, processesFinished)
		WriteProcessInfo(events, eventsFile)

		// Updates to output are when processes change
		updates := MarshalUpdateInfo(processesUpdated)
		WriteProcessInfo(updates, updateFile)

		fmt.Println("Logged at", now)

		// set oldProcessMap to current
		oldProcessMap = processMap

	}
}

func main() {

	updateAmount, _ := strconv.Atoi(os.Args[1]) // If an error is thrown, an invalid amount was entered
	updatePeriod := time.Duration(updateAmount)

	GetProcessInfo(updatePeriod)

}
