package main

import(
	"fmt"
	"os/exec"
//	"strconv"
//	"strings"
)


// So for some reason you need to use Output and then print ps in order to get all the data, otherwise you don't get the proper output
// Very Strange

func main(){
	// Call ps for process info, -e gives all, -o gives specified output
	ps := exec.Command("ps", "-e", "-o pid,comm,time,vsz") 
	output, _ := ps.Output()
	fmt.Println(ps)
//	fmt.Println(output)
	_ = output
	}
