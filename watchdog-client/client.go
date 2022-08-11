package main

import (
	"encoding/binary"
	"fmt"
	"log"
	"os"
	"time"

	"github.com/rivo/tview"
	"golang.org/x/sys/unix"
)

type hustate struct {
	IsMuted                      bool
	RadioVolume                  int
	ScreenBrightness             int
	NavigationAddrInd            int
	IsHuOn                       bool
	IsCameraReverseOn            bool
	IsFuelLow                    bool
	Event                        string
	HasSpeedLimit                bool
	ShouldShowRoundaboutDistance bool
	ShouldShowRadioMessage       bool
}

func computeBoundedValue(value int, isNegative bool) int {
	if isNegative {
		if value <= 10 {
			return 0
		}
		return value - 10
	} else {
		if value >= 90 {
			return 100
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

func (state *hustate) GetNavigationAddress() (ret string) {
	if state.NavigationAddrInd == 0 {
		if state.ShouldShowRoundaboutDistance {
			ret = "[red]No destination set[white]"
		}
		return
	}

	ret = `Navigating to:
	PKU, Yiheyuan Rd 5, Haidian District, Beijing`

	if state.ShouldShowRoundaboutDistance {
		ret = ret + "\n\t[green]Turn right in 5 km[white]"
	}

	return
}

func (state *hustate) GetRadioMessage() (ret string) {
	if !state.ShouldShowRadioMessage {
		return
	}
	return "Playing: Y.M.C.A by Village People [1:23/3:49]"
}

func (state *hustate) ToString() string {
	if !state.IsHuOn {
		return "\n\n\n\n[yellow]HEAD UNIT TURNED OFF[white]"
	}

	return fmt.Sprintf(`Volume: %d
Brightness: %d
Reverse cam: %s
Speed limit: %s

%s
%s
%s
`,
		state.GetEffectiveVolume(),
		state.ScreenBrightness,
		getOnOffString(state.IsCameraReverseOn),
		getOnOffString(state.HasSpeedLimit),
		state.GetNavigationAddress(),
		state.GetFuelMessage(),
		state.GetRadioMessage(),
	)
}

func (state *hustate) Update(code uint64, textView *tview.TextView, footer *tview.TextView) {
	switch code {
	case 1:
		state.IsMuted = !state.IsMuted
		state.Event = "toggle_radio_mute"
	case 2:
		state.LowerVolume()
		state.Event = "reduce_radio_volume"
	case 3:
		state.RadioVolume = 100
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
		state.Event = "navigation_full_screen [white]([red] not implemented[white])"
	case 8:
		state.NavigationAddrInd = 1
		state.Event = "set_navigation_address"
	case 9:
		// seek_down_search
		state.Event = "seek_down_search [white]([red] not implemented[white])"
	case 10:
		// seek_up_search
		state.Event = "seek_up_search [white]([red] not implemented[white])"
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
		state.Event = "cluster_change_language [white]([red] not implemented[white])"
	case 16:
		state.HasSpeedLimit = !state.HasSpeedLimit
		state.Event = "cluster_speed_limit"
	case 17:
		state.ShouldShowRoundaboutDistance = !state.ShouldShowRoundaboutDistance
		state.Event = "cluster_roundabout_faraway"
	case 18:
		// cluster_random_navigation
		state.Event = "cluster_random_navigation [white]([red] not implemented[white])"
	case 19:
		state.ShouldShowRadioMessage = !state.ShouldShowRadioMessage
		state.Event = "cluster_radio_info"
	default:
		state.Event = "unknown_event"
	}

	textView.Clear()
	fmt.Fprintf(textView, "%s", state.ToString())

	t := time.Now()
	fmt.Fprintf(footer, "%d:%02d:%02d.%03d [yellow]%s[white]\n",
		t.Hour(), t.Minute(), t.Second(), t.Nanosecond()/1_000_000, state.Event)
}

func main() {
	mobileState := hustate{
		RadioVolume:      50,
		ScreenBrightness: 50,
		IsHuOn:           true,
		Event:            "",
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
		SetRegions(true).
		SetText(mobileState.ToString())
	header := tview.NewTextView().
		SetTextAlign(tview.AlignCenter).
		SetText("Head Unit Emulator")
	footer := tview.NewTextView().
		SetDynamicColors(true).
		SetRegions(true).
		SetText("Event Histories\n")
	grid := tview.NewGrid().
		SetRows(1, 0).
		SetColumns(-2, -1).
		AddItem(header, 0, 0, 1, 2, 0, 0, false).
		AddItem(textView, 1, 0, 1, 1, 0, 0, true).
		AddItem(footer, 1, 1, 1, 1, 0, 0, false)
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
