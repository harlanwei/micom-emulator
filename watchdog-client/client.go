// Vian Chen <imvianchen@stu.pku.edu.cn>

package main

import (
	"encoding/binary"
	"fmt"
	"log"
	"os"
	"time"

	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"
	"golang.org/x/sys/unix"
)

const NOT_IMPL_STRING = " [-]([red]not implemented[-])"
const BASE_BRIGHTNESS = 60
const BRIGHTNESS_COEEFFICENT = 255 - BASE_BRIGHTNESS

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

func getTextViewColor(state *hustate) tcell.Color {
	if !state.IsHuOn {
		return tcell.Color255
	}

	b := int32(
		BASE_BRIGHTNESS + state.ScreenBrightness*BRIGHTNESS_COEEFFICENT/100)
	return tcell.NewRGBColor(b, b, b)
}

func (state *hustate) ToggleMute() {
	state.IsMuted = !state.IsMuted
}

func (state *hustate) LowerVolume() {
	state.RadioVolume = computeBoundedValue(state.RadioVolume, true)
}

func (state *hustate) LowerBrightness() {
	state.ScreenBrightness = computeBoundedValue(
		state.ScreenBrightness, true)
}

func (state *hustate) IncreaseBrightness() {
	state.ScreenBrightness = computeBoundedValue(
		state.ScreenBrightness, false)
}

func (state *hustate) GetEffectiveVolume() int {
	if state.IsMuted {
		return 0
	}
	return state.RadioVolume
}

func (state *hustate) GetFuelMessage() string {
	if state.IsFuelLow {
		return "\n[red]WARNING: LOW FUEL[-]"
	}
	return ""
}

func (state *hustate) GetNavigationAddress() (ret string) {
	if state.NavigationAddrInd == 0 {
		if state.ShouldShowRoundaboutDistance {
			ret = "\n[red]No destination set[-]"
		}
		return
	}

	ret = `
🏠 Navigating to:
	PKU, Yiheyuan Rd 5, Haidian District, Beijing`

	if state.ShouldShowRoundaboutDistance {
		ret = ret + "\n\t[green]Turn right in 5 km[-]"
	}

	return
}

func (state *hustate) GetRadioMessage() string {
	if !state.ShouldShowRadioMessage {
		return ""
	}
	return "\n💽 Playing: Y.M.C.A by Village People (1:23/3:49)"
}

func (state *hustate) ToString() string {
	if !state.IsHuOn {
		return "\n\n\n\n[black:white:bl]HEAD UNIT TURNED OFF[-:-:-]"
	}

	return fmt.Sprintf(`Volume: %d
Brightness: %d
Reverse cam: %s
Speed limit: %s
%s%s%s
`,

		state.GetEffectiveVolume(),
		state.ScreenBrightness,
		getOnOffString(state.IsCameraReverseOn),
		getOnOffString(state.HasSpeedLimit),

		state.GetFuelMessage(),
		state.GetNavigationAddress(),
		state.GetRadioMessage(),
	)
}

func (state *hustate) Update(
	code uint64,
	stateView *tview.TextView,
	historyView *tview.TextView,
) {
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
		state.Event = "navigation_full_screen" + NOT_IMPL_STRING
	case 8:
		state.NavigationAddrInd = 1
		state.Event = "set_navigation_address"
	case 9:
		state.Event = "seek_down_search" + NOT_IMPL_STRING
	case 10:
		state.Event = "seek_up_search" + NOT_IMPL_STRING
	case 11:
		stateView.SetTextAlign(tview.AlignLeft)
		state.IsHuOn = true
		state.Event = "switch_on_hu"
	case 12:
		stateView.SetTextAlign(tview.AlignCenter)
		state.IsHuOn = false
		state.Event = "switch_off_hu"
	case 13:
		state.IsCameraReverseOn = true
		state.Event = "camera_reverse_on"
	case 14:
		state.Event = "camera_reverse_off"
		state.IsCameraReverseOn = false
	case 15:
		state.Event = "toggle_change_language" + NOT_IMPL_STRING
	case 16:
		state.HasSpeedLimit = !state.HasSpeedLimit
		state.Event = "toggle_speed_limit"
	case 17:
		state.ShouldShowRoundaboutDistance = !state.ShouldShowRoundaboutDistance
		state.Event = "toggle_roundabout_faraway"
	case 18:
		state.Event = "toggle_random_navigation" + NOT_IMPL_STRING
	case 19:
		// FIXME
		state.ShouldShowRadioMessage = true
		state.Event = "toggle_radio_info"
	default:
		state.Event = "unknown_event"
	}

	stateView.
		SetText(state.ToString()).
		SetTextColor(getTextViewColor(state))

	t := time.Now()
	fmt.Fprintf(historyView,
		"%d:%02d:%02d.%03d [yellow]%s[-]\n",
		t.Hour(), t.Minute(), t.Second(), t.Nanosecond()/1_000_000,
		state.Event)
}

func main() {
	mobileState := hustate{
		RadioVolume:      50,
		ScreenBrightness: 50,
		IsHuOn:           true,
	}

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
	unix.Close(micomfd)

	// Create app
	app := tview.NewApplication()
	header := tview.NewTextView().
		SetTextAlign(tview.AlignCenter).
		SetText("Head Unit Emulator")
	stateView := tview.NewTextView().
		SetDynamicColors(true).
		SetRegions(true).
		SetText(mobileState.ToString()).
		SetTextColor(getTextViewColor(&mobileState))
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

			command := binary.LittleEndian.Uint64(ctr[:])
			mobileState.Update(command, stateView, historyView)
			app.Draw()
		}
	}()

	// Kick off the app
	app.SetRoot(grid, true).SetFocus(grid).Run()
}
