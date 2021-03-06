# NBN Test
Test your nbn connection

This program can be used to record your modem statistics and the results of speedtest.net tests over a period of time

![image](https://user-images.githubusercontent.com/506680/32878365-73ba8f06-caf1-11e7-8cb3-823e6a09708d.png)

# Supported Outputs

* InfluxDB [wiki](https://github.com/freman/nbntest/wiki/Output-InfluxDB)
* Console

# Supported Modems

* TPLink TD-9970
* Technicolor TGiiNet-1
* Technicolor TG789vac
* Generic Telnet via LUA

# Releases

Binary releases for a handful of platforms are [available on the releases page](https://github.com/freman/nbntest/releases) and will basically be any version I think is stable enough not to be awful.

# Building your own

I've started a [wiki page](https://github.com/freman/nbntest/wiki/Developing-Building) covering the quick start to installing the Go build environment and building the binary

# Extras

In the [Extra](extra) dir there's some bits and pieces to get you started including the grafana dashboard in the
screenshot above

# TODO

* Remove the external dependancies for people who want to run it stand alone
  * Internal storage
  * Internal graphing
* Add more modems
* Add a http lua module
