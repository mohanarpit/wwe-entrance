# WWE Entrance Music 

Everyone deserves to enter a room with their favourite tune playing; just like the 
WWE wrestlers. 

This project attempts to replicate that behaviour by polling the WiFi network to check 
if a pre-configured device connects to the Home/Office Wifi network. 

When a known device is seen to connect to the WiFi network, a pre-configured music file
will be played on the system. This will allow devices to play the music before the 
user enters the room. 

The configurations for the devices and sound files are placed in config.json.

### Requirements

* Currently only the DLink router DIR-800 is supported. Ideally, you should have a router
  which should be able to 
  * Open a telnet session with the client
  * Ask for username & password in the beginning
  * Switch to a shell once the login succeeds

* A music player that can be invoked from the command line. The default is `/usr/local/bin/vlc`

### How To Run

You can run the program by executing the following command: 

```bash
$ go run main.go -default-audio=/usr/local/bin/vlc -username=router_admin -password=router_password
```

### Options
You can print the options by typing the command
```bash
$ go run main.go --help
```
-  -default-audio: Default audio player command (default "/usr/local/bin/vlc")
-  -delay: The delay (in seconds) with which the program will attempt to poll the router for connected devices (default 5)
-  -password: The password for your router login (default "Password")
-  -property-file: The location of the property file (default "config.json")
-  -username: The username for your router login (default "Admin")


### TODOs: 
- [ ] Get the broadcast IP from the system instead of hardcoding it
- [x] Support for multiple clients connecting at the same time. It should enqueue the music files and play one after    another
- [ ] Manage statuses for devices instead of iterating over all the devices on the network
- [ ] Currently only DLink DIR-800 router is supported. I don't have knowledge of how other routers work
