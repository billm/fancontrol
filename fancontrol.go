package main

import (
	"bufio"
	"fmt"
	"log"
	"os"
	"os/exec"
	"strconv"
	"strings"
	"time"
)

var fanLowSpeed float64 = 80
var fanLowTemp float64 = 30

var fanHighSpeed float64 = 255
var fanHighTemp float64 = 80

// Figure out the number of CPU sockets
func getSocketCount() int {
	totSockets := 0

	file, err := os.OpenFile("/proc/cpuinfo", os.O_RDONLY, os.ModePerm)
	if err != nil {
		log.Fatal(err)
		os.Exit(1)
	}
	defer file.Close()

	sc := bufio.NewScanner(file)
	for sc.Scan() {
		line := sc.Text()
		if strings.Contains(line, "physical id") {
			socketId, _ := strconv.Atoi(strings.Split(line, ": ")[1])
			if socketId+1 > totSockets {
				totSockets = socketId + 1
			}
		}

	}
	if err := sc.Err(); err != nil {
		log.Fatalf("scan file error: %v", err)
		os.Exit(1)
	}
	return totSockets
}

// Get the CPU package temperature
func getPackageTemp(id int) int {
	// Location of the temperature sensor
	var tempLoc = fmt.Sprintf("/sys/class/thermal/thermal_zone%d/temp", id)
	var packageTemp int

	file, err := os.OpenFile(tempLoc, os.O_RDONLY, os.ModePerm)
	if err != nil {
		log.Fatal(err)
		os.Exit(1)
	}
	defer file.Close()

	sc := bufio.NewScanner(file)
	for sc.Scan() {
		line, _ := strconv.Atoi(sc.Text())
		packageTemp = line / 1000
	}

	return packageTemp
}

// Get the package with the hottest temp
func getMaxPackageTemp(numSockets int) float64 {
	var temps = make([]int, numSockets)
	var maxTemp = 0
	for i := 0; i < numSockets; i++ {
		temps[i] = getPackageTemp(i)
	}

	for i, e := range temps {
		if i == 0 || e > maxTemp {
			maxTemp = e
		}
	}

	return float64(maxTemp)
}

// Manage the fan speed
// TODO refactor this to pull the socket code out, that logic doesn't belong in this function
func manageFans() {
	var fanSpeed float64 = fanLowSpeed

	numSockets := getSocketCount()
	fmt.Printf("Found %d CPU packages\n", numSockets)
	curTemp := getMaxPackageTemp(numSockets)
	fmt.Printf("Current hottest CPU at %dC\n", int(curTemp))

	if curTemp < fanLowTemp {
		fanSpeed = fanLowSpeed
	} else if curTemp > fanHighTemp {
		fanSpeed = fanHighSpeed
	} else {
		//var fanSpeedF float64 = fanLowSpeed + ((curTemp - fanLowTemp) / (fanHighTemp - fanLowTemp) * (fanHighSpeed - fanLowSpeed))
		var fanSpeedF float64 = ((curTemp - fanLowTemp) / (fanHighTemp - fanLowTemp)) * (fanHighSpeed)
		fanSpeed = fanSpeedF
	}
	fmt.Printf("Setting fan speed to %d%%\n", int((fanSpeed/fanHighSpeed)*100))

	ipmiArgs := strings.Split(fmt.Sprintf("/usr/bin/ipmitool raw 0x30 0x91 0x5A 0x3 0x10 0x%x", int(fanSpeed)), " ")
	ipmitool := exec.Command("./ipmitool.sh")
	ipmitool.Args = ipmiArgs
	//_, err := ipmitool.Output()
	ipmitool.Run()
	/*if err != nil {
		fmt.Println(err.Error())
		return
	}
	*/
}

// Set zone 1 (0x01) to full speed (0x01).
func initializeFans() {
	//ipmiArgs := strings.Split(fmt.Sprintf("raw 0x30 0x91 0x5a 0x03 0x10 0x01"), " ")
	ipmiArgs := strings.Split(fmt.Sprintf("/usr/bin/ipmitool raw 0x30 0x45 0x01 0x01"), " ")
	ipmitool := exec.Command("/usr/bin/ipmitool")
	ipmitool.Args = ipmiArgs
	ipmitool.Run()
}

func resetBmc() {
	ipmiArgs := strings.Split(fmt.Sprintf("/usr/bin/ipmitool bmc reset cold"), " ")
	ipmitool := exec.Command("/usr/bin/ipmitool")
	ipmitool.Args = ipmiArgs
	ipmitool.Run()
}

func main() {
	// Shouldn't be needed, but reset the BMC to clear all the things
	//resetBmc()

	initializeFans()
	for true {
		manageFans()
		time.Sleep(time.Second * 5)
	}
}
