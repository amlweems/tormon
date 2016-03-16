package main

import "fmt"
import "log"
import "time"
import "net/http"
import "strconv"
import "os"
import "os/exec"
import "html/template"
import "io/ioutil"

type Page struct {
	Title   string
	Refresh int
	Body    []byte
}

var rate = 5
var pane []byte
var ticker *time.Ticker

func update() {
	tempFile, err := ioutil.TempFile("/tmp", "tormon")
	if err != nil {
		fmt.Println(err)
		return
	}
	defer tempFile.Close()
	defer os.Remove(tempFile.Name())

	resize := exec.Command("screen", "-X", "width", "-d", "120", "250")
	resize.Run()

	proc := exec.Command("screen", "-X", "hardcopy", tempFile.Name())
	proc.Run()

	// hardcopy is non-blocking, we'll just wait a bit
	time.Sleep(50 * time.Millisecond)

	pane, err = ioutil.ReadFile(tempFile.Name())
	if err != nil {
		fmt.Println(err)
		return
	}
}

func handleRequest(w http.ResponseWriter, req *http.Request) {
	title := req.URL.Path[1:]

	p := &Page{
		Title:   title,
		Refresh: rate,
		Body:    pane,
	}
	t, _ := template.ParseFiles("template.html")
	err := t.Execute(w, p)
	if err != nil {
		fmt.Println(err)
	}
}

func handleTicker(w http.ResponseWriter, req *http.Request) {
	path := req.URL.Path[len("/ticker/"):]
	rate, err := strconv.Atoi(path)
	if err != nil || rate < 0 || rate > 300 {
		rate = 5
	}
	ticker = time.NewTicker(time.Duration(rate) * time.Second)

	fmt.Fprintf(w, "Set refresh rate to %d seconds.", rate)
}

func main() {
	ticker = time.NewTicker(time.Duration(rate) * time.Second)
	update()
	go func() {
		for {
			select {
			case <-ticker.C:
				update()
			}
		}
	}()

	http.HandleFunc("/", handleRequest)
	http.HandleFunc("/ticker/", handleTicker)
	log.Fatal(http.ListenAndServe("127.0.0.1:4002", nil))
}
