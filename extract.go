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

// unmarshalled.txt will be output file
// take 0 arguments
// use events.txt and output.txt
// Print format:
// Processes Information
// 	-process info- started
//	-process info- changed
//	-process info- changed
//	-process info- exited

// asdfasdfasdfasdfasdfasdfasdfadfssdfasdf
// asdfasdfasdfasdfasdfasdfasdfasdfasfasdf
