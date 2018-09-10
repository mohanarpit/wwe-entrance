package main

import (
	"bytes"
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"io/ioutil"
	"os"
	"os/exec"
	"regexp"
	"strings"
	"time"
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

func parseArpOutput(output []byte) (devices Devices, err error) {

	// Parse the output of ARP Command
	lines := strings.Split(string(output), "\n")
	for _, line := range lines {
		ipRegex := regexp.MustCompile("\\((.*?)\\)")
		ipMatch := ipRegex.FindStringSubmatch(line)
		if len(ipMatch) < 2 {
			continue
		}

		macAddressRegex := regexp.MustCompile("\\) at (.*?) ")
		macMatch := macAddressRegex.FindStringSubmatch(line)
		if len(macMatch) < 2 {
			continue
		}
		device := DeviceInfo{
			IP:         ipMatch[1],
			MacAddress: macMatch[1],
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
	fmt.Printf("In the main")
	var defaultAudioCmd = flag.String("default-audio", "/usr/local/bin/vlc", "Default audio player command")
	flag.Parse()
	fmt.Printf("Default audio: %s", *defaultAudioCmd)

	config, err := parsePropertyFile("config.json")
	if err != nil {
		fmt.Println(err)
	}
	fmt.Printf("Configs : %+v", config)

	// Clear the ARP cache and rebuild it
	// output, _ := exec.Command("sh", "-c", "sudo arp -a -d").Output()
	// fmt.Printf("\nARP Clear Output : %s\n", output)

	// time.Sleep(30 * time.Second)
	//Checking the ARP cache for connected devices
	output, _ := exec.Command("sh", "-c", "ifconfig | grep broadcast | arp -a | grep -v incomplete").Output()
	fmt.Printf("\nARP Output : %s\n", output)
	devices, _ := parseArpOutput(output)
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
