# WWE Entrance Music 

Everyone deserves to enter a room with their favourite tune playing; just like the 
WWE wrestlers. 

This project attempts to do that by polling the WiFi network to check 
if a pre-configured device connects to the Home/Office Wifi network. 

When a known device is seen to change it's status, a pre-configured music file
will be played on the system. This will allow devices to play the music before the 
user enters the room. 

The configurations for the devices and sound files are placed in config.json.

TODOs: 

- [ ] Check the `nmap` and `arp` tools. They aren't giving reliable results
- [ ] Iterate over the config instead of using the hard-coded first result
- [ ] Create a poll every 2 mins to check for new devices
- [ ] Manage statuses for devices instead of iterating over all the devices on the network