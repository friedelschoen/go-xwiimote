package main

import (
	"bufio"
	"flag"
	"fmt"
	"io"
	"log"
	"os"
	"strings"
	"time"

	"github.com/friedelschoen/go-xwiimote"
	"github.com/friedelschoen/go-xwiimote/pkg/vinput"
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

var realKeynames = map[string]vinput.Key{
	"KEY_ESC":              vinput.KeyEsc,
	"KEY_1":                vinput.Key1,
	"KEY_2":                vinput.Key2,
	"KEY_3":                vinput.Key3,
	"KEY_4":                vinput.Key4,
	"KEY_5":                vinput.Key5,
	"KEY_6":                vinput.Key6,
	"KEY_7":                vinput.Key7,
	"KEY_8":                vinput.Key8,
	"KEY_9":                vinput.Key9,
	"KEY_0":                vinput.Key0,
	"KEY_MINUS":            vinput.KeyMinus,
	"KEY_EQUAL":            vinput.KeyEqual,
	"KEY_BACKSPACE":        vinput.KeyBackspace,
	"KEY_TAB":              vinput.KeyTab,
	"KEY_Q":                vinput.KeyQ,
	"KEY_W":                vinput.KeyW,
	"KEY_E":                vinput.KeyE,
	"KEY_R":                vinput.KeyR,
	"KEY_T":                vinput.KeyT,
	"KEY_Y":                vinput.KeyY,
	"KEY_U":                vinput.KeyU,
	"KEY_I":                vinput.KeyI,
	"KEY_O":                vinput.KeyO,
	"KEY_P":                vinput.KeyP,
	"KEY_LEFTBRACE":        vinput.KeyLeftbrace,
	"KEY_RIGHTBRACE":       vinput.KeyRightbrace,
	"KEY_ENTER":            vinput.KeyEnter,
	"KEY_LEFTCTRL":         vinput.KeyLeftctrl,
	"KEY_A":                vinput.KeyA,
	"KEY_S":                vinput.KeyS,
	"KEY_D":                vinput.KeyD,
	"KEY_F":                vinput.KeyF,
	"KEY_G":                vinput.KeyG,
	"KEY_H":                vinput.KeyH,
	"KEY_J":                vinput.KeyJ,
	"KEY_K":                vinput.KeyK,
	"KEY_L":                vinput.KeyL,
	"KEY_SEMICOLON":        vinput.KeySemicolon,
	"KEY_APOSTROPHE":       vinput.KeyApostrophe,
	"KEY_GRAVE":            vinput.KeyGrave,
	"KEY_LEFTSHIFT":        vinput.KeyLeftshift,
	"KEY_BACKSLASH":        vinput.KeyBackslash,
	"KEY_Z":                vinput.KeyZ,
	"KEY_X":                vinput.KeyX,
	"KEY_C":                vinput.KeyC,
	"KEY_V":                vinput.KeyV,
	"KEY_B":                vinput.KeyB,
	"KEY_N":                vinput.KeyN,
	"KEY_M":                vinput.KeyM,
	"KEY_COMMA":            vinput.KeyComma,
	"KEY_DOT":              vinput.KeyDot,
	"KEY_SLASH":            vinput.KeySlash,
	"KEY_RIGHTSHIFT":       vinput.KeyRightshift,
	"KEY_KPASTERISK":       vinput.KeyKpasterisk,
	"KEY_LEFTALT":          vinput.KeyLeftalt,
	"KEY_SPACE":            vinput.KeySpace,
	"KEY_CAPSLOCK":         vinput.KeyCapslock,
	"KEY_F1":               vinput.KeyF1,
	"KEY_F2":               vinput.KeyF2,
	"KEY_F3":               vinput.KeyF3,
	"KEY_F4":               vinput.KeyF4,
	"KEY_F5":               vinput.KeyF5,
	"KEY_F6":               vinput.KeyF6,
	"KEY_F7":               vinput.KeyF7,
	"KEY_F8":               vinput.KeyF8,
	"KEY_F9":               vinput.KeyF9,
	"KEY_F10":              vinput.KeyF10,
	"KEY_NUMLOCK":          vinput.KeyNumlock,
	"KEY_SCROLLLOCK":       vinput.KeyScrolllock,
	"KEY_KP7":              vinput.KeyKp7,
	"KEY_KP8":              vinput.KeyKp8,
	"KEY_KP9":              vinput.KeyKp9,
	"KEY_KPMINUS":          vinput.KeyKpminus,
	"KEY_KP4":              vinput.KeyKp4,
	"KEY_KP5":              vinput.KeyKp5,
	"KEY_KP6":              vinput.KeyKp6,
	"KEY_KPPLUS":           vinput.KeyKpplus,
	"KEY_KP1":              vinput.KeyKp1,
	"KEY_KP2":              vinput.KeyKp2,
	"KEY_KP3":              vinput.KeyKp3,
	"KEY_KP0":              vinput.KeyKp0,
	"KEY_KPDOT":            vinput.KeyKpdot,
	"KEY_ZENKAKUHANKAKU":   vinput.KeyZenkakuhankaku,
	"KEY_102ND":            vinput.Key102nd,
	"KEY_F11":              vinput.KeyF11,
	"KEY_F12":              vinput.KeyF12,
	"KEY_RO":               vinput.KeyRo,
	"KEY_KATAKANA":         vinput.KeyKatakana,
	"KEY_HIRAGANA":         vinput.KeyHiragana,
	"KEY_HENKAN":           vinput.KeyHenkan,
	"KEY_KATAKANAHIRAGANA": vinput.KeyKatakanahiragana,
	"KEY_MUHENKAN":         vinput.KeyMuhenkan,
	"KEY_KPJPCOMMA":        vinput.KeyKpjpcomma,
	"KEY_KPENTER":          vinput.KeyKpenter,
	"KEY_RIGHTCTRL":        vinput.KeyRightctrl,
	"KEY_KPSLASH":          vinput.KeyKpslash,
	"KEY_SYSRQ":            vinput.KeySysrq,
	"KEY_RIGHTALT":         vinput.KeyRightalt,
	"KEY_LINEFEED":         vinput.KeyLinefeed,
	"KEY_HOME":             vinput.KeyHome,
	"KEY_UP":               vinput.KeyUp,
	"KEY_PAGEUP":           vinput.KeyPageup,
	"KEY_LEFT":             vinput.KeyLeft,
	"KEY_RIGHT":            vinput.KeyRight,
	"KEY_END":              vinput.KeyEnd,
	"KEY_DOWN":             vinput.KeyDown,
	"KEY_PAGEDOWN":         vinput.KeyPagedown,
	"KEY_INSERT":           vinput.KeyInsert,
	"KEY_DELETE":           vinput.KeyDelete,
	"KEY_MACRO":            vinput.KeyMacro,
	"KEY_MUTE":             vinput.KeyMute,
	"KEY_VOLUMEDOWN":       vinput.KeyVolumedown,
	"KEY_VOLUMEUP":         vinput.KeyVolumeup,
	"KEY_POWER":            vinput.KeyPower,
	"KEY_KPEQUAL":          vinput.KeyKpequal,
	"KEY_KPPLUSMINUS":      vinput.KeyKpplusminus,
	"KEY_PAUSE":            vinput.KeyPause,
	"KEY_SCALE":            vinput.KeyScale,
	"KEY_KPCOMMA":          vinput.KeyKpcomma,
	"KEY_HANGEUL":          vinput.KeyHangeul,
	"KEY_HANJA":            vinput.KeyHanja,
	"KEY_YEN":              vinput.KeyYen,
	"KEY_LEFTMETA":         vinput.KeyLeftmeta,
	"KEY_RIGHTMETA":        vinput.KeyRightmeta,
	"KEY_COMPOSE":          vinput.KeyCompose,
	"KEY_STOP":             vinput.KeyStop,
	"KEY_AGAIN":            vinput.KeyAgain,
	"KEY_PROPS":            vinput.KeyProps,
	"KEY_UNDO":             vinput.KeyUndo,
	"KEY_FRONT":            vinput.KeyFront,
	"KEY_COPY":             vinput.KeyCopy,
	"KEY_OPEN":             vinput.KeyOpen,
	"KEY_PASTE":            vinput.KeyPaste,
	"KEY_FIND":             vinput.KeyFind,
	"KEY_CUT":              vinput.KeyCut,
	"KEY_HELP":             vinput.KeyHelp,
	"KEY_MENU":             vinput.KeyMenu,
	"KEY_CALC":             vinput.KeyCalc,
	"KEY_SETUP":            vinput.KeySetup,
	"KEY_SLEEP":            vinput.KeySleep,
	"KEY_WAKEUP":           vinput.KeyWakeup,
	"KEY_FILE":             vinput.KeyFile,
	"KEY_SENDFILE":         vinput.KeySendfile,
	"KEY_DELETEFILE":       vinput.KeyDeletefile,
	"KEY_XFER":             vinput.KeyXfer,
	"KEY_PROG1":            vinput.KeyProg1,
	"KEY_PROG2":            vinput.KeyProg2,
	"KEY_WWW":              vinput.KeyWww,
	"KEY_MSDOS":            vinput.KeyMsdos,
	"KEY_COFFEE":           vinput.KeyCoffee,
	"KEY_DIRECTION":        vinput.KeyDirection,
	"KEY_CYCLEWINDOWS":     vinput.KeyCyclewindows,
	"KEY_MAIL":             vinput.KeyMail,
	"KEY_BOOKMARKS":        vinput.KeyBookmarks,
	"KEY_COMPUTER":         vinput.KeyComputer,
	"KEY_BACK":             vinput.KeyBack,
	"KEY_FORWARD":          vinput.KeyForward,
	"KEY_CLOSECD":          vinput.KeyClosecd,
	"KEY_EJECTCD":          vinput.KeyEjectcd,
	"KEY_EJECTCLOSECD":     vinput.KeyEjectclosecd,
	"KEY_NEXTSONG":         vinput.KeyNextsong,
	"KEY_PLAYPAUSE":        vinput.KeyPlaypause,
	"KEY_PREVIOUSSONG":     vinput.KeyPrevioussong,
	"KEY_STOPCD":           vinput.KeyStopcd,
	"KEY_RECORD":           vinput.KeyRecord,
	"KEY_REWIND":           vinput.KeyRewind,
	"KEY_PHONE":            vinput.KeyPhone,
	"KEY_ISO":              vinput.KeyIso,
	"KEY_CONFIG":           vinput.KeyConfig,
	"KEY_HOMEPAGE":         vinput.KeyHomepage,
	"KEY_REFRESH":          vinput.KeyRefresh,
	"KEY_EXIT":             vinput.KeyExit,
	"KEY_MOVE":             vinput.KeyMove,
	"KEY_EDIT":             vinput.KeyEdit,
	"KEY_SCROLLUP":         vinput.KeyScrollup,
	"KEY_SCROLLDOWN":       vinput.KeyScrolldown,
	"KEY_KPLEFTPAREN":      vinput.KeyKpleftparen,
	"KEY_KPRIGHTPAREN":     vinput.KeyKprightparen,
	"KEY_NEW":              vinput.KeyNew,
	"KEY_REDO":             vinput.KeyRedo,
	"KEY_F13":              vinput.KeyF13,
	"KEY_F14":              vinput.KeyF14,
	"KEY_F15":              vinput.KeyF15,
	"KEY_F16":              vinput.KeyF16,
	"KEY_F17":              vinput.KeyF17,
	"KEY_F18":              vinput.KeyF18,
	"KEY_F19":              vinput.KeyF19,
	"KEY_F20":              vinput.KeyF20,
	"KEY_F21":              vinput.KeyF21,
	"KEY_F22":              vinput.KeyF22,
	"KEY_F23":              vinput.KeyF23,
	"KEY_F24":              vinput.KeyF24,
	"KEY_PLAYCD":           vinput.KeyPlaycd,
	"KEY_PAUSECD":          vinput.KeyPausecd,
	"KEY_PROG3":            vinput.KeyProg3,
	"KEY_PROG4":            vinput.KeyProg4,
	"KEY_DASHBOARD":        vinput.KeyDashboard,
	"KEY_SUSPEND":          vinput.KeySuspend,
	"KEY_CLOSE":            vinput.KeyClose,
	"KEY_PLAY":             vinput.KeyPlay,
	"KEY_FASTFORWARD":      vinput.KeyFastforward,
	"KEY_BASSBOOST":        vinput.KeyBassboost,
	"KEY_PRINT":            vinput.KeyPrint,
	"KEY_HP":               vinput.KeyHp,
	"KEY_CAMERA":           vinput.KeyCamera,
	"KEY_SOUND":            vinput.KeySound,
	"KEY_QUESTION":         vinput.KeyQuestion,
	"KEY_EMAIL":            vinput.KeyEmail,
	"KEY_CHAT":             vinput.KeyChat,
	"KEY_SEARCH":           vinput.KeySearch,
	"KEY_CONNECT":          vinput.KeyConnect,
	"KEY_FINANCE":          vinput.KeyFinance,
	"KEY_SPORT":            vinput.KeySport,
	"KEY_SHOP":             vinput.KeyShop,
	"KEY_ALTERASE":         vinput.KeyAlterase,
	"KEY_CANCEL":           vinput.KeyCancel,
	"KEY_BRIGHTNESSDOWN":   vinput.KeyBrightnessdown,
	"KEY_BRIGHTNESSUP":     vinput.KeyBrightnessup,
	"KEY_MEDIA":            vinput.KeyMedia,
	"KEY_SWITCHVIDEOMODE":  vinput.KeySwitchvideomode,
	"KEY_KBDILLUMTOGGLE":   vinput.KeyKbdillumtoggle,
	"KEY_KBDILLUMDOWN":     vinput.KeyKbdillumdown,
	"KEY_KBDILLUMUP":       vinput.KeyKbdillumup,
	"KEY_SEND":             vinput.KeySend,
	"KEY_REPLY":            vinput.KeyReply,
	"KEY_FORWARDMAIL":      vinput.KeyForwardmail,
	"KEY_SAVE":             vinput.KeySave,
	"KEY_DOCUMENTS":        vinput.KeyDocuments,
	"KEY_BATTERY":          vinput.KeyBattery,
	"KEY_BLUETOOTH":        vinput.KeyBluetooth,
	"KEY_WLAN":             vinput.KeyWlan,
	"KEY_UWB":              vinput.KeyUwb,
	"KEY_UNKNOWN":          vinput.KeyUnknown,
	"KEY_VIDEONEXT":        vinput.KeyVideoNext,
	"KEY_VIDEOPREV":        vinput.KeyVideoPrev,
	"KEY_BRIGHTNESSCYCLE":  vinput.KeyBrightnessCycle,
	"KEY_BRIGHTNESSZERO":   vinput.KeyBrightnessZero,
	"KEY_DISPLAYOFF":       vinput.KeyDisplayOff,
	"KEY_WIMAX":            vinput.KeyWimax,
	"KEY_RFKILL":           vinput.KeyRfkill,
	"KEY_MICMUTE":          vinput.KeyMicmute,
}

var (
	kbname = flag.String("name", "xwiimote-virtual", "Name to use")
)

func loadMapping(r io.Reader) map[xwiimote.Key]vinput.Key {
	mapping := make(map[xwiimote.Key]vinput.Key)
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

func watchDevice(dev *xwiimote.Device, mapping map[xwiimote.Key]vinput.Key) {
	fmt.Printf("new device: %s\n", dev.String())
	time.Sleep(100 * time.Millisecond)
	coreif := xwiimote.InterfaceCore{}
	if err := dev.OpenInterfaces(true, &coreif); err != nil {
		fmt.Fprintf(os.Stderr, "error: unable to open device: %s", err)
	}

	kb, err := vinput.CreateKeyboard(*kbname)
	if err != nil {
		panic(err)
	}
	defer kb.Close()
	var leds xwiimote.Led

	dev.Watch(true)
	for {
		ev, err := dev.Wait(-1)
		if err != nil {
			log.Printf("unable to poll event: %v\n", err)
		}
		fmt.Printf("%T: %+v\n", ev, ev)
		switch ev := ev.(type) {
		case *xwiimote.EventKey:
			if ev.Code == xwiimote.KeyHome {
				coreif.Rumble(ev.State == xwiimote.StatePressed)
				continue
			} else if ev.Code == xwiimote.KeyTwo {
				if ev.State == xwiimote.StatePressed {
					leds++
					leds %= 16

					fmt.Println(dev.SetLED(leds))
					continue
				}
			}

			realkey, ok := mapping[ev.Code]
			if !ok {
				continue
			}
			kb.Key(realkey, ev.State != xwiimote.StateReleased)
		case *xwiimote.EventGone:
			return
		}
	}
}

func main() {
	flag.Parse()

	mapping := loadMapping(os.Stdin)

	monitor, err := xwiimote.NewMonitor(xwiimote.MonitorUdev)
	if err != nil {
		log.Fatalln("error: ", err)
	}

	for {
		fmt.Println("waiting for devices...")
		dev, err := monitor.Wait(-1)
		if err != nil || dev == nil {
			log.Printf("error while polling: %v\n", err)
			continue
		}
		watchDevice(dev, mapping)
	}
}
