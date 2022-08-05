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

/*
 * Fan control software for the following board(s):
 * Supermicro PIO-617R-TLN4F+-ST031/X9DRi-LN4+/X9DR3-LN4+
 *
 * This is intended to provide a better fan curve for home usage on
 * this motherboards so you neighbors don't think you are working
 * on jet engines.
 *
 * DISCLAIMER: The author of this code provides the code free of
 * charge with no conditions on usage and provides no warrantee or
 * guarantee of any kind. Compile and run this at your own risk
 * if your computer blows up, you win the lotto, or world war 3
 * coincidentally starts after running this code, that is on you.
 *
 * Some day I will put a real license here, until then whatever
 * happens, it is not my fault.
 *
 */

const ipmitool = "/usr/bin/ipmitool"

var fanLowSpeed float64 = 80
var fanLowTemp float64 = 30

var fanHighSpeed float64 = 255
var fanHighTemp float64 = 80

// Temperature on last check
var lastTemp float64

// Temperature on last fan update
var lastSetTemp float64

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

func runIpmiTool(args string) {
	ipmitool := exec.Command(ipmitool)
	ipmitool.Args = strings.Split(args, " ")
	ipmitool.Run()
}

func updateFanSpeed(zone string, speed float64) {
	ipmiArgs := fmt.Sprintf("%s raw 0x30 0x91 0x5A 0x3 %s 0x%x", ipmitool, zone, int(speed))
	runIpmiTool(ipmiArgs)
}

// Manage the fan speed
func manageFans(curTemp float64) {
	var fanSpeed float64 = fanLowSpeed
	var zone = "0x10"

	// Only need to change the fan speed if the temperature has changed
	if int(curTemp) != int(lastTemp) {
		if curTemp <= fanLowTemp {
			fanSpeed = fanLowSpeed
		} else if curTemp >= fanHighTemp {
			fanSpeed = fanHighSpeed
		} else {
			var fanSpeedF float64 = ((curTemp - fanLowTemp) / (fanHighTemp - fanLowTemp)) * (fanHighSpeed)
			fanSpeed = fanSpeedF
		}

		// Only update fan speed if temp has gone up by 2C or more since last
		// time it was set. Or if temp has gone down by 5C or more since last
		// time it was set.
		if 2 >= int(curTemp)-int(lastSetTemp) {
			fmt.Printf("Increasing fan speed to %d%%\n", int((fanSpeed/fanHighSpeed)*100))
			updateFanSpeed(zone, fanSpeed)
			lastSetTemp = curTemp
		} else if 5 >= int(lastSetTemp)-int(curTemp) {
			fmt.Printf("Decreasing fan speed to %d%%\n", int((fanSpeed/fanHighSpeed)*100))
			updateFanSpeed(zone, fanSpeed)
			lastSetTemp = curTemp
		}
	}
	// Cache our last temp
	lastTemp = curTemp
}

// Set zone 1 (0x01) to full speed (0x01).
func initializeFans() {
	// Cooling profiles in BMC
	// Standard: 0
	// Full: 1
	// Optimal: 2
	// Heavy IO: 4
	// We set to Full profile because the BMC seems to stop managing the fans
	// unless there is a thermal violation in this mode, allowing us to fully
	// control the fan profile
	var profile = "0x01"
	ipmiArgs := fmt.Sprintf("%s raw 0x30 0x45 0x01 %s", ipmitool, profile)
	runIpmiTool(ipmiArgs)
}

// Unused function - show what our current fan profile is
func showCurrentFanProfile() {
	ipmiArgs := fmt.Sprintf("%s raw 0x30 0x45 0x00", ipmitool)
	runIpmiTool(ipmiArgs)
}

// Unused function - reset the BMC to clear all the things
func resetBmc() {
	ipmiArgs := fmt.Sprintf("%s bmc reset cold", ipmitool)
	runIpmiTool(ipmiArgs)
}

func main() {
	initializeFans()
	numSockets := getSocketCount()
	fmt.Printf("Found %d CPU packages\n", numSockets)

	for true {
		curTemp := getMaxPackageTemp(numSockets)
		fmt.Printf("Current hottest CPU at %dC\n", int(curTemp))
		manageFans(curTemp)
		time.Sleep(time.Second * 5)
	}
}
