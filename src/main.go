package main

import (
	"fmt"
	"net"
	"os/exec"
	"regexp"
	"strconv"
	"strings"
	"time"
)

// IP address regex pattern
var pattern, _ = regexp.Compile("^([01]?\\d\\d?|2[0-4]\\d|25[0-5])\\.([01]?\\d\\d?|2[0-4]\\d|25[0-5])\\.([01]?\\d\\d?|2[0-4]\\d|25[0-5])\\.([01]?\\d\\d?|2[0-4]\\d|25[0-5])$")

// arp line matcher
var arpPattern, _ = regexp.Compile(".*(00-0d-b9).*")

// request an IP address from console and validates it
func getIPAddress() string {
	var validIP = false
	var ip string
	for validIP == false {
		fmt.Scanln(&ip)
		if len(ip) < 6 || !pattern.MatchString(ip) {
			fmt.Print("Enter a valid address: ")
		} else {
			validIP = true
		}
	}
	return ip
}

// returns the last number in an IP address
func getLastOctet(s string) int {
	split := strings.Split(s, ".")
	lastOctet, _ := strconv.Atoi(split[len(split)-1])
	return lastOctet
}

// checks wether the given host has port 80 open
func checkHost(ipAddress string, suffix int, c chan string) {
	_, err := net.DialTimeout("tcp", ipAddress+":80", time.Second*3)
	if err == nil {
		c <- strconv.Itoa(suffix) + " is up"
	} else {
		c <- strconv.Itoa(suffix) + " is down"
	}
}

// waits for all hosts scan to complete and then matches "arp -a" MACs
func checkResults(max int, c chan string) {
	current := 0
	for {
		fmt.Print(<-c + ", ")
		current++
		if current == max {
			fmt.Println("\n\nDone scanning, checking MACs...")
			break
		}
	}

	// execute arp -a and grap output of it
	output, _ := exec.Command("arp", "-a").Output()
	splitted := strings.Split(string(output), "\n")

	// get lines with cockpit/a4w MACs
	fmt.Println("IPs matching Cockpit/A4W MAC:")
	for _, v := range splitted {
		match := strings.TrimSpace(arpPattern.FindString(v))
		if len(match) > 5{
			fmt.Println(strings.TrimSpace(arpPattern.FindString(match)))		
		}
	}
}

// holds the program while goroutines are running
func wait() {
	var breaker string
	fmt.Scanln(&breaker)
}

func main() {

	// output some infos
	fmt.Println("" +
		"Cockpit/A4W Finder - Luca Moser " +
		"\nFinds cockpits/a4w in your LAN. Performs searches only on a 24 bit subnetmask!\n")

	// get start IP
	var startIP string
	fmt.Print("Enter start IP address: ")
	startIP = getIPAddress()

	// get end IP
	var endIP string
	fmt.Print("Enter end IP address: ")
	endIP = getIPAddress()

	// compute last octet numbers
	startHost := getLastOctet(startIP)
	endHost := getLastOctet(endIP)
	difference := endHost - startHost

	if endHost < startHost {
		fmt.Println("Ending IP address is bigger than starting address! Terminating.")
		wait()
		return
	}

	// input information
	fmt.Println("\nStart address:", startIP)
	fmt.Println("End address:", endIP)
	fmt.Println("Start host:", startHost)
	fmt.Println("End host:", endHost)
	fmt.Println("Amount of hosts to scan:", difference)

	// get IP address prefix
	prefix := startIP[0:strings.LastIndex(startIP, strconv.Itoa(startHost))]

	// spawn channel to get results from goroutines
	var c chan string = make(chan string, difference)
	go checkResults(difference, c)

	fmt.Println("\nScanning...")
	// loop through addresses and check hosts
	for ; startHost < endHost; startHost++ {
		hostToScan := prefix + strconv.Itoa(startHost)
		//fmt.Println("Checking host",hostToScan)
		go checkHost(hostToScan, startHost, c)
	}

	wait()
}
