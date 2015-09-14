package processUtility

import (
	"bufio"
	"fmt"
	"github.com/golang/protobuf/proto"
	"github.com/vanderwater/processutility/proto"
	"math"
	"os"
	"os/exec"
	"strconv"
	"strings"
	"time"
)

/* TODO/Vanderwater:
*	Make unmarshal more modular
*	Make it easier to extend ProcessInfo
*	Go-test
*	Make this idiomatic Go!!!!!
 */

// In order to add a new field to Protobuf
// Add field to Process Info, update TryParse and PBProcess

// make a config file
// const (name=value)

type ProcessInfo struct {
	ProcessID            int32
	VirtualSize          int32
	Command              string
	TimeStarted          string
	CPUUsage             float64
	WasOpened, WasClosed bool
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
func PBProcess(src ProcessInfo) *processProto.Process {

	result := new(processProto.Process)
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

// Marshals started and terminated process Info
func MarshalEventInfo(start []ProcessInfo, finished []ProcessInfo) []byte {

	processSet := new(processProto.ProcessSet)

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

	processSet := new(processProto.ProcessSet)

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
	// Just put in the thing
	nilProcess := ProcessInfo{ProcessID: 0, VirtualSize: 0, TimeStarted: "", Command: "", CPUUsage: 0}
	// checks for first iteration and last iteration of outputFields
	// First iteration is top line of ps created by output
	if len(outputFields) != 5 {
		// TODO/Vanderwater: Getting error on returning nil, figure out what to return
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
func WriteProcessInfo(processData []byte, outputFile io.Writer) {

	processBuffer := proto.NewBuffer(nil)
	length := len(processData)
	err := processBuffer.EncodeVarint(uint64(length))
	if err != nil {
		panic(err)
	}

	outputFile.Write(processBuffer.Bytes())
	// TODO/Vanderwater: Make sure len of eventBuffer.Bytes() is not > 8
	if len(processBuffer.Bytes() > 8) {
		panic("Set size too large, increase delimiter to the next power of 2 in config")
	}
	// Ensures I write 8 bytes to the file
	blank := make([]byte, 8-len(processBuffer.Bytes()))
	outputFile.Write(blank)
	outputFile.Write(processData)

}

func GetProcessInfo(updatePeriod time.Duration, updateName string, eventName string) /* ([]byte, error)*/ {

	oldProcessMap := make(map[int32]ProcessInfo)

	updateFile, _ := os.Create(updateName) // todo: implement err checking
	defer updateFile.Close()
	eventsFile, _ := os.Create(eventName)
	defer eventsFile.Close()

	clock := time.NewTicker(updatePeriod * time.Second)

	// removed go func{ }() around this for loop
	for now := range clock.C {

		processMap := make(map[int32]ProcessInfo)

		output := GetCurrentProcesses()
		outputLines := strings.Split(string(output), "\n")
		// Loop over all the outputs, check against current slice of processes
		for _, line := range outputLines {
			// TODO/Vanderwater: Something if its invalid
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

// TODO/Vanderwater::
// Refactor!
// Try to look at ps and see how they print in that format
// Remove blank field if the process has been updated

// Takes an unmarshalled Process and puts it in a ProcessInfo
func DecodeProcess(input *processProto.Process) ProcessInfo {
	var result ProcessInfo

	result.ProcessID = input.GetProcessID()
	result.VirtualSize = input.GetVirtualSize()
	result.Command = input.GetCommand()
	result.TimeStarted = input.GetTimeStarted()
	result.CPUUsage = input.GetCPUUsage()
	result.WasOpened = input.GetWasOpened()
	result.WasClosed = input.GetWasClosed()

	return result
}

func FormatProcess(input ProcessInfo) string {
	processBuilder := make([]string, 6)
	processBuilder[0] = fmt.Sprintf("Process %d", input.ProcessID)
	if input.WasOpened {
		processBuilder[1] = "was opened"
	} else if input.WasClosed {
		processBuilder[1] = "was closed"
	} else {
		processBuilder[1] = ""
	}
	processBuilder[2] = input.TimeStarted
	processBuilder[3] = input.Command
	processBuilder[4] = strconv.Itoa(int(input.VirtualSize))
	processBuilder[5] = strconv.FormatFloat(input.CPUUsage, 'f', 3, 64)

	result := strings.Join(processBuilder, ", ")
	return result + "\n"

}

// Writes a process Set to file
func PrintProcessSet(outputFile *os.File, processSet *processProto.ProcessSet) {

	processes := processSet.GetProcesses()
	for _, current := range processes {
		processInfo := DecodeProcess(current)
		outputString := FormatProcess(processInfo)
		// I should check return values of Write
		outputFile.WriteString(outputString)
	}
}

// Reads an entire file and returns all data of the file in a buffer
func ReadFile(currentFile io.Reader) []byte {

	// Get length of file
	fileInfo, _ := currentFile.Stat()
	fileSize := int(fileInfo.Size())

	fileReader := bufio.NewReaderSize(currentFile, fileSize)

	result, err := fileReader.Peek(fileSize)

	if err != nil {
		panic(err)
	}

	return result

}

func getSizeOfData(set []byte) int {
	//Implement Error Checking
	eventsInfo, _ := events.Stat()
	return int(eventsInfo.Size())
}

func unmarshalNextSet(incomingData []byte, setPosition int) ([]byte, int) {
	setBuffer := proto.NewBuffer(incomingData[setPosition : setPosition+8])
	setLength32, _ := setBuffer.DecodeVarint()
	// Add error checking
	nextSet := new(processProto.ProcessSet)
	setPosition += 8
	err = proto.Unmarshal(eventsData[eventsPosition:eventsPosition+setLength], nextSet)
	if err != nil {
		panic(err)
	}
	return nextSet, setPosition + setLength
}

// TODO/Vanderwater: I reaaallly need to rename these file pointers so they are obviously file pointers
func UnmarshalProcessSet(events io.Reader, updates io.Reader, demarshalled io.Writer) {

	eventsData := ReadFile(events)
	updatesData := ReadFile(updates)
	eventsSize := getSizeOfData(events)

	var eventsPosition int = 0
	var updatesPosition int = 0

	for eventsPosition < eventsSize && updatesPosition < updatesSize {
		eventsSet, eventsPosition := unmarshalNextSet(eventsData, eventsPosition)
		PrintProcessSet(demarshalled, eventsSet)

		updatesSet, updatesPosition := unmarshalNextSet(updatesData, updatesPosition)
		PrintProcessSet(demarshalled, updatesSet)
	}

	if eventsPosition < eventsSize {
		panic("All of Events not read\n")
	}
	if updatesPosition < updatesSize {
		panic("All of Updates not read\n")
	}

}
