package main

import (
	"fmt"
	"net"
	"os"
	"strings"
)

const SocketPath = "/tmp/lokstt.sock"

func main() {
	if len(os.Args) < 2 {
		sendCommand("TOGGLE")
		return
	}

	cmd := strings.ToUpper(os.Args[1])
	if cmd != "START" && cmd != "STOP" && cmd != "TOGGLE" && cmd != "SETTINGS" && cmd != "CANCEL" {
		fmt.Println("Usage: client [START|STOP|TOGGLE|SETTINGS|CANCEL]")
		return
	}

	sendCommand(cmd)
}

func sendCommand(cmd string) {
	c, err := net.Dial("unix", SocketPath)
	if err != nil {
		fmt.Println("Error connecting to daemon. Is it running?")
		return
	}
	defer c.Close()

	_, err = c.Write([]byte(cmd))
	if err != nil {
		fmt.Println("Write error:", err)
		return
	}

	buf := make([]byte, 1024)
	n, err := c.Read(buf)
	if err != nil {
		fmt.Println("Read error:", err)
		return
	}

	fmt.Print(string(buf[:n]))
}
