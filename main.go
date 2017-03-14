package main

import (
	"flag"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"strings"

	"github.com/jroimartin/gocui"
	"github.com/satyrius/gonx"
)

var (
	format  string
	host    string
	logFile string

	done     = make(chan struct{})
	messages = make(chan string)
	out      = make(chan string, 10)
)

func init() {
	flag.StringVar(&host, "host", "", "Remote host to talk with.")
	flag.StringVar(&format, "format", `$remote_addr - $remote_user [$time_local] "$request" $status $body_bytes_sent "$http_referer" "$http_user_agent"`, "Log format (default is 'combined' nginx format)")
	flag.StringVar(&logFile, "log", "-", "Log file name to read. Defaults to stdin.")

}

func main() {
	flag.Parse()

	g, err := gocui.NewGui(gocui.OutputNormal)
	if err != nil {
		log.Panicln(err)
	}
	defer g.Close()

	g.Cursor = true

	g.SetManagerFunc(layout)

	if err := keybindings(g); err != nil {
		log.Panicln(err)
	}

	go func() {
		var logReader io.Reader

		if logFile == "-" {
			logReader = os.Stdin
		} else {
			file, err := os.Open(logFile)
			if err != nil {
				panic(err)
			}
			logReader = file
			defer file.Close()
		}

		reader := gonx.NewReader(logReader, format)
		for {
			rec, err := reader.Read()
			if err != nil {
				return
			}
			msg, _ := rec.Field("http_user_agent")
			ip, _ := rec.Field("remote_addr")
			messages <- fmt.Sprintf("%s: %s\n", ip, msg)
		}
	}()

	go func() {
		for {
			select {
			case <-done:
				return
			case msg := <-messages:
				g.Execute(func(g *gocui.Gui) error {
					v, err := g.View("messages")
					if err != nil {
						return err
					}
					fmt.Fprint(v, msg)
					return nil
				})
			}
		}
	}()

	go func() {
		client := &http.Client{
			CheckRedirect: func(req *http.Request, via []*http.Request) error {
				return http.ErrUseLastResponse
			},
		}
		for msg := range out {
			req, err := http.NewRequest("GET", fmt.Sprintf("http://%s/", host), nil)
			if err != nil {
				continue
			}
			req.Header.Set("User-Agent", msg)
			resp, err := client.Do(req)
			if err != nil {
				continue
			}
			resp.Body.Close()
		}
	}()

	if err := g.MainLoop(); err != nil && err != gocui.ErrQuit {
		log.Panicln(err)
	}
}

func layout(g *gocui.Gui) error {
	maxX, maxY := g.Size()
	if v, err := g.SetView("input", 0, maxY-3, maxX, maxY); err != nil {
		if err != gocui.ErrUnknownView {
			return err
		}
		v.Frame = false
		v.Editable = true
		if _, err := g.SetCurrentView("input"); err != nil {
			return err
		}
	}
	if v, err := g.SetView("messages", -1, -1, maxX, maxY-3); err != nil {
		if err != gocui.ErrUnknownView {
			return err
		}
		v.Frame = false
	}
	return nil
}

func keybindings(g *gocui.Gui) error {
	if err := g.SetKeybinding("", gocui.KeyCtrlC, gocui.ModNone, quit); err != nil {
		return err
	}

	if err := g.SetKeybinding("", gocui.KeyEnter, gocui.ModNone, send); err != nil {
		return err
	}
	return nil
}

func send(g *gocui.Gui, v *gocui.View) error {
	msg, err := v.Line(0)
	msg = strings.Replace(msg, "\x00", "", 1)
	v.Clear()
	v.SetCursor(0, 0)
	if err != nil {
		return err
	}
	out <- msg
	return nil
}

func quit(g *gocui.Gui, v *gocui.View) error {
	close(done)
	return gocui.ErrQuit
}
