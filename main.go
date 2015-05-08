package main

import "fmt"
import "log"
import "net"
import "time"
import "strconv"
import "os/exec"

const screens int = 10

type CapturePane struct {
	snap [screens][]byte
	ts   time.Time
}

var pane CapturePane

func snapshot(screen int) []byte {
	move := exec.Command("tmux", "send-keys", "-t", "rtorrent", strconv.Itoa(screen))
	_, err := move.Output()
	if err != nil {
		fmt.Println(err)
		return []byte{}
	}

	cmd := exec.Command("tmux", "capture-pane", "-t", "rtorrent", "-p")
	stdout, err := cmd.Output()
	if err != nil {
		fmt.Println(err)
		return []byte{}
	}
	return stdout
}

func handleConn(conn net.Conn) {
	defer conn.Close()

	now := time.Now()
	if now.Sub(pane.ts) > 3*time.Second {
		for i := 0; i < screens; i++ {
			pane.snap[i] = snapshot(i)
		}
		pane.ts = now
	}
	snap := pane.snap[5]
	conn.Write(snap)
}

func main() {
	ln, err := net.Listen("tcp", ":4002")
	if err != nil {
		log.Fatal(err)
	}
	for {
		conn, err := ln.Accept()
		if err != nil {
			fmt.Println(err)
		} else {
			go handleConn(conn)
		}
	}
}
