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
	// Is it possible that addresses won't be cleaned up or could be incorrect due to this?
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
// To-do: check for errors before returning to keep error checking within this function
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

func MapDifference(primary map[int32]ProcessInfo, secondary map[int32]ProcessInfo) []ProcessInfo {

	result := make([]ProcessInfo, 0)

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

		processesUpdated := make([]ProcessInfo, 0)

		output := GetCurrentProcesses()
		outputLines := strings.Split(string(output), "\n")
		// Loop over all the outputs, check against current slice of processes
		for _, line := range outputLines {
			// To-do: Something if its invalid
			if currentProcess, valid := TryParse(line); valid {
				processMap[currentProcess.ProcessID] = currentProcess
				// So this mess updates whether the process has significantly changed or not
				if oldProcess, ok := oldProcessMap[currentProcess.ProcessID]; ok {
					if HasProcessChanged(currentProcess, oldProcess) {
						processesUpdated = append(processesUpdated, currentProcess)
					}
				}

			}

		}

		// Gets processes started and finished
		processesStarted := MapDifference(processMap, oldProcessMap)
		processesFinished := MapDifference(oldProcessMap, processMap)

		// Marshals Processes
		events, err := MarshalEventInfo(processesStarted, processesFinished)
		if err != nil {
			panic(err)
		}

		// Events are when Processes Started and Finished
		eventBuffer := proto.NewBuffer(nil)
		length := len(events)
		err = eventBuffer.EncodeVarint(uint64(length))
		if err != nil {
			panic(err)
		}

		filee.Write(eventBuffer.Bytes())
		// So this is pretty hacky, but it ensures I write 8 bytes to the file
		blank := make([]byte, 8-len(eventBuffer.Bytes()))
		filee.Write(blank)
		filee.Write(events)

		// Updates to output are when processes change
		updates, err := MarshalUpdateInfo(processesUpdated)
		if err != nil {
			panic(err)
		}
		updateBuffer := proto.NewBuffer(nil)
		length = len(updates)
		updateBuffer.EncodeVarint(uint64(length))
		fileo.Write(updateBuffer.Bytes())
		// Hacky thing from above
		blank = make([]byte, 8-len(updateBuffer.Bytes()))
		fileo.Write(blank)
		fileo.Write(updates)

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
