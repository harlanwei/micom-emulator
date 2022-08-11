package main

import (
	"encoding/binary"
	"fmt"
	"log"
	"os"

	"github.com/rivo/tview"
	"golang.org/x/sys/unix"
)

type hustate struct {
	IsMuted           bool
	RadioVolume       int
	ScreenBrightness  int
	NavigationAddrInd int
	IsHuOn            bool
	IsCameraReverseOn bool
	IsFuelLow         bool
	Event             string
}

func computeBoundedValue(value int, isNegative bool) int {
	if isNegative {
		if value <= 10 {
			return 0
		}
		return value - 10
	} else {
		if value >= 10 {
			return 0
		}
		return value + 10
	}
}

func getOnOffString(value bool) string {
	if value {
		return "ON"
	}
	return "OFF"
}

func (state *hustate) ToggleMute() {
	state.IsMuted = !state.IsMuted
}

func (state *hustate) LowerVolume() {
	state.RadioVolume = computeBoundedValue(state.RadioVolume, true)
}

func (state *hustate) IncreaseVolume() {
	state.RadioVolume = computeBoundedValue(state.RadioVolume, false)
}

func (state *hustate) LowerBrightness() {
	state.ScreenBrightness = computeBoundedValue(state.ScreenBrightness, true)
}

func (state *hustate) IncreaseBrightness() {
	state.ScreenBrightness = computeBoundedValue(state.ScreenBrightness, false)
}

func (state *hustate) GetEffectiveVolume() int {
	if state.IsMuted {
		return 0
	}
	return state.RadioVolume
}

func (state *hustate) GetFuelMessage() string {
	if state.IsFuelLow {
		return "[red]WARNING: LOW FUEL[white]"
	}
	return ""
}

func (state *hustate) ToString() string {
	if !state.IsHuOn {
		return "\n\n\n\n[yellow]HEAD UNIT TURNED OFF[white]"
	}

	return fmt.Sprintf(`Volume: %d
Brightness: %d
Reverse cam: %s

%s
`,
		state.GetEffectiveVolume(),
		state.ScreenBrightness,
		getOnOffString(state.IsCameraReverseOn),
		state.GetFuelMessage(),
	)
}

func (state *hustate) Update(code uint64, textView *tview.TextView, footer *tview.TextView) {
	switch code {
	case 1:
		state.ToggleMute()
		state.Event = "toggle_radio_mute"
	case 2:
		state.LowerVolume()
		state.Event = "reduce_radio_volume"
	case 3:
		state.IncreaseVolume()
		state.Event = "max_radio_volume"
	case 4:
		state.LowerBrightness()
		state.Event = "low_screen_brightness"
	case 5:
		state.IncreaseBrightness()
		state.Event = "high_screen_brightness"
	case 6:
		state.IsFuelLow = true
		state.Event = "low_fuel_warning"
	case 7:
		// navigation_full_screen
		state.Event = "navigation_full_screen"
	case 8:
		// set_navigation_address
		state.Event = "set_navigation_address"
	case 9:
		// seek_down_search
		state.Event = "seek_down_search"
	case 10:
		// seek_up_search
		state.Event = "seek_up_search"
	case 11:
		textView.SetTextAlign(tview.AlignLeft)
		state.IsHuOn = true
		state.Event = "switch_on_hu"
	case 12:
		textView.SetTextAlign(tview.AlignCenter)
		state.IsHuOn = false
		state.Event = "switch_off_hu"
	case 13:
		state.IsCameraReverseOn = true
		state.Event = "camera_reverse_on"
	case 14:
		state.Event = "camera_reverse_off"
		state.IsCameraReverseOn = false
	case 15:
		// cluster_change_language
		state.Event = "cluster_change_language"
	case 16:
		// cluster_speed_limit
		state.Event = "cluster_speed_limit"
	case 17:
		// cluster_roundabout_faraway
		state.Event = "cluster_roundabout_faraway"
	case 18:
		// cluster_random_navigation
		state.Event = "cluster_random_navigation"
	case 19:
		// cluster_radio_info
		state.Event = "cluster_radio_info"
	case 20:
		// inject_custom
		state.Event = "inject_custom"
	}

	textView.Clear()
	fmt.Fprintf(textView, "%s", state.ToString())
	footer.SetText(fmt.Sprintf("Recent event: %s", state.Event))
}

func main() {
	mobileState := hustate{
		IsMuted:           false,
		RadioVolume:       50,
		ScreenBrightness:  50,
		NavigationAddrInd: 0,
		IsHuOn:            true,
		IsCameraReverseOn: false,
		Event:             "none",
	}

	// Set up eventfd
	errOutput := log.New(os.Stderr, "watchdog-client: ", 0)

	efd, err := unix.Eventfd(0, 0)
	if err != nil {
		errOutput.Println("failed to create an eventfd")
		os.Exit(-1)
	}

	micomfd, err := unix.Open("/dev/micom", unix.O_WRONLY, 0)
	if err != nil {
		errOutput.Println("failed to open /dev/micom")
		os.Exit(-1)
	}

	rfds := unix.FdSet{}
	rfds.Zero()
	rfds.Set(efd)
	unix.IoctlSetInt(micomfd, 267520, efd)
	unix.Close(micomfd)

	// Create app
	app := tview.NewApplication()
	textView := tview.NewTextView()
	textView.SetDynamicColors(true).
		SetRegions(true)
	fmt.Fprintf(textView, "%s", mobileState.ToString())
	header := tview.NewTextView().
		SetTextAlign(tview.AlignCenter).
		SetText("Head Unit Emulator")
	footer := tview.NewTextView().
		SetText(fmt.Sprintf("Recent event: %s", mobileState.Event))
	grid := tview.NewGrid().
		SetRows(1, 0, 1).
		SetColumns(0).
		AddItem(header, 0, 0, 1, 1, 0, 0, false).
		AddItem(textView, 1, 0, 1, 1, 0, 0, true).
		AddItem(footer, 2, 0, 1, 1, 0, 0, false)
	grid.SetBorders(true)

	// Set up eventfd callback
	ctr := [8]byte{0}
	go func() {
		for {
			_, err = unix.Read(efd, ctr[:])
			if err != nil {
				if err.Error() != "EOF" {
					fmt.Fprintln(textView, "unix.Read failed")
				} else {
					continue
				}
			}

			command := binary.LittleEndian.Uint64(ctr[:])
			mobileState.Update(command, textView, footer)
			app.Draw()
		}
	}()

	app.SetRoot(grid, true).SetFocus(grid).Run()
}
