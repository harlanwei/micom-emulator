package main

import (
	"fmt"

	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"
)

const NOT_IMPL_STRING = " [-]([red]not implemented[-])"
const BASE_BRIGHTNESS = 60
const BRIGHTNESS_COEEFFICENT = 255 - BASE_BRIGHTNESS

type HuState struct {
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

	TextAlignment int
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

func (state *HuState) GetTextViewColor() tcell.Color {
	if !state.IsHuOn {
		return tcell.Color255
	}

	b := int32(
		BASE_BRIGHTNESS + state.ScreenBrightness*BRIGHTNESS_COEEFFICENT/100)
	return tcell.NewRGBColor(b, b, b)
}

func (state *HuState) ToggleMute() {
	state.IsMuted = !state.IsMuted
}

func (state *HuState) LowerVolume() {
	state.RadioVolume = computeBoundedValue(state.RadioVolume, true)
}

func (state *HuState) LowerBrightness() {
	state.ScreenBrightness = computeBoundedValue(
		state.ScreenBrightness, true)
}

func (state *HuState) IncreaseBrightness() {
	state.ScreenBrightness = computeBoundedValue(
		state.ScreenBrightness, false)
}

func (state *HuState) GetEffectiveVolume() int {
	if state.IsMuted {
		return 0
	}
	return state.RadioVolume
}

func (state *HuState) GetFuelMessage() string {
	if state.IsFuelLow {
		return "\n[red]WARNING: LOW FUEL[-]"
	}
	return ""
}

func (state *HuState) GetNavigationAddress() (ret string) {
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

func (state *HuState) GetRadioMessage() string {
	if !state.ShouldShowRadioMessage {
		return ""
	}
	return "\n💽 Playing: Y.M.C.A by Village People (1:23/3:49)"
}

func (state *HuState) ToString() string {
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

func (state *HuState) Update(code uint64) {
	eventAppendix := ""

	switch code {
	case TOGGLE_RADIO_MUTE:
		state.IsMuted = !state.IsMuted
	case REDUCE_RADIO_VOLUME:
		state.LowerVolume()
	case MAX_RADIO_VOLUME:
		state.RadioVolume = 100
	case LOW_SCREEN_BRIGHTNESS:
		state.LowerBrightness()
	case HIGH_SCREEN_BRIGHTNESS:
		state.IncreaseBrightness()
	case LOW_FUEL_WARNING:
		state.IsFuelLow = true
	case NAVIGATION_FULL_SCREEN:
		eventAppendix = NOT_IMPL_STRING
	case SET_NAVIGATION_ADDRESS:
		state.NavigationAddrInd = 1
	case SEEK_DOWN_SEARCH:
		eventAppendix = NOT_IMPL_STRING
	case SEEK_UP_SEARCH:
		eventAppendix = NOT_IMPL_STRING
	case SWITCH_ON_HU:
		state.TextAlignment = tview.AlignLeft
		state.IsHuOn = true
	case SWITCH_OFF_HU:
		state.TextAlignment = tview.AlignCenter
		state.IsHuOn = false
	case CAMERA_REVERSE_ON:
		state.IsCameraReverseOn = true
	case CAMERA_REVERSE_OFF:
		state.IsCameraReverseOn = false
	case TOGGLE_CHANGE_LANGUAGE:
		eventAppendix = NOT_IMPL_STRING
	case TOGGLE_SPEED_LIMIT:
		state.HasSpeedLimit = !state.HasSpeedLimit
	case TOGGLE_ROUNDABOUT_FARAWAY:
		state.ShouldShowRoundaboutDistance = !state.ShouldShowRoundaboutDistance
	case TOGGLE_RANDOM_NAVIGATION:
		eventAppendix = NOT_IMPL_STRING
	case TOGGLE_RADIO_INFO:
		state.ShouldShowRadioMessage = true
		eventAppendix = NOT_IMPL_STRING
	case INJECT_SCENE:
		// do nothing
	}

	if event, ok := CodeEventMap[code]; ok {
		state.Event = event + eventAppendix
	} else {
		state.Event = "(unknown event)"
	}
}
