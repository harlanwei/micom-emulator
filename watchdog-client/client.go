// Vian Chen <imvianchen@stu.pku.edu.cn>

package main

import (
	"encoding/binary"
	"fmt"
	"log"
	"os"
	"strings"
	"time"

	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"
	"golang.org/x/sys/unix"
)

type Hu struct {
	state *HuState
	mgr   *SceneManager
}

func InitHu() Hu {
	mgr := InitSceneManager()

	return Hu{
		state: &HuState{
			RadioVolume:      50,
			ScreenBrightness: 50,
			IsHuOn:           true,
			TextAlignment:    tview.AlignLeft,
		},
		mgr: &mgr,
	}
}

func (hu *Hu) ToString() string {
	return hu.state.ToString()
}

func (hu *Hu) GetTextColor() tcell.Color {
	return hu.state.GetTextViewColor()
}

func (hu *Hu) ProcessCommand(code uint64) *Hu {
	if code == INJECT_SCENE {
		fd, err := os.OpenFile("scene.tmp", os.O_RDONLY, 0)
		if err != nil {
			log.Panicln("No scene.tmp file.")
		}

		buf := make([]byte, 1024)
		_, err = fd.Read(buf)
		if err != nil {
			log.Panicln("Cannot read from scene.tmp")
		}

		fd.Close()
		content := string(buf[:])
		keyValue := strings.Split(content, " ")
		if len(keyValue) != 2 {
			log.Panicln("Incorrect format of scene.tmp")
		}

		key, value := keyValue[0], keyValue[1]
		hu.mgr.InjectScene(key, value)
	}
	hu.state.Update(code)
	return hu
}

func (hu *Hu) FlushToViews(
	stateView *tview.TextView,
	historyView *tview.TextView,
) {
	state, mgr := hu.state, hu.mgr
	text := state.ToString() + "\n" + mgr.ToString()

	stateView.
		SetText(text).
		SetTextColor(state.GetTextViewColor()).
		SetTextAlign(state.TextAlignment)

	t := time.Now()
	fmt.Fprintf(historyView,
		"%d:%02d:%02d.%03d [yellow]%s[-]\n",
		t.Hour(), t.Minute(), t.Second(), t.Nanosecond()/1_000_000,
		state.Event)
}

func main() {
	hu := InitHu()

	// Set up eventfd
	errOutput := log.New(os.Stderr, "watchdog-client: ", 0)

	efd, err := unix.Eventfd(0, 0)
	if err != nil {
		errOutput.Panicln("failed to create an eventfd")
	}

	micomfd, err := unix.Open("/dev/micom", unix.O_WRONLY, 0)
	if err != nil {
		errOutput.Panicln("failed to open /dev/micom")
	}

	unix.IoctlSetInt(micomfd, 267520, efd)
	defer unix.Close(micomfd)

	// Create app
	app := tview.NewApplication()
	header := tview.NewTextView().
		SetTextAlign(tview.AlignCenter).
		SetText("Head Unit Emulator")
	stateView := tview.NewTextView().
		SetDynamicColors(true).
		SetRegions(true).
		SetText(hu.ToString()).
		SetTextColor(hu.GetTextColor())
	historyHeader := tview.NewTextView().
		SetTextAlign(tview.AlignCenter).
		SetText("Event Histories")
	historyView := tview.NewTextView().
		SetDynamicColors(true).
		SetRegions(true).
		SetScrollable(true)
	grid := tview.NewGrid().
		SetBorders(true).
		SetRows(1, 1, 0).
		SetColumns(-2, -1).
		AddItem(header, 0, 0, 1, 2, 0, 0, false).
		AddItem(stateView, 1, 0, 2, 1, 0, 0, false).
		AddItem(historyHeader, 1, 1, 1, 1, 0, 0, false).
		AddItem(historyView, 2, 1, 1, 1, 0, 0, true)

	// Set up eventfd callback
	ctr := [8]byte{0}
	go func() {
		for {
			_, err = unix.Read(efd, ctr[:])
			if err != nil {
				if err.Error() != "EOF" {
					fmt.Fprintln(historyView,
						"[::bl]⚠️ FATAL!!! unix.Read failed[::-]")
				} else {
					continue
				}
			}

			rep := binary.LittleEndian.Uint64(ctr[:])
			hu.ProcessCommand(rep).FlushToViews(stateView, historyView)
			app.Draw()
		}
	}()

	// Kick off the app
	app.SetRoot(grid, true).SetFocus(grid).Run()
}
