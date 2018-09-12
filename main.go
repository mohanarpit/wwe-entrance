package main

import (
	"bytes"
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"os/exec"
	"strings"
	"time"

	"github.com/mohanarpit/wwe-entrance/router"
)

type Config struct {
	MacAddress string `json:"mac_address"`
	SoundFile  string `json:"sound_file"`
}

type DeviceInfo struct {
	IP         string
	MacAddress string
}

type Devices []DeviceInfo

type macAddresses []string

func (m *macAddresses) String() string {
	return "These are the mac ids"
}

func (m *macAddresses) Set(value string) error {
	*m = append(*m, value)
	return nil
}

func parsePropertyFile(filename string) (config []Config, err error) {
	jsonFile, err := os.Open(filename)
	if err != nil {
		return nil, err
	}

	defer jsonFile.Close()
	byteValue, _ := ioutil.ReadAll(jsonFile)
	json.Unmarshal(byteValue, &config)
	return config, err
}

func parseArpOutput(output string) (devices Devices, err error) {

	// Parse the output of ARP Command
	lines := strings.Split(output, "\n")
	for _, line := range lines {
		// This function tokenizes the line by any number of whitespaces whitespaces
		fields := strings.Fields(line)

		if len(fields) < 4 {
			continue
		}

		device := DeviceInfo{
			IP:         fields[0],
			MacAddress: fields[3],
		}
		devices = append(devices, device)
	}
	return devices, nil
}

func playMusic(devices Devices, config []Config, audioCmd string) error {
	for _, device := range devices {
		if device.MacAddress == config[0].MacAddress {
			fmt.Printf("\nFound a match: %s", device.MacAddress)
			ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
			defer cancel()

			musicCmd := exec.CommandContext(ctx, "sh", "-c", audioCmd+" "+config[0].SoundFile)
			var out bytes.Buffer
			var stderr bytes.Buffer
			musicCmd.Stdout = &out
			musicCmd.Stderr = &stderr
			err := musicCmd.Run()
			if err != nil {
				fmt.Printf(fmt.Sprint(err) + ": " + stderr.String())
				return err
			}
			if ctx.Err() == context.DeadlineExceeded {
				fmt.Println("Deadline exceeded")
				return ctx.Err()
			}

			fmt.Printf("\nMusic Output: %+v", out.String())
		}
	}
	return nil
}

func remove(slice Devices, s int) Devices {
	return append(slice[:s], slice[s+1:]...)
}

func main() {
	var defaultAudioCmd = flag.String("default-audio", "/usr/local/bin/vlc", "Default audio player command")
	var routerUsername = flag.String("username", "Admin", "The username for your router login")
	var routerPwd = flag.String("password", "Password", "The password for your router login")
	var propertyFile = flag.String("property-file", "config.json", "The location of the property file")
	flag.Parse()

	config, err := parsePropertyFile(*propertyFile)
	if err != nil {
		fmt.Println(err)
	}

	//Connect to the router
	dlink := router.DlinkRouter{
		ConnectionType: "tcp",
		Command:        "cat /proc/net/arp",
	}

	conn, err := dlink.Connect(*routerUsername, *routerPwd, "192.168.0.1:23")
	if err != nil {
		log.Fatalln(err)
	}
	defer conn.Close()

	output, err := dlink.GetArpOutput(conn)
	if err != nil {
		log.Printf("Error in connecting to router: %+v", err)
		return
	}

	devices, _ := parseArpOutput(string(output))
	fmt.Printf("\nDevices: %+v", devices)
	var oldDevices Devices
	if oldDevices == nil {
		oldDevices = devices
	}

	// De-duplicate the device and run a check on only new devices
	// for newIdx, device := range devices {
	// 	for _, oldDevice := range oldDevices {
	// 		if device.MacAddress == oldDevice.MacAddress {
	// 			devices = remove(devices, newIdx)
	// 		}
	// 	}
	// }

	fmt.Printf("\nNew Devices: %+v", devices)

	playMusic(devices, config, *defaultAudioCmd)
}
