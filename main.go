package main

import "fmt"
import "log"
import "time"
import "net/http"
import "strconv"
import "os/exec"

const screens int = 10

var pageTable = map[string]int{
    "/active":     0, "/0": 0,
    "/main":       1, "/1": 1,
    "/name":       2, "/2": 2,
    "/started":    3, "/3": 3,
    "/stopped":    4, "/4": 4,
    "/complete":   5, "/5": 5,
    "/incomplete": 6, "/6": 6,
    "/hashing":    7, "/7": 7,
    "/seeding":    8, "/8": 8,
    "/leeching":   9, "/9": 9,
}

type CapturePane struct {
	snap [screens][]byte
}

var pane CapturePane

func snapshot(screen int) []byte {
	move := exec.Command("tmux", "send-keys", "-t", "rtorrent", strconv.Itoa(screen))
	_, err := move.Output()
	if err != nil {
		fmt.Println(err)
		return []byte{}
	}

	time.Sleep(50 * time.Millisecond)

	cmd := exec.Command("tmux", "capture-pane", "-t", "rtorrent", "-p")
	stdout, err := cmd.Output()
	if err != nil {
		fmt.Println(err)
		return []byte{}
	}
	return stdout
}

func update() {
	for i := 0; i < screens; i++ {
		pane.snap[i] = snapshot(i)
	}
}

func handleRequest(w http.ResponseWriter, req *http.Request) {
	index := pageTable[req.URL.Path]
	w.Write(pane.snap[index])
}

func main() {
	ticker := time.NewTicker(5 * time.Second)
	quit := make(chan struct{})
	go func() {
		for {
			select {
				case <- ticker.C:
					update()
				case <- quit:
					ticker.Stop()
					return
			}
		}
	}()

	http.HandleFunc("/", handleRequest)
	log.Fatal(http.ListenAndServe(":4002", nil))
}
