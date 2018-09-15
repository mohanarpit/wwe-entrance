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

// Config is the data representation of the config.json file where the user can define the
// devices for which music must be played
type Config struct {
	MacAddress string `json:"mac_address"`
	SoundFile  string `json:"sound_file"`
}

// MusicInfo is the data representation for passing the actual song file and command to be used
type MusicInfo struct {
	SoundFile string
	Command   string
}

// DeviceInfo is the data representation of the parsed arp output
type DeviceInfo struct {
	IP         string
	MacAddress string
}

// DeviceMap is the map of devices to their mac address. Makes it easy for us to check if a device is new
type DeviceMap map[string]DeviceInfo

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

func parseArpOutput(output string) (deviceMap DeviceMap, err error) {
	if deviceMap == nil {
		deviceMap = make(map[string]DeviceInfo)
	}

	// Parse the output of ARP Command
	lines := strings.Split(output, "\n")
	for _, line := range lines {
		// This function tokenizes the line by any number of whitespaces whitespaces
		fields := strings.Fields(line)

		if len(fields) < 4 {
			continue
		}

		flag := fields[2]
		// 0x2 means the device is connected
		if flag != "0x2" {
			continue
		}

		device := DeviceInfo{
			IP:         fields[0],
			MacAddress: fields[3],
		}
		deviceMap[device.MacAddress] = device
	}
	return deviceMap, nil
}

func getSoundFile(device DeviceInfo, configs []Config) string {
	for _, config := range configs {
		if device.MacAddress == config.MacAddress {
			return config.SoundFile
		}
	}
	return ""
}

func playMusic(musicChannel <-chan MusicInfo, errChan chan<- error) {
	for musicInfo := range musicChannel {
		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()

		musicCmd := exec.CommandContext(ctx, "sh", "-c", musicInfo.Command+" "+musicInfo.SoundFile)
		var out bytes.Buffer
		var stderr bytes.Buffer
		musicCmd.Stdout = &out
		musicCmd.Stderr = &stderr
		err := musicCmd.Run()
		if err != nil {
			log.Printf(fmt.Sprint(err) + ": " + stderr.String())
			errChan <- err
		}
		if ctx.Err() == context.DeadlineExceeded {
			log.Println("Deadline exceeded", ctx.Err())
			errChan <- err
		}
	}
}

func getGatewayAddress() string {
	//TODO: Get the broadcast IP instead of hardcoding it here
	return "192.168.0.1:23"
}

func main() {
	var defaultAudioCmd = flag.String("default-audio", "/usr/local/bin/vlc", "Default audio player command")
	var routerUsername = flag.String("username", "Admin", "The username for your router login")
	var routerPwd = flag.String("password", "Password", "The password for your router login")
	var propertyFile = flag.String("property-file", "config.json", "The location of the property file")
	var delay = flag.Duration("delay", 5*time.Second, "The delay (in seconds) with which the program will attempt to poll the router for connected devices")
	flag.Parse()

	config, err := parsePropertyFile(*propertyFile)
	if err != nil {
		log.Fatalln(err)
	}

	// Initialize the musicInfo channel
	musicChannel := make(chan MusicInfo)
	errChan := make(chan error)
	go playMusic(musicChannel, errChan)

	//Connect to the router
	dlink := router.DlinkRouter{
		ConnectionType: "tcp",
		Command:        "cat /proc/net/arp",
	}

	conn, err := dlink.Connect(*routerUsername, *routerPwd, getGatewayAddress(), *delay)
	if err != nil {
		log.Fatalln(err)
	}
	defer conn.Close()

	// Run the program periodically to check for new devices
	tick := time.Tick(*delay)
	var oldDevices DeviceMap

	for {
		select {
		case <-tick:
			output, err := dlink.GetArpOutput(conn)
			if err != nil {
				log.Printf("Error in connecting to router: %+v", err)
				return
			}

			devices, _ := parseArpOutput(output)
			log.Println("Devices: ", devices)

			// Play the music only if a new device is connecting
			for ip, device := range devices {
				if _, ok := oldDevices[ip]; !ok {
					soundFile := getSoundFile(device, config)
					if soundFile != "" {
						// This is a new device connecting
						musicInfo := MusicInfo{
							SoundFile: soundFile,
							Command:   *defaultAudioCmd,
						}
						musicChannel <- musicInfo
					}
				}
			}
			oldDevices = devices
		case err := <-errChan:
			log.Fatal(err)
		}
	}
}
