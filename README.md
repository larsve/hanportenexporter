# Hanporten exporter
Hanporten exporter exposes OBIS data from a Swedish smartmeter as Prometheus metrics.

This software connects to a ESP8266/Tasmota TCP/serial bridge to read serial data from the smartmeter, that then is exposed as Prometheus metrics on port `9102`. The program just have one option, and that is what TCP/serial bridge(s) to connect to.

    go install github.com/larsve/hanportenexporter
    hanportenexporter tasmota:8232

# Setup
An ESP8266 with the [Tasmota ZigbeeBridge](https://tasmota.github.io/install/) firmware is connected to the smartmeter through HAN-porten (aka P1/H1) and configured as a serial bridge. See [Hanporten.se](https://hanporten.se/) for more information about the Swedish smartmeter HAN/H1/P1-port.

## Configure ESP8266/Tasmota ZigbeeBridge serial bridge
Connect to the Tasmota web GUI on the ESP8266 in a webbrowser, and complete the steps below to setup the TCP/Serial bridge:
 1. In Main Menu -> Configuration -> Configure Module
    * Set `Module type` to `Generic (18)`
    * Click `Save`.
 2. In Main Menu -> Configuration -> Configure Module
    * Set `TX GPIO1` to `TCP Tx`
    * Set `RX GPIO3` to `TCP Rx`
    * Click `Save`.
 3. In Main Menu -> Consoles -> Console
    * Write `Rule1 + ON System#Boot DO Backlog TCPBaudRate 115200; TCPStart 8232 ENDON` to setup the rule that will start the serial bridge when the device starts (the number after `TCPStart` is the TCP port that the serial bridge will listen on).
    * Write `Rule1 1` to trigger the rule to start the serial bridge without restarting the device.
 4. Start hanportenexporter with the name/IP of the ESP8266 and the port that the TCP serial bridge is configured to listen on.
