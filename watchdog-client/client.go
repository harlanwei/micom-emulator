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
	case TOGGLE_RADIO_MUTE:
		state.IsMuted = !state.IsMuted
		state.Event = "toggle_radio_mute"
	case REDUCE_RADIO_VOLUME:
		state.LowerVolume()
		state.Event = "reduce_radio_volume"
	case MAX_RADIO_VOLUME:
		state.RadioVolume = 100
		state.Event = "max_radio_volume"
	case LOW_SCREEN_BRIGHTNESS:
		state.LowerBrightness()
		state.Event = "low_screen_brightness"
	case HIGH_SCREEN_BRIGHTNESS:
		state.IncreaseBrightness()
		state.Event = "high_screen_brightness"
	case LOW_FUEL_WARNING:
		state.IsFuelLow = true
		state.Event = "low_fuel_warning"
	case NAVIGATION_FULL_SCREEN:
		state.Event = "navigation_full_screen" + NOT_IMPL_STRING
	case SET_NAVIGATION_ADDRESS:
		state.NavigationAddrInd = 1
		state.Event = "set_navigation_address"
	case SEEK_DOWN_SEARCH:
		state.Event = "seek_down_search" + NOT_IMPL_STRING
	case SEEK_UP_SEARCH:
		state.Event = "seek_up_search" + NOT_IMPL_STRING
	case SWITCH_ON_HU:
		stateView.SetTextAlign(tview.AlignLeft)
		state.IsHuOn = true
		state.Event = "switch_on_hu"
	case SWITCH_OFF_HU:
		stateView.SetTextAlign(tview.AlignCenter)
		state.IsHuOn = false
		state.Event = "switch_off_hu"
	case CAMERA_REVERSE_ON:
		state.IsCameraReverseOn = true
		state.Event = "camera_reverse_on"
	case CAMERA_REVERSE_OFF:
		state.Event = "camera_reverse_off"
		state.IsCameraReverseOn = false
	case TOGGLE_CHANGE_LANGUAGE:
		state.Event = "toggle_change_language" + NOT_IMPL_STRING
	case TOGGLE_SPEED_LIMIT:
		state.HasSpeedLimit = !state.HasSpeedLimit
		state.Event = "toggle_speed_limit"
	case TOGGLE_ROUNDABOUT_FARAWAY:
		state.ShouldShowRoundaboutDistance = !state.ShouldShowRoundaboutDistance
		state.Event = "toggle_roundabout_faraway"
	case TOGGLE_RANDOM_NAVIGATION:
		state.Event = "toggle_random_navigation" + NOT_IMPL_STRING
	case TOGGLE_RADIO_INFO:
		// FIXME
		state.ShouldShowRadioMessage = true
		state.Event = "toggle_radio_info"
	case INJECT_SCENE:
		state.Event = "inject_scene"
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
	defer unix.Close(micomfd)

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
