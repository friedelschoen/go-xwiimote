package main

import (
	"bufio"
	"flag"
	"fmt"
	"io"
	"log"
	"os"
	"strings"

	"github.com/bendahl/uinput"
	"github.com/friedelschoen/go-xwiimote"
)

var wiiKeynames = map[string]xwiimote.Key{
	"LEFT":           xwiimote.KeyLeft,
	"RIGHT":          xwiimote.KeyRight,
	"UP":             xwiimote.KeyUp,
	"DOWN":           xwiimote.KeyDown,
	"A":              xwiimote.KeyA,
	"B":              xwiimote.KeyB,
	"PLUS":           xwiimote.KeyPlus,
	"MINUS":          xwiimote.KeyMinus,
	"HOME":           xwiimote.KeyHome,
	"ONE":            xwiimote.KeyOne,
	"TWO":            xwiimote.KeyTwo,
	"X":              xwiimote.KeyX,
	"Y":              xwiimote.KeyY,
	"TL":             xwiimote.KeyTL,
	"TR":             xwiimote.KeyTR,
	"ZL":             xwiimote.KeyZL,
	"ZR":             xwiimote.KeyZR,
	"THUMBL":         xwiimote.KeyThumbL,
	"THUMBR":         xwiimote.KeyThumbR,
	"C":              xwiimote.KeyC,
	"Z":              xwiimote.KeyZ,
	"STRUM_BAR_UP":   xwiimote.KeyStrumBarUp,
	"STRUM_BAR_DOWN": xwiimote.KeyStrumBarDown,
	"FRET_FAR_UP":    xwiimote.KeyFretFarUp,
	"FRET_UP":        xwiimote.KeyFretUp,
	"FRET_MID":       xwiimote.KeyFretMid,
	"FRET_LOW":       xwiimote.KeyFretLow,
	"FRET_FAR_LOW":   xwiimote.KeyFretFarLow,
}

var realKeynames = map[string]int{
	"KEY_ESC":              uinput.KeyEsc,
	"KEY_1":                uinput.Key1,
	"KEY_2":                uinput.Key2,
	"KEY_3":                uinput.Key3,
	"KEY_4":                uinput.Key4,
	"KEY_5":                uinput.Key5,
	"KEY_6":                uinput.Key6,
	"KEY_7":                uinput.Key7,
	"KEY_8":                uinput.Key8,
	"KEY_9":                uinput.Key9,
	"KEY_0":                uinput.Key0,
	"KEY_MINUS":            uinput.KeyMinus,
	"KEY_EQUAL":            uinput.KeyEqual,
	"KEY_BACKSPACE":        uinput.KeyBackspace,
	"KEY_TAB":              uinput.KeyTab,
	"KEY_Q":                uinput.KeyQ,
	"KEY_W":                uinput.KeyW,
	"KEY_E":                uinput.KeyE,
	"KEY_R":                uinput.KeyR,
	"KEY_T":                uinput.KeyT,
	"KEY_Y":                uinput.KeyY,
	"KEY_U":                uinput.KeyU,
	"KEY_I":                uinput.KeyI,
	"KEY_O":                uinput.KeyO,
	"KEY_P":                uinput.KeyP,
	"KEY_LEFTBRACE":        uinput.KeyLeftbrace,
	"KEY_RIGHTBRACE":       uinput.KeyRightbrace,
	"KEY_ENTER":            uinput.KeyEnter,
	"KEY_LEFTCTRL":         uinput.KeyLeftctrl,
	"KEY_A":                uinput.KeyA,
	"KEY_S":                uinput.KeyS,
	"KEY_D":                uinput.KeyD,
	"KEY_F":                uinput.KeyF,
	"KEY_G":                uinput.KeyG,
	"KEY_H":                uinput.KeyH,
	"KEY_J":                uinput.KeyJ,
	"KEY_K":                uinput.KeyK,
	"KEY_L":                uinput.KeyL,
	"KEY_SEMICOLON":        uinput.KeySemicolon,
	"KEY_APOSTROPHE":       uinput.KeyApostrophe,
	"KEY_GRAVE":            uinput.KeyGrave,
	"KEY_LEFTSHIFT":        uinput.KeyLeftshift,
	"KEY_BACKSLASH":        uinput.KeyBackslash,
	"KEY_Z":                uinput.KeyZ,
	"KEY_X":                uinput.KeyX,
	"KEY_C":                uinput.KeyC,
	"KEY_V":                uinput.KeyV,
	"KEY_B":                uinput.KeyB,
	"KEY_N":                uinput.KeyN,
	"KEY_M":                uinput.KeyM,
	"KEY_COMMA":            uinput.KeyComma,
	"KEY_DOT":              uinput.KeyDot,
	"KEY_SLASH":            uinput.KeySlash,
	"KEY_RIGHTSHIFT":       uinput.KeyRightshift,
	"KEY_KPASTERISK":       uinput.KeyKpasterisk,
	"KEY_LEFTALT":          uinput.KeyLeftalt,
	"KEY_SPACE":            uinput.KeySpace,
	"KEY_CAPSLOCK":         uinput.KeyCapslock,
	"KEY_F1":               uinput.KeyF1,
	"KEY_F2":               uinput.KeyF2,
	"KEY_F3":               uinput.KeyF3,
	"KEY_F4":               uinput.KeyF4,
	"KEY_F5":               uinput.KeyF5,
	"KEY_F6":               uinput.KeyF6,
	"KEY_F7":               uinput.KeyF7,
	"KEY_F8":               uinput.KeyF8,
	"KEY_F9":               uinput.KeyF9,
	"KEY_F10":              uinput.KeyF10,
	"KEY_NUMLOCK":          uinput.KeyNumlock,
	"KEY_SCROLLLOCK":       uinput.KeyScrolllock,
	"KEY_KP7":              uinput.KeyKp7,
	"KEY_KP8":              uinput.KeyKp8,
	"KEY_KP9":              uinput.KeyKp9,
	"KEY_KPMINUS":          uinput.KeyKpminus,
	"KEY_KP4":              uinput.KeyKp4,
	"KEY_KP5":              uinput.KeyKp5,
	"KEY_KP6":              uinput.KeyKp6,
	"KEY_KPPLUS":           uinput.KeyKpplus,
	"KEY_KP1":              uinput.KeyKp1,
	"KEY_KP2":              uinput.KeyKp2,
	"KEY_KP3":              uinput.KeyKp3,
	"KEY_KP0":              uinput.KeyKp0,
	"KEY_KPDOT":            uinput.KeyKpdot,
	"KEY_ZENKAKUHANKAKU":   uinput.KeyZenkakuhankaku,
	"KEY_102ND":            uinput.Key102Nd,
	"KEY_F11":              uinput.KeyF11,
	"KEY_F12":              uinput.KeyF12,
	"KEY_RO":               uinput.KeyRo,
	"KEY_KATAKANA":         uinput.KeyKatakana,
	"KEY_HIRAGANA":         uinput.KeyHiragana,
	"KEY_HENKAN":           uinput.KeyHenkan,
	"KEY_KATAKANAHIRAGANA": uinput.KeyKatakanahiragana,
	"KEY_MUHENKAN":         uinput.KeyMuhenkan,
	"KEY_KPJPCOMMA":        uinput.KeyKpjpcomma,
	"KEY_KPENTER":          uinput.KeyKpenter,
	"KEY_RIGHTCTRL":        uinput.KeyRightctrl,
	"KEY_KPSLASH":          uinput.KeyKpslash,
	"KEY_SYSRQ":            uinput.KeySysrq,
	"KEY_RIGHTALT":         uinput.KeyRightalt,
	"KEY_LINEFEED":         uinput.KeyLinefeed,
	"KEY_HOME":             uinput.KeyHome,
	"KEY_UP":               uinput.KeyUp,
	"KEY_PAGEUP":           uinput.KeyPageup,
	"KEY_LEFT":             uinput.KeyLeft,
	"KEY_RIGHT":            uinput.KeyRight,
	"KEY_END":              uinput.KeyEnd,
	"KEY_DOWN":             uinput.KeyDown,
	"KEY_PAGEDOWN":         uinput.KeyPagedown,
	"KEY_INSERT":           uinput.KeyInsert,
	"KEY_DELETE":           uinput.KeyDelete,
	"KEY_MACRO":            uinput.KeyMacro,
	"KEY_MUTE":             uinput.KeyMute,
	"KEY_VOLUMEDOWN":       uinput.KeyVolumedown,
	"KEY_VOLUMEUP":         uinput.KeyVolumeup,
	"KEY_POWER":            uinput.KeyPower,
	"KEY_KPEQUAL":          uinput.KeyKpequal,
	"KEY_KPPLUSMINUS":      uinput.KeyKpplusminus,
	"KEY_PAUSE":            uinput.KeyPause,
	"KEY_SCALE":            uinput.KeyScale,
	"KEY_KPCOMMA":          uinput.KeyKpcomma,
	"KEY_HANGEUL":          uinput.KeyHangeul,
	"KEY_HANJA":            uinput.KeyHanja,
	"KEY_YEN":              uinput.KeyYen,
	"KEY_LEFTMETA":         uinput.KeyLeftmeta,
	"KEY_RIGHTMETA":        uinput.KeyRightmeta,
	"KEY_COMPOSE":          uinput.KeyCompose,
	"KEY_STOP":             uinput.KeyStop,
	"KEY_AGAIN":            uinput.KeyAgain,
	"KEY_PROPS":            uinput.KeyProps,
	"KEY_UNDO":             uinput.KeyUndo,
	"KEY_FRONT":            uinput.KeyFront,
	"KEY_COPY":             uinput.KeyCopy,
	"KEY_OPEN":             uinput.KeyOpen,
	"KEY_PASTE":            uinput.KeyPaste,
	"KEY_FIND":             uinput.KeyFind,
	"KEY_CUT":              uinput.KeyCut,
	"KEY_HELP":             uinput.KeyHelp,
	"KEY_MENU":             uinput.KeyMenu,
	"KEY_CALC":             uinput.KeyCalc,
	"KEY_SETUP":            uinput.KeySetup,
	"KEY_SLEEP":            uinput.KeySleep,
	"KEY_WAKEUP":           uinput.KeyWakeup,
	"KEY_FILE":             uinput.KeyFile,
	"KEY_SENDFILE":         uinput.KeySendfile,
	"KEY_DELETEFILE":       uinput.KeyDeletefile,
	"KEY_XFER":             uinput.KeyXfer,
	"KEY_PROG1":            uinput.KeyProg1,
	"KEY_PROG2":            uinput.KeyProg2,
	"KEY_WWW":              uinput.KeyWww,
	"KEY_MSDOS":            uinput.KeyMsdos,
	"KEY_COFFEE":           uinput.KeyCoffee,
	"KEY_DIRECTION":        uinput.KeyDirection,
	"KEY_CYCLEWINDOWS":     uinput.KeyCyclewindows,
	"KEY_MAIL":             uinput.KeyMail,
	"KEY_BOOKMARKS":        uinput.KeyBookmarks,
	"KEY_COMPUTER":         uinput.KeyComputer,
	"KEY_BACK":             uinput.KeyBack,
	"KEY_FORWARD":          uinput.KeyForward,
	"KEY_CLOSECD":          uinput.KeyClosecd,
	"KEY_EJECTCD":          uinput.KeyEjectcd,
	"KEY_EJECTCLOSECD":     uinput.KeyEjectclosecd,
	"KEY_NEXTSONG":         uinput.KeyNextsong,
	"KEY_PLAYPAUSE":        uinput.KeyPlaypause,
	"KEY_PREVIOUSSONG":     uinput.KeyPrevioussong,
	"KEY_STOPCD":           uinput.KeyStopcd,
	"KEY_RECORD":           uinput.KeyRecord,
	"KEY_REWIND":           uinput.KeyRewind,
	"KEY_PHONE":            uinput.KeyPhone,
	"KEY_ISO":              uinput.KeyIso,
	"KEY_CONFIG":           uinput.KeyConfig,
	"KEY_HOMEPAGE":         uinput.KeyHomepage,
	"KEY_REFRESH":          uinput.KeyRefresh,
	"KEY_EXIT":             uinput.KeyExit,
	"KEY_MOVE":             uinput.KeyMove,
	"KEY_EDIT":             uinput.KeyEdit,
	"KEY_SCROLLUP":         uinput.KeyScrollup,
	"KEY_SCROLLDOWN":       uinput.KeyScrolldown,
	"KEY_KPLEFTPAREN":      uinput.KeyKpleftparen,
	"KEY_KPRIGHTPAREN":     uinput.KeyKprightparen,
	"KEY_NEW":              uinput.KeyNew,
	"KEY_REDO":             uinput.KeyRedo,
	"KEY_F13":              uinput.KeyF13,
	"KEY_F14":              uinput.KeyF14,
	"KEY_F15":              uinput.KeyF15,
	"KEY_F16":              uinput.KeyF16,
	"KEY_F17":              uinput.KeyF17,
	"KEY_F18":              uinput.KeyF18,
	"KEY_F19":              uinput.KeyF19,
	"KEY_F20":              uinput.KeyF20,
	"KEY_F21":              uinput.KeyF21,
	"KEY_F22":              uinput.KeyF22,
	"KEY_F23":              uinput.KeyF23,
	"KEY_F24":              uinput.KeyF24,
	"KEY_PLAYCD":           uinput.KeyPlaycd,
	"KEY_PAUSECD":          uinput.KeyPausecd,
	"KEY_PROG3":            uinput.KeyProg3,
	"KEY_PROG4":            uinput.KeyProg4,
	"KEY_DASHBOARD":        uinput.KeyDashboard,
	"KEY_SUSPEND":          uinput.KeySuspend,
	"KEY_CLOSE":            uinput.KeyClose,
	"KEY_PLAY":             uinput.KeyPlay,
	"KEY_FASTFORWARD":      uinput.KeyFastforward,
	"KEY_BASSBOOST":        uinput.KeyBassboost,
	"KEY_PRINT":            uinput.KeyPrint,
	"KEY_HP":               uinput.KeyHp,
	"KEY_CAMERA":           uinput.KeyCamera,
	"KEY_SOUND":            uinput.KeySound,
	"KEY_QUESTION":         uinput.KeyQuestion,
	"KEY_EMAIL":            uinput.KeyEmail,
	"KEY_CHAT":             uinput.KeyChat,
	"KEY_SEARCH":           uinput.KeySearch,
	"KEY_CONNECT":          uinput.KeyConnect,
	"KEY_FINANCE":          uinput.KeyFinance,
	"KEY_SPORT":            uinput.KeySport,
	"KEY_SHOP":             uinput.KeyShop,
	"KEY_ALTERASE":         uinput.KeyAlterase,
	"KEY_CANCEL":           uinput.KeyCancel,
	"KEY_BRIGHTNESSDOWN":   uinput.KeyBrightnessdown,
	"KEY_BRIGHTNESSUP":     uinput.KeyBrightnessup,
	"KEY_MEDIA":            uinput.KeyMedia,
	"KEY_SWITCHVIDEOMODE":  uinput.KeySwitchvideomode,
	"KEY_KBDILLUMTOGGLE":   uinput.KeyKbdillumtoggle,
	"KEY_KBDILLUMDOWN":     uinput.KeyKbdillumdown,
	"KEY_KBDILLUMUP":       uinput.KeyKbdillumup,
	"KEY_SEND":             uinput.KeySend,
	"KEY_REPLY":            uinput.KeyReply,
	"KEY_FORWARDMAIL":      uinput.KeyForwardmail,
	"KEY_SAVE":             uinput.KeySave,
	"KEY_DOCUMENTS":        uinput.KeyDocuments,
	"KEY_BATTERY":          uinput.KeyBattery,
	"KEY_BLUETOOTH":        uinput.KeyBluetooth,
	"KEY_WLAN":             uinput.KeyWlan,
	"KEY_UWB":              uinput.KeyUwb,
	"KEY_UNKNOWN":          uinput.KeyUnknown,
	"KEY_VIDEONEXT":        uinput.KeyVideoNext,
	"KEY_VIDEOPREV":        uinput.KeyVideoPrev,
	"KEY_BRIGHTNESSCYCLE":  uinput.KeyBrightnessCycle,
	"KEY_BRIGHTNESSZERO":   uinput.KeyBrightnessZero,
	"KEY_DISPLAYOFF":       uinput.KeyDisplayOff,
	"KEY_WIMAX":            uinput.KeyWimax,
	"KEY_RFKILL":           uinput.KeyRfkill,
	"KEY_MICMUTE":          uinput.KeyMicmute,

	"BUTTON_GAMEPAD": uinput.ButtonGamepad,

	"BUTTON_SOUTH": uinput.ButtonSouth,
	"BUTTON_EAST":  uinput.ButtonEast,
	"BUTTON_NORTH": uinput.ButtonNorth,
	"BUTTON_WEST":  uinput.ButtonWest,

	"BUTTON_BUMPERLEFT":   uinput.ButtonBumperLeft,
	"BUTTON_BUMPERRIGHT":  uinput.ButtonBumperRight,
	"BUTTON_TRIGGERLEFT":  uinput.ButtonTriggerLeft,
	"BUTTON_TRIGGERRIGHT": uinput.ButtonTriggerRight,
	"BUTTON_THUMBLEFT":    uinput.ButtonThumbLeft,
	"BUTTON_THUMBRIGHT":   uinput.ButtonThumbRight,

	"BUTTON_SELECT": uinput.ButtonSelect,
	"BUTTON_START":  uinput.ButtonStart,

	"BUTTON_DPADUP":    uinput.ButtonDpadUp,
	"BUTTON_DPADDOWN":  uinput.ButtonDpadDown,
	"BUTTON_DPADLEFT":  uinput.ButtonDpadLeft,
	"BUTTON_DPADRIGHT": uinput.ButtonDpadRight,

	"BUTTON_MODE": uinput.ButtonMode,
}

var (
	kbname = flag.String("name", "xwiimote-virtual", "Name to use")
)

func loadMapping(r io.Reader) map[xwiimote.Key]int {
	mapping := make(map[xwiimote.Key]int)
	scan := bufio.NewScanner(r)
	for scan.Scan() {
		line := scan.Text()
		wiibuttonstr, realkeystr, ok := strings.Cut(line, "->")
		if !ok {
			fmt.Fprintf(os.Stderr, "error: missing delimiter: %s\n", line)
			continue
		}
		wiibutton, ok := wiiKeynames[strings.TrimSpace(wiibuttonstr)]
		if !ok {
			fmt.Fprintf(os.Stderr, "error: unknown button: %s\n", wiibuttonstr)
			continue
		}
		realkey, ok := realKeynames[strings.TrimSpace(realkeystr)]
		if !ok {
			fmt.Fprintf(os.Stderr, "error: unknown key: %s\n", realkeystr)
			continue
		}
		mapping[wiibutton] = realkey
	}
	return mapping
}

func watchDevice(path string, mapping map[xwiimote.Key]int) {
	dev, err := xwiimote.NewDevice(path)
	if err != nil {
		fmt.Fprintf(os.Stderr, "error: unable to get device: %s", err)
	}
	defer dev.Free()
	poll := xwiimote.NewPoller(dev)

	if err := dev.Open(xwiimote.InterfaceCore); err != nil {
		fmt.Fprintf(os.Stderr, "error: unable to open device: %s", err)
	}

	kb, err := uinput.CreateKeyboard("/dev/uinput", []byte(*kbname))
	if err != nil {
		panic(err)
	}
	defer kb.Close()

	for {
		ev, err := poll.WaitEvent(-1)
		if err != nil {
			log.Printf("unable to poll event: %v\n", err)
		}
		switch ev := ev.(type) {
		case *xwiimote.EventKey:
			realkey, ok := mapping[ev.Code]
			if !ok {
				continue
			}
			switch ev.State {
			case xwiimote.StatePressed:
				kb.KeyDown(realkey)
			case xwiimote.StateReleased:
				kb.KeyUp(realkey)
			}
		}
	}
}

func main() {
	flag.Parse()

	mapping := loadMapping(os.Stdin)

	monitor := xwiimote.NewMonitor(false)
	defer monitor.Free()

	poll := xwiimote.NewPoller(monitor)
	for {
		path, err := poll.WaitEvent(-1)
		if err != nil || path == "" {
			log.Printf("error while polling: %v\n", err)
			continue
		}
		watchDevice(path, mapping)
	}
}
