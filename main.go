package main

import (
	"flag"
	"github.com/Vanderwater/Process-Utility/src"
	"os"
	"time"
)

// TODO/Vanderwater:
// Force files to end in .txt

var (
	// These are all pointers
	unmarshal     = flag.Bool("unmarshal", false, "Unmarshals Process Info from Protobufs")
	interval      = flag.Int("interval", 5, "Gathers process info every x seconds")
	updateFile    = flag.String("updateFile", "update.txt", "Filename to record significant changes")
	eventFile     = flag.String("eventFile", "events.txt", "Filename to record closed and openend processes")
	unmarshalFile = flag.String("unmarshalFile", "unmarshal.txt", "Filename to print unmarshalled data")
)

func main() {

	flag.Parse()

	// Add Error Checking
	if !(*unmarshal) {
		updatePeriod := time.Duration(*interval)
		processUtility.GetProcessInfo(updatePeriod, *updateFile, *eventFile)
	} else {
		fileo, _ := os.Open(*updateFile)
		defer fileo.Close()
		filee, _ := os.Open(*eventFile)
		defer filee.Close()
		filed, _ := os.Create(*unmarshalFile)
		defer filed.Close()

		processUtility.UnmarshalProcessSet(filee, fileo, filed)
	}

}
