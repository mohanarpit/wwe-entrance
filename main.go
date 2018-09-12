package main

import (
	"bufio"
	"bytes"
	"context"
	"crypto/tls"
	"encoding/json"
	"flag"
	"fmt"
	"io/ioutil"
	"net"
	"os"
	"os/exec"
	"regexp"
	"strings"
	"time"

	oi "github.com/reiver/go-oi"
	"github.com/reiver/go-telnet"
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

	//Connect to the router
	err = connectRouter()
	if err != nil {
		fmt.Printf("Error in connecting to router: %+v", err)
		return
	}

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

type caller struct{}

func (c caller) CallTELNET(ctx telnet.Context, w telnet.Writer, r telnet.Reader) {
	fmt.Println("In the callTELNET")
	scanner := bufio.NewScanner(os.Stdin)
	for scanner.Scan() {
		oi.LongWrite(w, scanner.Bytes())
		oi.LongWrite(w, []byte("Admin\n"))
		oi.LongWrite(w, []byte("admin\n"))
		oi.LongWrite(w, []byte("ls\n"))
	}
	var telnetResponse = make([]byte, 0)
	bytesRead, err := r.Read(telnetResponse)
	if err != nil {
		fmt.Printf("%+v", err)
	}
	fmt.Printf("BytesRead: %d", bytesRead)
}

func connectRouter() error {
	fmt.Println("In the connectRouter")
	conn, err := net.Dial("tcp", "192.168.0.1:23")
	defer conn.Close()

	if err != nil {
		return err
	}
	resp, err := bufio.NewReader(conn).ReadString(':')
	fmt.Println(resp)
	fmt.Fprintf(conn, "Admin\n")
	resp, err = bufio.NewReader(conn).ReadString(':')
	fmt.Println(resp)
	fmt.Println("sent command")
	fmt.Fprintf(conn, "admin\n")
	fmt.Println("sent passwd")
	resp, err = bufio.NewReader(conn).ReadString('#')
	fmt.Println(resp)
	fmt.Fprintf(conn, "ls\n")
	fmt.Println("sent ls")
	resp, err = bufio.NewReader(conn).ReadString('\n')
	fmt.Println(resp)
	fmt.Fprintf(conn, "\n")
	fmt.Println("sent ls")
	resp, err = bufio.NewReader(conn).ReadString('\n')
	fmt.Println(resp)

	if err != nil {
		return err
	}

	return nil
}

func connectRouterTelnet() error {
	fmt.Println("In the connectRouter")
	//@TODO: Configure the TLS connection here, if you need to.
	tlsConfig := &tls.Config{}

	//@TOOD: replace "example.net:5555" with address you want to connect to.
	err := telnet.DialToAndCallTLS("192.168.0.1:23", caller{}, tlsConfig)

	if err != nil {
		return err
	}
	return nil
}
