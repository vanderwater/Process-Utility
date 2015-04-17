package main

import (
	"bufio"
	"fmt"
	"github.com/golang/protobuf/proto"
	"os"
	"processUtility"
	"strconv"
	"strings"
	//	"flag"		todo: implement -h
)

// To-do:
// Refactor!
// Try to look at ps and see how they print in that format
// Remove blank field if the process has been updated
// Alter Unmarshal function

// Exact same as in main.go but has two bools, WasOpened and WasClosed
type ProcessInfo struct {
	ProcessID, VirtualSize int32
	Command, TimeStarted   string
	CPUUsage               float64
	WasOpened, WasClosed   bool
}

// Takes an unmarshalled Process and puts it in a ProcessInfo
func DecodeProcess(input *processUtility.Process) ProcessInfo {
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

// Creates a string of the processInfo, currently ugly
func FormatProcess(input ProcessInfo) string {
	goodName := make([]string, 6)
	goodName[0] = fmt.Sprintf("Process %d", input.ProcessID)
	if input.WasOpened {
		goodName[1] = "was opened"
	} else if input.WasClosed {
		goodName[1] = "was closed"
	} else {
		goodName[1] = ""
	}
	goodName[2] = input.TimeStarted
	goodName[3] = input.Command
	goodName[4] = strconv.Itoa(int(input.VirtualSize))
	goodName[5] = strconv.FormatFloat(input.CPUUsage, 'f', 3, 64)

	result := strings.Join(goodName, ", ")
	return result + "\n"

}

// Writes a process Set to file
func PrintProcessSet(outputFile *os.File, processSet *processUtility.ProcessSet) {

	// Get Processes
	processes := processSet.GetProcesses()
	// Decode Processes
	for _, current := range processes {
		processInfo := DecodeProcess(current)
		outputString := FormatProcess(processInfo)
		// I should check return values of Write
		outputFile.WriteString(outputString)
	}
}

// Reads an entire file and returns all data of the file in a buffer
func ReadFile(currentFile *os.File) []byte {

	// Get length of file
	fileInfo, _ := currentFile.Stat()
	fileSize := int(fileInfo.Size())

	// I should probably make sure size isn't humongous so it doesn't read in a massive file

	fileReader := bufio.NewReaderSize(currentFile, fileSize)

	result, err := fileReader.Peek(fileSize)

	if err != nil {
		panic(err)
	}

	return result

	// So I was thinking of doing reads until I read the whole file
	// But I figured peek can do that all in one call...
	// There's definitely something I'm missing
	//	var currentPosition inti = 0
	//	for currentPosition < fileSize {
	//	Read....
	//	}

}

func UnmarshalProcessSet(events *os.File, updates *os.File, demarshalled *os.File) {

	// Start by Decoding a Varint
	// Read Varint bytes from file
	// Demarshal into a ProcessSet
	// Use processUtility.GetProcesses(procesSet) to get slice
	// Send each one into DecodeProcess Function
	// Send ProcessInfo to Print Process

	// To-do: This assumes updates and events were marshalled without error, if one errors without the other
	// Then this doesn't run since I only check eventsPosition and eventsSize

	eventsData := ReadFile(events)
	updatesData := ReadFile(updates)

	// To-do: the classic error check
	eventsInfo, _ := events.Stat()
	eventsSize := int(eventsInfo.Size())

	var eventsPosition int = 0
	var updatesPosition int = 0

	// I could relegate most of this work to a function call and change this to:
	// 	GetSizeOfSet
	//	increment positions
	//	Unmarshal and print set

	for eventsPosition < eventsSize {
		// slice off 8 bytes
		// Decode the fixed32 from that
		setBuffer := proto.NewBuffer(eventsData[eventsPosition : eventsPosition+8])
		eventsPosition += 8

		setLength32, err := setBuffer.DecodeVarint()
		if err != nil {
			panic(err)
		}
		// Need to convert to regular integer
		setLength := int(setLength32)

		eventsSet := new(processUtility.ProcessSet)

		err = proto.Unmarshal(eventsData[eventsPosition:eventsPosition+setLength], eventsSet)
		if err != nil {
			panic(err)
		}

		eventsPosition += setLength

		PrintProcessSet(demarshalled, eventsSet)

		// Literally copypasted from above... function calls could help this

		updatesBuffer := proto.NewBuffer(updatesData[updatesPosition : updatesPosition+8])
		updatesPosition += 8

		setLength32, err = updatesBuffer.DecodeVarint()
		if err != nil {
			panic(err)
		}
		// Need to convert to regular integer
		setLength = int(setLength32)

		updatesSet := new(processUtility.ProcessSet)

		err = proto.Unmarshal(updatesData[updatesPosition:updatesPosition+setLength], updatesSet)
		if err != nil {
			panic(err)
		}

		updatesPosition += setLength

		PrintProcessSet(demarshalled, updatesSet)

	}

}

func main() {

	fileo, _ := os.Open("output.txt") // todo: implement err checking
	defer fileo.Close()
	filee, _ := os.Open("events.txt")
	defer filee.Close()
	filed, _ := os.Create("unmarshal.txt")
	defer filed.Close()

	// until fileo end of file
	// Decode output.txt
	// Decode events.txt
	UnmarshalProcessSet(filee, fileo, filed)
}
