package main

import (
	"bufio"
	"flag"
	"fmt"
	"io"
	"log"
	"os"
	"strings"

	"github.com/friedelschoen/go-xwiimote"
	"github.com/friedelschoen/go-xwiimote/pkg/virtdev"
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

var realKeynames = map[string]virtdev.Key{
	"KEY_ESC":              virtdev.KeyEsc,
	"KEY_1":                virtdev.Key1,
	"KEY_2":                virtdev.Key2,
	"KEY_3":                virtdev.Key3,
	"KEY_4":                virtdev.Key4,
	"KEY_5":                virtdev.Key5,
	"KEY_6":                virtdev.Key6,
	"KEY_7":                virtdev.Key7,
	"KEY_8":                virtdev.Key8,
	"KEY_9":                virtdev.Key9,
	"KEY_0":                virtdev.Key0,
	"KEY_MINUS":            virtdev.KeyMinus,
	"KEY_EQUAL":            virtdev.KeyEqual,
	"KEY_BACKSPACE":        virtdev.KeyBackspace,
	"KEY_TAB":              virtdev.KeyTab,
	"KEY_Q":                virtdev.KeyQ,
	"KEY_W":                virtdev.KeyW,
	"KEY_E":                virtdev.KeyE,
	"KEY_R":                virtdev.KeyR,
	"KEY_T":                virtdev.KeyT,
	"KEY_Y":                virtdev.KeyY,
	"KEY_U":                virtdev.KeyU,
	"KEY_I":                virtdev.KeyI,
	"KEY_O":                virtdev.KeyO,
	"KEY_P":                virtdev.KeyP,
	"KEY_LEFTBRACE":        virtdev.KeyLeftbrace,
	"KEY_RIGHTBRACE":       virtdev.KeyRightbrace,
	"KEY_ENTER":            virtdev.KeyEnter,
	"KEY_LEFTCTRL":         virtdev.KeyLeftctrl,
	"KEY_A":                virtdev.KeyA,
	"KEY_S":                virtdev.KeyS,
	"KEY_D":                virtdev.KeyD,
	"KEY_F":                virtdev.KeyF,
	"KEY_G":                virtdev.KeyG,
	"KEY_H":                virtdev.KeyH,
	"KEY_J":                virtdev.KeyJ,
	"KEY_K":                virtdev.KeyK,
	"KEY_L":                virtdev.KeyL,
	"KEY_SEMICOLON":        virtdev.KeySemicolon,
	"KEY_APOSTROPHE":       virtdev.KeyApostrophe,
	"KEY_GRAVE":            virtdev.KeyGrave,
	"KEY_LEFTSHIFT":        virtdev.KeyLeftshift,
	"KEY_BACKSLASH":        virtdev.KeyBackslash,
	"KEY_Z":                virtdev.KeyZ,
	"KEY_X":                virtdev.KeyX,
	"KEY_C":                virtdev.KeyC,
	"KEY_V":                virtdev.KeyV,
	"KEY_B":                virtdev.KeyB,
	"KEY_N":                virtdev.KeyN,
	"KEY_M":                virtdev.KeyM,
	"KEY_COMMA":            virtdev.KeyComma,
	"KEY_DOT":              virtdev.KeyDot,
	"KEY_SLASH":            virtdev.KeySlash,
	"KEY_RIGHTSHIFT":       virtdev.KeyRightshift,
	"KEY_KPASTERISK":       virtdev.KeyKpasterisk,
	"KEY_LEFTALT":          virtdev.KeyLeftalt,
	"KEY_SPACE":            virtdev.KeySpace,
	"KEY_CAPSLOCK":         virtdev.KeyCapslock,
	"KEY_F1":               virtdev.KeyF1,
	"KEY_F2":               virtdev.KeyF2,
	"KEY_F3":               virtdev.KeyF3,
	"KEY_F4":               virtdev.KeyF4,
	"KEY_F5":               virtdev.KeyF5,
	"KEY_F6":               virtdev.KeyF6,
	"KEY_F7":               virtdev.KeyF7,
	"KEY_F8":               virtdev.KeyF8,
	"KEY_F9":               virtdev.KeyF9,
	"KEY_F10":              virtdev.KeyF10,
	"KEY_NUMLOCK":          virtdev.KeyNumlock,
	"KEY_SCROLLLOCK":       virtdev.KeyScrolllock,
	"KEY_KP7":              virtdev.KeyKp7,
	"KEY_KP8":              virtdev.KeyKp8,
	"KEY_KP9":              virtdev.KeyKp9,
	"KEY_KPMINUS":          virtdev.KeyKpminus,
	"KEY_KP4":              virtdev.KeyKp4,
	"KEY_KP5":              virtdev.KeyKp5,
	"KEY_KP6":              virtdev.KeyKp6,
	"KEY_KPPLUS":           virtdev.KeyKpplus,
	"KEY_KP1":              virtdev.KeyKp1,
	"KEY_KP2":              virtdev.KeyKp2,
	"KEY_KP3":              virtdev.KeyKp3,
	"KEY_KP0":              virtdev.KeyKp0,
	"KEY_KPDOT":            virtdev.KeyKpdot,
	"KEY_ZENKAKUHANKAKU":   virtdev.KeyZenkakuhankaku,
	"KEY_102ND":            virtdev.Key102nd,
	"KEY_F11":              virtdev.KeyF11,
	"KEY_F12":              virtdev.KeyF12,
	"KEY_RO":               virtdev.KeyRo,
	"KEY_KATAKANA":         virtdev.KeyKatakana,
	"KEY_HIRAGANA":         virtdev.KeyHiragana,
	"KEY_HENKAN":           virtdev.KeyHenkan,
	"KEY_KATAKANAHIRAGANA": virtdev.KeyKatakanahiragana,
	"KEY_MUHENKAN":         virtdev.KeyMuhenkan,
	"KEY_KPJPCOMMA":        virtdev.KeyKpjpcomma,
	"KEY_KPENTER":          virtdev.KeyKpenter,
	"KEY_RIGHTCTRL":        virtdev.KeyRightctrl,
	"KEY_KPSLASH":          virtdev.KeyKpslash,
	"KEY_SYSRQ":            virtdev.KeySysrq,
	"KEY_RIGHTALT":         virtdev.KeyRightalt,
	"KEY_LINEFEED":         virtdev.KeyLinefeed,
	"KEY_HOME":             virtdev.KeyHome,
	"KEY_UP":               virtdev.KeyUp,
	"KEY_PAGEUP":           virtdev.KeyPageup,
	"KEY_LEFT":             virtdev.KeyLeft,
	"KEY_RIGHT":            virtdev.KeyRight,
	"KEY_END":              virtdev.KeyEnd,
	"KEY_DOWN":             virtdev.KeyDown,
	"KEY_PAGEDOWN":         virtdev.KeyPagedown,
	"KEY_INSERT":           virtdev.KeyInsert,
	"KEY_DELETE":           virtdev.KeyDelete,
	"KEY_MACRO":            virtdev.KeyMacro,
	"KEY_MUTE":             virtdev.KeyMute,
	"KEY_VOLUMEDOWN":       virtdev.KeyVolumedown,
	"KEY_VOLUMEUP":         virtdev.KeyVolumeup,
	"KEY_POWER":            virtdev.KeyPower,
	"KEY_KPEQUAL":          virtdev.KeyKpequal,
	"KEY_KPPLUSMINUS":      virtdev.KeyKpplusminus,
	"KEY_PAUSE":            virtdev.KeyPause,
	"KEY_SCALE":            virtdev.KeyScale,
	"KEY_KPCOMMA":          virtdev.KeyKpcomma,
	"KEY_HANGEUL":          virtdev.KeyHangeul,
	"KEY_HANJA":            virtdev.KeyHanja,
	"KEY_YEN":              virtdev.KeyYen,
	"KEY_LEFTMETA":         virtdev.KeyLeftmeta,
	"KEY_RIGHTMETA":        virtdev.KeyRightmeta,
	"KEY_COMPOSE":          virtdev.KeyCompose,
	"KEY_STOP":             virtdev.KeyStop,
	"KEY_AGAIN":            virtdev.KeyAgain,
	"KEY_PROPS":            virtdev.KeyProps,
	"KEY_UNDO":             virtdev.KeyUndo,
	"KEY_FRONT":            virtdev.KeyFront,
	"KEY_COPY":             virtdev.KeyCopy,
	"KEY_OPEN":             virtdev.KeyOpen,
	"KEY_PASTE":            virtdev.KeyPaste,
	"KEY_FIND":             virtdev.KeyFind,
	"KEY_CUT":              virtdev.KeyCut,
	"KEY_HELP":             virtdev.KeyHelp,
	"KEY_MENU":             virtdev.KeyMenu,
	"KEY_CALC":             virtdev.KeyCalc,
	"KEY_SETUP":            virtdev.KeySetup,
	"KEY_SLEEP":            virtdev.KeySleep,
	"KEY_WAKEUP":           virtdev.KeyWakeup,
	"KEY_FILE":             virtdev.KeyFile,
	"KEY_SENDFILE":         virtdev.KeySendfile,
	"KEY_DELETEFILE":       virtdev.KeyDeletefile,
	"KEY_XFER":             virtdev.KeyXfer,
	"KEY_PROG1":            virtdev.KeyProg1,
	"KEY_PROG2":            virtdev.KeyProg2,
	"KEY_WWW":              virtdev.KeyWww,
	"KEY_MSDOS":            virtdev.KeyMsdos,
	"KEY_COFFEE":           virtdev.KeyCoffee,
	"KEY_DIRECTION":        virtdev.KeyDirection,
	"KEY_CYCLEWINDOWS":     virtdev.KeyCyclewindows,
	"KEY_MAIL":             virtdev.KeyMail,
	"KEY_BOOKMARKS":        virtdev.KeyBookmarks,
	"KEY_COMPUTER":         virtdev.KeyComputer,
	"KEY_BACK":             virtdev.KeyBack,
	"KEY_FORWARD":          virtdev.KeyForward,
	"KEY_CLOSECD":          virtdev.KeyClosecd,
	"KEY_EJECTCD":          virtdev.KeyEjectcd,
	"KEY_EJECTCLOSECD":     virtdev.KeyEjectclosecd,
	"KEY_NEXTSONG":         virtdev.KeyNextsong,
	"KEY_PLAYPAUSE":        virtdev.KeyPlaypause,
	"KEY_PREVIOUSSONG":     virtdev.KeyPrevioussong,
	"KEY_STOPCD":           virtdev.KeyStopcd,
	"KEY_RECORD":           virtdev.KeyRecord,
	"KEY_REWIND":           virtdev.KeyRewind,
	"KEY_PHONE":            virtdev.KeyPhone,
	"KEY_ISO":              virtdev.KeyIso,
	"KEY_CONFIG":           virtdev.KeyConfig,
	"KEY_HOMEPAGE":         virtdev.KeyHomepage,
	"KEY_REFRESH":          virtdev.KeyRefresh,
	"KEY_EXIT":             virtdev.KeyExit,
	"KEY_MOVE":             virtdev.KeyMove,
	"KEY_EDIT":             virtdev.KeyEdit,
	"KEY_SCROLLUP":         virtdev.KeyScrollup,
	"KEY_SCROLLDOWN":       virtdev.KeyScrolldown,
	"KEY_KPLEFTPAREN":      virtdev.KeyKpleftparen,
	"KEY_KPRIGHTPAREN":     virtdev.KeyKprightparen,
	"KEY_NEW":              virtdev.KeyNew,
	"KEY_REDO":             virtdev.KeyRedo,
	"KEY_F13":              virtdev.KeyF13,
	"KEY_F14":              virtdev.KeyF14,
	"KEY_F15":              virtdev.KeyF15,
	"KEY_F16":              virtdev.KeyF16,
	"KEY_F17":              virtdev.KeyF17,
	"KEY_F18":              virtdev.KeyF18,
	"KEY_F19":              virtdev.KeyF19,
	"KEY_F20":              virtdev.KeyF20,
	"KEY_F21":              virtdev.KeyF21,
	"KEY_F22":              virtdev.KeyF22,
	"KEY_F23":              virtdev.KeyF23,
	"KEY_F24":              virtdev.KeyF24,
	"KEY_PLAYCD":           virtdev.KeyPlaycd,
	"KEY_PAUSECD":          virtdev.KeyPausecd,
	"KEY_PROG3":            virtdev.KeyProg3,
	"KEY_PROG4":            virtdev.KeyProg4,
	"KEY_DASHBOARD":        virtdev.KeyDashboard,
	"KEY_SUSPEND":          virtdev.KeySuspend,
	"KEY_CLOSE":            virtdev.KeyClose,
	"KEY_PLAY":             virtdev.KeyPlay,
	"KEY_FASTFORWARD":      virtdev.KeyFastforward,
	"KEY_BASSBOOST":        virtdev.KeyBassboost,
	"KEY_PRINT":            virtdev.KeyPrint,
	"KEY_HP":               virtdev.KeyHp,
	"KEY_CAMERA":           virtdev.KeyCamera,
	"KEY_SOUND":            virtdev.KeySound,
	"KEY_QUESTION":         virtdev.KeyQuestion,
	"KEY_EMAIL":            virtdev.KeyEmail,
	"KEY_CHAT":             virtdev.KeyChat,
	"KEY_SEARCH":           virtdev.KeySearch,
	"KEY_CONNECT":          virtdev.KeyConnect,
	"KEY_FINANCE":          virtdev.KeyFinance,
	"KEY_SPORT":            virtdev.KeySport,
	"KEY_SHOP":             virtdev.KeyShop,
	"KEY_ALTERASE":         virtdev.KeyAlterase,
	"KEY_CANCEL":           virtdev.KeyCancel,
	"KEY_BRIGHTNESSDOWN":   virtdev.KeyBrightnessdown,
	"KEY_BRIGHTNESSUP":     virtdev.KeyBrightnessup,
	"KEY_MEDIA":            virtdev.KeyMedia,
	"KEY_SWITCHVIDEOMODE":  virtdev.KeySwitchvideomode,
	"KEY_KBDILLUMTOGGLE":   virtdev.KeyKbdillumtoggle,
	"KEY_KBDILLUMDOWN":     virtdev.KeyKbdillumdown,
	"KEY_KBDILLUMUP":       virtdev.KeyKbdillumup,
	"KEY_SEND":             virtdev.KeySend,
	"KEY_REPLY":            virtdev.KeyReply,
	"KEY_FORWARDMAIL":      virtdev.KeyForwardmail,
	"KEY_SAVE":             virtdev.KeySave,
	"KEY_DOCUMENTS":        virtdev.KeyDocuments,
	"KEY_BATTERY":          virtdev.KeyBattery,
	"KEY_BLUETOOTH":        virtdev.KeyBluetooth,
	"KEY_WLAN":             virtdev.KeyWlan,
	"KEY_UWB":              virtdev.KeyUwb,
	"KEY_UNKNOWN":          virtdev.KeyUnknown,
	"KEY_VIDEONEXT":        virtdev.KeyVideoNext,
	"KEY_VIDEOPREV":        virtdev.KeyVideoPrev,
	"KEY_BRIGHTNESSCYCLE":  virtdev.KeyBrightnessCycle,
	"KEY_BRIGHTNESSZERO":   virtdev.KeyBrightnessZero,
	"KEY_DISPLAYOFF":       virtdev.KeyDisplayOff,
	"KEY_WIMAX":            virtdev.KeyWimax,
	"KEY_RFKILL":           virtdev.KeyRfkill,
	"KEY_MICMUTE":          virtdev.KeyMicmute,
}

var (
	kbname = flag.String("name", "xwiimote-virtual", "Name to use")
)

func loadMapping(r io.Reader) map[xwiimote.Key]virtdev.Key {
	mapping := make(map[xwiimote.Key]virtdev.Key)
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

func watchDevice(path string, mapping map[xwiimote.Key]virtdev.Key) {
	dev, err := xwiimote.NewDevice(path)
	if err != nil {
		fmt.Fprintf(os.Stderr, "error: unable to get device: %s", err)
	}
	defer dev.Free()

	if err := dev.Open(xwiimote.InterfaceCore); err != nil {
		fmt.Fprintf(os.Stderr, "error: unable to open device: %s", err)
	}

	kb, err := virtdev.CreateKeyboard(*kbname)
	if err != nil {
		panic(err)
	}
	defer kb.Close()

	for {
		ev, err := dev.Wait(-1)
		if err != nil {
			log.Printf("unable to poll event: %v\n", err)
		}
		switch ev := ev.(type) {
		case *xwiimote.EventKey:
			realkey, ok := mapping[ev.Code]
			if !ok {
				continue
			}
			kb.Key(realkey, ev.State != xwiimote.StateReleased)
		}
	}
}

func main() {
	flag.Parse()

	mapping := loadMapping(os.Stdin)

	monitor := xwiimote.NewMonitor(xwiimote.MonitorUdev)
	defer monitor.Free()

	for {
		path, err := monitor.Wait(-1)
		if err != nil || path == "" {
			log.Printf("error while polling: %v\n", err)
			continue
		}
		watchDevice(path, mapping)
	}
}
