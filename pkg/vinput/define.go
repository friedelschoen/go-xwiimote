package vinput

// #include <linux/uinput.h>
// #define SYSNAME_LEN 64
// #define GET_SYSNAME UI_GET_SYSNAME(SYSNAME_LEN)
import "C"
import (
	"syscall"
	"unsafe"
)

const (
	uiMaxNameSize = C.UINPUT_MAX_NAME_SIZE

	uiDevSetup   = C.UI_DEV_SETUP
	uiDevCreate  = C.UI_DEV_CREATE
	uiDevDestroy = C.UI_DEV_DESTROY
	uiAbsSetup   = C.UI_ABS_SETUP
	uiSysnameLen = C.SYSNAME_LEN
	uiGetSysname = C.GET_SYSNAME
	uiSetEvBit   = C.UI_SET_EVBIT
	uiSetKeyBit  = C.UI_SET_KEYBIT
	uiSetRelBit  = C.UI_SET_RELBIT

	busUsb = C.BUS_USB

	evSyn = C.EV_SYN
	evKey = C.EV_KEY
	evRel = C.EV_REL
	evAbs = C.EV_ABS

	relX      = C.REL_X
	relY      = C.REL_Y
	relHWheel = C.REL_HWHEEL
	relWheel  = C.REL_WHEEL

	absX = C.ABS_X
	absY = C.ABS_Y
	// absSize = C.ABS_CNT

	synReport = C.SYN_REPORT
)

type uinputSetup struct {
	id           inputID
	name         [uiMaxNameSize]byte
	ffEffectsMax uint32
}

type inputID struct {
	Bustype uint16
	Vendor  uint16
	Product uint16
	Version uint16
}

// translated to go from input.h
type inputEvent struct {
	Time  syscall.Timeval
	Type  uint16
	Code  uint16
	Value int32
}

func (iev inputEvent) buffer() []byte {
	buf := (*[unsafe.Sizeof(iev)]byte)(unsafe.Pointer(&iev))
	return buf[:]
}

type absInfo struct {
	value      int32
	minimum    int32
	maximum    int32
	fuzz       int32
	flat       int32
	resolution int32
}

type absSetup struct {
	code    uint16 // axis code
	absinfo absInfo
}

// Key is used by Keyboard and Mouse to represent a physical key-press. The keys are obtained directly from input.h.
type Key uint16

const (
	KeyReserved   Key = C.KEY_RESERVED
	KeyEsc        Key = C.KEY_ESC
	Key1          Key = C.KEY_1
	Key2          Key = C.KEY_2
	Key3          Key = C.KEY_3
	Key4          Key = C.KEY_4
	Key5          Key = C.KEY_5
	Key6          Key = C.KEY_6
	Key7          Key = C.KEY_7
	Key8          Key = C.KEY_8
	Key9          Key = C.KEY_9
	Key0          Key = C.KEY_0
	KeyMinus      Key = C.KEY_MINUS
	KeyEqual      Key = C.KEY_EQUAL
	KeyBackspace  Key = C.KEY_BACKSPACE
	KeyTab        Key = C.KEY_TAB
	KeyQ          Key = C.KEY_Q
	KeyW          Key = C.KEY_W
	KeyE          Key = C.KEY_E
	KeyR          Key = C.KEY_R
	KeyT          Key = C.KEY_T
	KeyY          Key = C.KEY_Y
	KeyU          Key = C.KEY_U
	KeyI          Key = C.KEY_I
	KeyO          Key = C.KEY_O
	KeyP          Key = C.KEY_P
	KeyLeftbrace  Key = C.KEY_LEFTBRACE
	KeyRightbrace Key = C.KEY_RIGHTBRACE
	KeyEnter      Key = C.KEY_ENTER
	KeyLeftctrl   Key = C.KEY_LEFTCTRL
	KeyA          Key = C.KEY_A
	KeyS          Key = C.KEY_S
	KeyD          Key = C.KEY_D
	KeyF          Key = C.KEY_F
	KeyG          Key = C.KEY_G
	KeyH          Key = C.KEY_H
	KeyJ          Key = C.KEY_J
	KeyK          Key = C.KEY_K
	KeyL          Key = C.KEY_L
	KeySemicolon  Key = C.KEY_SEMICOLON
	KeyApostrophe Key = C.KEY_APOSTROPHE
	KeyGrave      Key = C.KEY_GRAVE
	KeyLeftshift  Key = C.KEY_LEFTSHIFT
	KeyBackslash  Key = C.KEY_BACKSLASH
	KeyZ          Key = C.KEY_Z
	KeyX          Key = C.KEY_X
	KeyC          Key = C.KEY_C
	KeyV          Key = C.KEY_V
	KeyB          Key = C.KEY_B
	KeyN          Key = C.KEY_N
	KeyM          Key = C.KEY_M
	KeyComma      Key = C.KEY_COMMA
	KeyDot        Key = C.KEY_DOT
	KeySlash      Key = C.KEY_SLASH
	KeyRightshift Key = C.KEY_RIGHTSHIFT
	KeyKpasterisk Key = C.KEY_KPASTERISK
	KeyLeftalt    Key = C.KEY_LEFTALT
	KeySpace      Key = C.KEY_SPACE
	KeyCapslock   Key = C.KEY_CAPSLOCK
	KeyF1         Key = C.KEY_F1
	KeyF2         Key = C.KEY_F2
	KeyF3         Key = C.KEY_F3
	KeyF4         Key = C.KEY_F4
	KeyF5         Key = C.KEY_F5
	KeyF6         Key = C.KEY_F6
	KeyF7         Key = C.KEY_F7
	KeyF8         Key = C.KEY_F8
	KeyF9         Key = C.KEY_F9
	KeyF10        Key = C.KEY_F10
	KeyNumlock    Key = C.KEY_NUMLOCK
	KeyScrolllock Key = C.KEY_SCROLLLOCK
	KeyKp7        Key = C.KEY_KP7
	KeyKp8        Key = C.KEY_KP8
	KeyKp9        Key = C.KEY_KP9
	KeyKpminus    Key = C.KEY_KPMINUS
	KeyKp4        Key = C.KEY_KP4
	KeyKp5        Key = C.KEY_KP5
	KeyKp6        Key = C.KEY_KP6
	KeyKpplus     Key = C.KEY_KPPLUS
	KeyKp1        Key = C.KEY_KP1
	KeyKp2        Key = C.KEY_KP2
	KeyKp3        Key = C.KEY_KP3
	KeyKp0        Key = C.KEY_KP0
	KeyKpdot      Key = C.KEY_KPDOT

	KeyZenkakuhankaku   Key = C.KEY_ZENKAKUHANKAKU
	Key102nd            Key = C.KEY_102ND
	KeyF11              Key = C.KEY_F11
	KeyF12              Key = C.KEY_F12
	KeyRo               Key = C.KEY_RO
	KeyKatakana         Key = C.KEY_KATAKANA
	KeyHiragana         Key = C.KEY_HIRAGANA
	KeyHenkan           Key = C.KEY_HENKAN
	KeyKatakanahiragana Key = C.KEY_KATAKANAHIRAGANA
	KeyMuhenkan         Key = C.KEY_MUHENKAN
	KeyKpjpcomma        Key = C.KEY_KPJPCOMMA
	KeyKpenter          Key = C.KEY_KPENTER
	KeyRightctrl        Key = C.KEY_RIGHTCTRL
	KeyKpslash          Key = C.KEY_KPSLASH
	KeySysrq            Key = C.KEY_SYSRQ
	KeyRightalt         Key = C.KEY_RIGHTALT
	KeyLinefeed         Key = C.KEY_LINEFEED
	KeyHome             Key = C.KEY_HOME
	KeyUp               Key = C.KEY_UP
	KeyPageup           Key = C.KEY_PAGEUP
	KeyLeft             Key = C.KEY_LEFT
	KeyRight            Key = C.KEY_RIGHT
	KeyEnd              Key = C.KEY_END
	KeyDown             Key = C.KEY_DOWN
	KeyPagedown         Key = C.KEY_PAGEDOWN
	KeyInsert           Key = C.KEY_INSERT
	KeyDelete           Key = C.KEY_DELETE
	KeyMacro            Key = C.KEY_MACRO
	KeyMute             Key = C.KEY_MUTE
	KeyVolumedown       Key = C.KEY_VOLUMEDOWN
	KeyVolumeup         Key = C.KEY_VOLUMEUP
	KeyPower            Key = C.KEY_POWER // SC System Power Down
	KeyKpequal          Key = C.KEY_KPEQUAL
	KeyKpplusminus      Key = C.KEY_KPPLUSMINUS
	KeyPause            Key = C.KEY_PAUSE
	KeyScale            Key = C.KEY_SCALE // AL Compiz Scale (Expose)

	KeyKpcomma   Key = C.KEY_KPCOMMA
	KeyHangeul   Key = C.KEY_HANGEUL
	KeyHanguel   Key = C.KEY_HANGUEL
	KeyHanja     Key = C.KEY_HANJA
	KeyYen       Key = C.KEY_YEN
	KeyLeftmeta  Key = C.KEY_LEFTMETA
	KeyRightmeta Key = C.KEY_RIGHTMETA
	KeyCompose   Key = C.KEY_COMPOSE

	KeyStop          Key = C.KEY_STOP // AC Stop
	KeyAgain         Key = C.KEY_AGAIN
	KeyProps         Key = C.KEY_PROPS // AC Properties
	KeyUndo          Key = C.KEY_UNDO  // AC Undo
	KeyFront         Key = C.KEY_FRONT
	KeyCopy          Key = C.KEY_COPY  // AC Copy
	KeyOpen          Key = C.KEY_OPEN  // AC Open
	KeyPaste         Key = C.KEY_PASTE // AC Paste
	KeyFind          Key = C.KEY_FIND  // AC Search
	KeyCut           Key = C.KEY_CUT   // AC Cut
	KeyHelp          Key = C.KEY_HELP  // AL Integrated Help Center
	KeyMenu          Key = C.KEY_MENU  // Menu (show menu)
	KeyCalc          Key = C.KEY_CALC  // AL Calculator
	KeySetup         Key = C.KEY_SETUP
	KeySleep         Key = C.KEY_SLEEP  // SC System Sleep
	KeyWakeup        Key = C.KEY_WAKEUP // System Wake Up
	KeyFile          Key = C.KEY_FILE   // AL Local Machine Browser
	KeySendfile      Key = C.KEY_SENDFILE
	KeyDeletefile    Key = C.KEY_DELETEFILE
	KeyXfer          Key = C.KEY_XFER
	KeyProg1         Key = C.KEY_PROG1
	KeyProg2         Key = C.KEY_PROG2
	KeyWww           Key = C.KEY_WWW // AL Internet Browser
	KeyMsdos         Key = C.KEY_MSDOS
	KeyCoffee        Key = C.KEY_COFFEE // AL Terminal Lock/Screensaver
	KeyScreenlock    Key = C.KEY_SCREENLOCK
	KeyRotateDisplay Key = C.KEY_ROTATE_DISPLAY // Display orientation for e.g. tablets
	KeyDirection     Key = C.KEY_DIRECTION
	KeyCyclewindows  Key = C.KEY_CYCLEWINDOWS
	KeyMail          Key = C.KEY_MAIL
	KeyBookmarks     Key = C.KEY_BOOKMARKS // AC Bookmarks
	KeyComputer      Key = C.KEY_COMPUTER
	KeyBack          Key = C.KEY_BACK    // AC Back
	KeyForward       Key = C.KEY_FORWARD // AC Forward
	KeyClosecd       Key = C.KEY_CLOSECD
	KeyEjectcd       Key = C.KEY_EJECTCD
	KeyEjectclosecd  Key = C.KEY_EJECTCLOSECD
	KeyNextsong      Key = C.KEY_NEXTSONG
	KeyPlaypause     Key = C.KEY_PLAYPAUSE
	KeyPrevioussong  Key = C.KEY_PREVIOUSSONG
	KeyStopcd        Key = C.KEY_STOPCD
	KeyRecord        Key = C.KEY_RECORD
	KeyRewind        Key = C.KEY_REWIND
	KeyPhone         Key = C.KEY_PHONE // Media Select Telephone
	KeyIso           Key = C.KEY_ISO
	KeyConfig        Key = C.KEY_CONFIG   // AL Consumer Control Configuration
	KeyHomepage      Key = C.KEY_HOMEPAGE // AC Home
	KeyRefresh       Key = C.KEY_REFRESH  // AC Refresh
	KeyExit          Key = C.KEY_EXIT     // AC Exit
	KeyMove          Key = C.KEY_MOVE
	KeyEdit          Key = C.KEY_EDIT
	KeyScrollup      Key = C.KEY_SCROLLUP
	KeyScrolldown    Key = C.KEY_SCROLLDOWN
	KeyKpleftparen   Key = C.KEY_KPLEFTPAREN
	KeyKprightparen  Key = C.KEY_KPRIGHTPAREN
	KeyNew           Key = C.KEY_NEW  // AC New
	KeyRedo          Key = C.KEY_REDO // AC Redo/Repeat

	KeyF13 Key = C.KEY_F13
	KeyF14 Key = C.KEY_F14
	KeyF15 Key = C.KEY_F15
	KeyF16 Key = C.KEY_F16
	KeyF17 Key = C.KEY_F17
	KeyF18 Key = C.KEY_F18
	KeyF19 Key = C.KEY_F19
	KeyF20 Key = C.KEY_F20
	KeyF21 Key = C.KEY_F21
	KeyF22 Key = C.KEY_F22
	KeyF23 Key = C.KEY_F23
	KeyF24 Key = C.KEY_F24

	KeyPlaycd          Key = C.KEY_PLAYCD
	KeyPausecd         Key = C.KEY_PAUSECD
	KeyProg3           Key = C.KEY_PROG3
	KeyProg4           Key = C.KEY_PROG4
	KeyAllApplications Key = C.KEY_ALL_APPLICATIONS // AC Desktop Show All Applications
	KeyDashboard       Key = C.KEY_DASHBOARD
	KeySuspend         Key = C.KEY_SUSPEND
	KeyClose           Key = C.KEY_CLOSE // AC Close
	KeyPlay            Key = C.KEY_PLAY
	KeyFastforward     Key = C.KEY_FASTFORWARD
	KeyBassboost       Key = C.KEY_BASSBOOST
	KeyPrint           Key = C.KEY_PRINT // AC Print
	KeyHp              Key = C.KEY_HP
	KeyCamera          Key = C.KEY_CAMERA
	KeySound           Key = C.KEY_SOUND
	KeyQuestion        Key = C.KEY_QUESTION
	KeyEmail           Key = C.KEY_EMAIL
	KeyChat            Key = C.KEY_CHAT
	KeySearch          Key = C.KEY_SEARCH
	KeyConnect         Key = C.KEY_CONNECT
	KeyFinance         Key = C.KEY_FINANCE // AL Checkbook/Finance
	KeySport           Key = C.KEY_SPORT
	KeyShop            Key = C.KEY_SHOP
	KeyAlterase        Key = C.KEY_ALTERASE
	KeyCancel          Key = C.KEY_CANCEL // AC Cancel
	KeyBrightnessdown  Key = C.KEY_BRIGHTNESSDOWN
	KeyBrightnessup    Key = C.KEY_BRIGHTNESSUP
	KeyMedia           Key = C.KEY_MEDIA

	KeySwitchvideomode Key = C.KEY_SWITCHVIDEOMODE // Cycle between available video	outputs (Monitor/LCD/TV-out/etc)
	KeyKbdillumtoggle  Key = C.KEY_KBDILLUMTOGGLE
	KeyKbdillumdown    Key = C.KEY_KBDILLUMDOWN
	KeyKbdillumup      Key = C.KEY_KBDILLUMUP

	KeySend        Key = C.KEY_SEND        // AC Send
	KeyReply       Key = C.KEY_REPLY       // AC Reply
	KeyForwardmail Key = C.KEY_FORWARDMAIL // AC Forward Msg
	KeySave        Key = C.KEY_SAVE        // AC Save
	KeyDocuments   Key = C.KEY_DOCUMENTS

	KeyBattery Key = C.KEY_BATTERY

	KeyBluetooth Key = C.KEY_BLUETOOTH
	KeyWlan      Key = C.KEY_WLAN
	KeyUwb       Key = C.KEY_UWB

	KeyUnknown Key = C.KEY_UNKNOWN

	KeyVideoNext       Key = C.KEY_VIDEO_NEXT       // drive next video source
	KeyVideoPrev       Key = C.KEY_VIDEO_PREV       // drive previous video source
	KeyBrightnessCycle Key = C.KEY_BRIGHTNESS_CYCLE // brightness up, after max is min
	KeyBrightnessAuto  Key = C.KEY_BRIGHTNESS_AUTO  // Set Auto Brightness: manual brightness control is off, rely on ambient
	KeyBrightnessZero  Key = C.KEY_BRIGHTNESS_ZERO
	KeyDisplayOff      Key = C.KEY_DISPLAY_OFF // display device to off state

	KeyWwan   Key = C.KEY_WWAN // Wireless WAN (LTE, UMTS, GSM, etc.)
	KeyWimax  Key = C.KEY_WIMAX
	KeyRfkill Key = C.KEY_RFKILL // Key that controls all radios

	KeyMicmute Key = C.KEY_MICMUTE // Mute / unmute the microphone

	// Code 255 is reserved for special needs of AT keyboard driver

	ButtonMisc Key = C.BTN_MISC
	Button0    Key = C.BTN_0
	Button1    Key = C.BTN_1
	Button2    Key = C.BTN_2
	Button3    Key = C.BTN_3
	Button4    Key = C.BTN_4
	Button5    Key = C.BTN_5
	Button6    Key = C.BTN_6
	Button7    Key = C.BTN_7
	Button8    Key = C.BTN_8
	Button9    Key = C.BTN_9

	ButtonMouse   Key = C.BTN_MOUSE
	ButtonLeft    Key = C.BTN_LEFT
	ButtonRight   Key = C.BTN_RIGHT
	ButtonMiddle  Key = C.BTN_MIDDLE
	ButtonSide    Key = C.BTN_SIDE
	ButtonExtra   Key = C.BTN_EXTRA
	ButtonForward Key = C.BTN_FORWARD
	ButtonBack    Key = C.BTN_BACK
	ButtonTask    Key = C.BTN_TASK

	ButtonJoystick Key = C.BTN_JOYSTICK
	ButtonTrigger  Key = C.BTN_TRIGGER
	ButtonThumb    Key = C.BTN_THUMB
	ButtonThumb2   Key = C.BTN_THUMB2
	ButtonTop      Key = C.BTN_TOP
	ButtonTop2     Key = C.BTN_TOP2
	ButtonPinkie   Key = C.BTN_PINKIE
	ButtonBase     Key = C.BTN_BASE
	ButtonBase2    Key = C.BTN_BASE2
	ButtonBase3    Key = C.BTN_BASE3
	ButtonBase4    Key = C.BTN_BASE4
	ButtonBase5    Key = C.BTN_BASE5
	ButtonBase6    Key = C.BTN_BASE6
	ButtonDead     Key = C.BTN_DEAD

	ButtonGamepad Key = C.BTN_GAMEPAD
	ButtonSouth   Key = C.BTN_SOUTH
	ButtonA       Key = C.BTN_A
	ButtonEast    Key = C.BTN_EAST
	ButtonB       Key = C.BTN_B
	ButtonC       Key = C.BTN_C
	ButtonNorth   Key = C.BTN_NORTH
	ButtonX       Key = C.BTN_X
	ButtonWest    Key = C.BTN_WEST
	ButtonY       Key = C.BTN_Y
	ButtonZ       Key = C.BTN_Z
	ButtonTl      Key = C.BTN_TL
	ButtonTr      Key = C.BTN_TR
	ButtonTl2     Key = C.BTN_TL2
	ButtonTr2     Key = C.BTN_TR2
	ButtonSelect  Key = C.BTN_SELECT
	ButtonStart   Key = C.BTN_START
	ButtonMode    Key = C.BTN_MODE
	ButtonThumbl  Key = C.BTN_THUMBL
	ButtonThumbr  Key = C.BTN_THUMBR

	ButtonDigi          Key = C.BTN_DIGI
	ButtonToolPen       Key = C.BTN_TOOL_PEN
	ButtonToolRubber    Key = C.BTN_TOOL_RUBBER
	ButtonToolBrush     Key = C.BTN_TOOL_BRUSH
	ButtonToolPencil    Key = C.BTN_TOOL_PENCIL
	ButtonToolAirbrush  Key = C.BTN_TOOL_AIRBRUSH
	ButtonToolFinger    Key = C.BTN_TOOL_FINGER
	ButtonToolMouse     Key = C.BTN_TOOL_MOUSE
	ButtonToolLens      Key = C.BTN_TOOL_LENS
	ButtonToolQuinttap  Key = C.BTN_TOOL_QUINTTAP // Five fingers on trackpad
	ButtonStylus3       Key = C.BTN_STYLUS3
	ButtonTouch         Key = C.BTN_TOUCH
	ButtonStylus        Key = C.BTN_STYLUS
	ButtonStylus2       Key = C.BTN_STYLUS2
	ButtonToolDoubletap Key = C.BTN_TOOL_DOUBLETAP
	ButtonToolTripletap Key = C.BTN_TOOL_TRIPLETAP
	ButtonToolQuadtap   Key = C.BTN_TOOL_QUADTAP // Four fingers on trackpad

	ButtonWheel    Key = C.BTN_WHEEL
	ButtonGearDown Key = C.BTN_GEAR_DOWN
	ButtonGearUp   Key = C.BTN_GEAR_UP

	KeyOk               Key = C.KEY_OK
	KeySelect           Key = C.KEY_SELECT
	KeyGoto             Key = C.KEY_GOTO
	KeyClear            Key = C.KEY_CLEAR
	KeyPower2           Key = C.KEY_POWER2
	KeyOption           Key = C.KEY_OPTION
	KeyInfo             Key = C.KEY_INFO // AL OEM Features/Tips/Tutorial
	KeyTime             Key = C.KEY_TIME
	KeyVendor           Key = C.KEY_VENDOR
	KeyArchive          Key = C.KEY_ARCHIVE
	KeyProgram          Key = C.KEY_PROGRAM // Media Select Program Guide
	KeyChannel          Key = C.KEY_CHANNEL
	KeyFavorites        Key = C.KEY_FAVORITES
	KeyEpg              Key = C.KEY_EPG
	KeyPvr              Key = C.KEY_PVR // Media Select Home
	KeyMhp              Key = C.KEY_MHP
	KeyLanguage         Key = C.KEY_LANGUAGE
	KeyTitle            Key = C.KEY_TITLE
	KeySubtitle         Key = C.KEY_SUBTITLE
	KeyAngle            Key = C.KEY_ANGLE
	KeyFullScreen       Key = C.KEY_FULL_SCREEN // AC View Toggle
	KeyZoom             Key = C.KEY_ZOOM
	KeyMode             Key = C.KEY_MODE
	KeyKeyboard         Key = C.KEY_KEYBOARD
	KeyAspectRatio      Key = C.KEY_ASPECT_RATIO // HUTRR37: Aspect
	KeyScreen           Key = C.KEY_SCREEN
	KeyPc               Key = C.KEY_PC   // Media Select Computer
	KeyTv               Key = C.KEY_TV   // Media Select TV
	KeyTv2              Key = C.KEY_TV2  // Media Select Cable
	KeyVcr              Key = C.KEY_VCR  // Media Select VCR
	KeyVcr2             Key = C.KEY_VCR2 // VCR Plus
	KeySat              Key = C.KEY_SAT  // Media Select Satellite
	KeySat2             Key = C.KEY_SAT2
	KeyCd               Key = C.KEY_CD   // Media Select CD
	KeyTape             Key = C.KEY_TAPE // Media Select Tape
	KeyRadio            Key = C.KEY_RADIO
	KeyTuner            Key = C.KEY_TUNER // Media Select Tuner
	KeyPlayer           Key = C.KEY_PLAYER
	KeyText             Key = C.KEY_TEXT
	KeyDvd              Key = C.KEY_DVD // Media Select DVD
	KeyAux              Key = C.KEY_AUX
	KeyMp3              Key = C.KEY_MP3
	KeyAudio            Key = C.KEY_AUDIO // AL Audio Browser
	KeyVideo            Key = C.KEY_VIDEO // AL Movie Browser
	KeyDirectory        Key = C.KEY_DIRECTORY
	KeyList             Key = C.KEY_LIST
	KeyMemo             Key = C.KEY_MEMO // Media Select Messages
	KeyCalendar         Key = C.KEY_CALENDAR
	KeyRed              Key = C.KEY_RED
	KeyGreen            Key = C.KEY_GREEN
	KeyYellow           Key = C.KEY_YELLOW
	KeyBlue             Key = C.KEY_BLUE
	KeyChannelup        Key = C.KEY_CHANNELUP   // Channel Increment
	KeyChanneldown      Key = C.KEY_CHANNELDOWN // Channel Decrement
	KeyFirst            Key = C.KEY_FIRST
	KeyLast             Key = C.KEY_LAST // Recall Last
	KeyAb               Key = C.KEY_AB
	KeyNext             Key = C.KEY_NEXT
	KeyRestart          Key = C.KEY_RESTART
	KeySlow             Key = C.KEY_SLOW
	KeyShuffle          Key = C.KEY_SHUFFLE
	KeyBreak            Key = C.KEY_BREAK
	KeyPrevious         Key = C.KEY_PREVIOUS
	KeyDigits           Key = C.KEY_DIGITS
	KeyTeen             Key = C.KEY_TEEN
	KeyTwen             Key = C.KEY_TWEN
	KeyVideophone       Key = C.KEY_VIDEOPHONE     // Media Select Video Phone
	KeyGames            Key = C.KEY_GAMES          // Media Select Games
	KeyZoomin           Key = C.KEY_ZOOMIN         // AC Zoom In
	KeyZoomout          Key = C.KEY_ZOOMOUT        // AC Zoom Out
	KeyZoomreset        Key = C.KEY_ZOOMRESET      // AC Zoom
	KeyWordprocessor    Key = C.KEY_WORDPROCESSOR  // AL Word Processor
	KeyEditor           Key = C.KEY_EDITOR         // AL Text Editor
	KeySpreadsheet      Key = C.KEY_SPREADSHEET    // AL Spreadsheet
	KeyGraphicseditor   Key = C.KEY_GRAPHICSEDITOR // AL Graphics Editor
	KeyPresentation     Key = C.KEY_PRESENTATION   // AL Presentation App
	KeyDatabase         Key = C.KEY_DATABASE       // AL Database App
	KeyNews             Key = C.KEY_NEWS           // AL Newsreader
	KeyVoicemail        Key = C.KEY_VOICEMAIL      // AL Voicemail
	KeyAddressbook      Key = C.KEY_ADDRESSBOOK    // AL Contacts/Address Book
	KeyMessenger        Key = C.KEY_MESSENGER      // AL Instant Messaging
	KeyDisplaytoggle    Key = C.KEY_DISPLAYTOGGLE  // Turn display (LCD) on and off
	KeyBrightnessToggle Key = C.KEY_BRIGHTNESS_TOGGLE
	KeySpellcheck       Key = C.KEY_SPELLCHECK // AL Spell Check
	KeyLogoff           Key = C.KEY_LOGOFF     // AL Logoff

	KeyDollar Key = C.KEY_DOLLAR
	KeyEuro   Key = C.KEY_EURO

	KeyFrameback          Key = C.KEY_FRAMEBACK // Consumer - transport controls
	KeyFrameforward       Key = C.KEY_FRAMEFORWARD
	KeyContextMenu        Key = C.KEY_CONTEXT_MENU        // GenDesc - system context menu
	KeyMediaRepeat        Key = C.KEY_MEDIA_REPEAT        // Consumer - transport control
	Key10channelsup       Key = C.KEY_10CHANNELSUP        // 10 channels up (10+)
	Key10channelsdown     Key = C.KEY_10CHANNELSDOWN      // 10 channels down (10-)
	KeyImages             Key = C.KEY_IMAGES              // AL Image Browser
	KeyNotificationCenter Key = C.KEY_NOTIFICATION_CENTER // Show/hide the notification center
	KeyPickupPhone        Key = C.KEY_PICKUP_PHONE        // Answer incoming call
	KeyHangupPhone        Key = C.KEY_HANGUP_PHONE        // Decline incoming call
	KeyLinkPhone          Key = C.KEY_LINK_PHONE          // AL Phone Syncing

	KeyDelEol  Key = C.KEY_DEL_EOL
	KeyDelEos  Key = C.KEY_DEL_EOS
	KeyInsLine Key = C.KEY_INS_LINE
	KeyDelLine Key = C.KEY_DEL_LINE

	KeyFn           Key = C.KEY_FN
	KeyFnEsc        Key = C.KEY_FN_ESC
	KeyFnF1         Key = C.KEY_FN_F1
	KeyFnF2         Key = C.KEY_FN_F2
	KeyFnF3         Key = C.KEY_FN_F3
	KeyFnF4         Key = C.KEY_FN_F4
	KeyFnF5         Key = C.KEY_FN_F5
	KeyFnF6         Key = C.KEY_FN_F6
	KeyFnF7         Key = C.KEY_FN_F7
	KeyFnF8         Key = C.KEY_FN_F8
	KeyFnF9         Key = C.KEY_FN_F9
	KeyFnF10        Key = C.KEY_FN_F10
	KeyFnF11        Key = C.KEY_FN_F11
	KeyFnF12        Key = C.KEY_FN_F12
	KeyFn1          Key = C.KEY_FN_1
	KeyFn2          Key = C.KEY_FN_2
	KeyFnD          Key = C.KEY_FN_D
	KeyFnE          Key = C.KEY_FN_E
	KeyFnF          Key = C.KEY_FN_F
	KeyFnS          Key = C.KEY_FN_S
	KeyFnB          Key = C.KEY_FN_B
	KeyFnRightShift Key = C.KEY_FN_RIGHT_SHIFT

	KeyBrlDot1  Key = C.KEY_BRL_DOT1
	KeyBrlDot2  Key = C.KEY_BRL_DOT2
	KeyBrlDot3  Key = C.KEY_BRL_DOT3
	KeyBrlDot4  Key = C.KEY_BRL_DOT4
	KeyBrlDot5  Key = C.KEY_BRL_DOT5
	KeyBrlDot6  Key = C.KEY_BRL_DOT6
	KeyBrlDot7  Key = C.KEY_BRL_DOT7
	KeyBrlDot8  Key = C.KEY_BRL_DOT8
	KeyBrlDot9  Key = C.KEY_BRL_DOT9
	KeyBrlDot10 Key = C.KEY_BRL_DOT10

	KeyNumeric0     Key = C.KEY_NUMERIC_0 // used by phones, remote controls,
	KeyNumeric1     Key = C.KEY_NUMERIC_1 // and other keypads
	KeyNumeric2     Key = C.KEY_NUMERIC_2
	KeyNumeric3     Key = C.KEY_NUMERIC_3
	KeyNumeric4     Key = C.KEY_NUMERIC_4
	KeyNumeric5     Key = C.KEY_NUMERIC_5
	KeyNumeric6     Key = C.KEY_NUMERIC_6
	KeyNumeric7     Key = C.KEY_NUMERIC_7
	KeyNumeric8     Key = C.KEY_NUMERIC_8
	KeyNumeric9     Key = C.KEY_NUMERIC_9
	KeyNumericStar  Key = C.KEY_NUMERIC_STAR
	KeyNumericPound Key = C.KEY_NUMERIC_POUND
	KeyNumericA     Key = C.KEY_NUMERIC_A // Phone key A - HUT Telephony 0xb9
	KeyNumericB     Key = C.KEY_NUMERIC_B
	KeyNumericC     Key = C.KEY_NUMERIC_C
	KeyNumericD     Key = C.KEY_NUMERIC_D

	KeyCameraFocus Key = C.KEY_CAMERA_FOCUS
	KeyWpsButton   Key = C.KEY_WPS_BUTTON // WiFi Protected Setup key

	KeyTouchpadToggle Key = C.KEY_TOUCHPAD_TOGGLE // Request switch touchpad on or off
	KeyTouchpadOn     Key = C.KEY_TOUCHPAD_ON
	KeyTouchpadOff    Key = C.KEY_TOUCHPAD_OFF

	KeyCameraZoomin  Key = C.KEY_CAMERA_ZOOMIN
	KeyCameraZoomout Key = C.KEY_CAMERA_ZOOMOUT
	KeyCameraUp      Key = C.KEY_CAMERA_UP
	KeyCameraDown    Key = C.KEY_CAMERA_DOWN
	KeyCameraLeft    Key = C.KEY_CAMERA_LEFT
	KeyCameraRight   Key = C.KEY_CAMERA_RIGHT

	KeyAttendantOn     Key = C.KEY_ATTENDANT_ON
	KeyAttendantOff    Key = C.KEY_ATTENDANT_OFF
	KeyAttendantToggle Key = C.KEY_ATTENDANT_TOGGLE // Attendant call on or off
	KeyLightsToggle    Key = C.KEY_LIGHTS_TOGGLE    // Reading light on or off

	ButtonDpadUp    Key = C.BTN_DPAD_UP
	ButtonDpadDown  Key = C.BTN_DPAD_DOWN
	ButtonDpadLeft  Key = C.BTN_DPAD_LEFT
	ButtonDpadRight Key = C.BTN_DPAD_RIGHT

	KeyAlsToggle         Key = C.KEY_ALS_TOGGLE          // Ambient light sensor
	KeyRotateLockToggle  Key = C.KEY_ROTATE_LOCK_TOGGLE  // Display rotation lock
	KeyRefreshRateToggle Key = C.KEY_REFRESH_RATE_TOGGLE // Display refresh rate toggle

	KeyButtonconfig        Key = C.KEY_BUTTONCONFIG          // AL Button Configuration
	KeyTaskmanager         Key = C.KEY_TASKMANAGER           // AL Task/Project Manager
	KeyJournal             Key = C.KEY_JOURNAL               // AL Log/Journal/Timecard
	KeyControlpanel        Key = C.KEY_CONTROLPANEL          // AL Control Panel
	KeyAppselect           Key = C.KEY_APPSELECT             // AL Select Task/Application
	KeyScreensaver         Key = C.KEY_SCREENSAVER           // AL Screen Saver
	KeyVoicecommand        Key = C.KEY_VOICECOMMAND          // Listening Voice Command
	KeyAssistant           Key = C.KEY_ASSISTANT             // AL Context-aware desktop assistant
	KeyKbdLayoutNext       Key = C.KEY_KBD_LAYOUT_NEXT       // AC Next Keyboard Layout Select
	KeyEmojiPicker         Key = C.KEY_EMOJI_PICKER          // Show/hide emoji picker (HUTRR101)
	KeyDictate             Key = C.KEY_DICTATE               // Start or Stop Voice Dictation Session (HUTRR99)
	KeyCameraAccessEnable  Key = C.KEY_CAMERA_ACCESS_ENABLE  // Enables programmatic access to camera devices. (HUTRR72)
	KeyCameraAccessDisable Key = C.KEY_CAMERA_ACCESS_DISABLE // Disables programmatic access to camera devices. (HUTRR72)
	KeyCameraAccessToggle  Key = C.KEY_CAMERA_ACCESS_TOGGLE  // Toggles the current state of the camera access control. (HUTRR72)
	KeyAccessibility       Key = C.KEY_ACCESSIBILITY         // Toggles the system bound accessibility UI/command (HUTRR116)
	KeyDoNotDisturb        Key = C.KEY_DO_NOT_DISTURB        // Toggles the system-wide "Do Not Disturb" control (HUTRR94)

	KeyBrightnessMin Key = C.KEY_BRIGHTNESS_MIN // Set Brightness to Minimum
	KeyBrightnessMax Key = C.KEY_BRIGHTNESS_MAX // Set Brightness to Maximum

	KeyKbdinputassistPrev      Key = C.KEY_KBDINPUTASSIST_PREV
	KeyKbdinputassistNext      Key = C.KEY_KBDINPUTASSIST_NEXT
	KeyKbdinputassistPrevgroup Key = C.KEY_KBDINPUTASSIST_PREVGROUP
	KeyKbdinputassistNextgroup Key = C.KEY_KBDINPUTASSIST_NEXTGROUP
	KeyKbdinputassistAccept    Key = C.KEY_KBDINPUTASSIST_ACCEPT
	KeyKbdinputassistCancel    Key = C.KEY_KBDINPUTASSIST_CANCEL

	// Diagonal movement keys
	KeyRightUp   Key = C.KEY_RIGHT_UP
	KeyRightDown Key = C.KEY_RIGHT_DOWN
	KeyLeftUp    Key = C.KEY_LEFT_UP
	KeyLeftDown  Key = C.KEY_LEFT_DOWN

	KeyRootMenu Key = C.KEY_ROOT_MENU // Show Device's Root Menu
	// Show Top Menu of the Media (e.g. DVD)
	KeyMediaTopMenu Key = C.KEY_MEDIA_TOP_MENU
	KeyNumeric11    Key = C.KEY_NUMERIC_11
	KeyNumeric12    Key = C.KEY_NUMERIC_12

	// Toggle Audio Description: refers to an audio service that helps blind and
	// visually impaired consumers understand the action in a program. Note: in
	// some countries this is referred to as "Video Description".
	KeyAudioDesc    Key = C.KEY_AUDIO_DESC
	Key3dMode       Key = C.KEY_3D_MODE
	KeyNextFavorite Key = C.KEY_NEXT_FAVORITE
	KeyStopRecord   Key = C.KEY_STOP_RECORD
	KeyPauseRecord  Key = C.KEY_PAUSE_RECORD
	KeyVod          Key = C.KEY_VOD // Video on Demand
	KeyUnmute       Key = C.KEY_UNMUTE
	KeyFastreverse  Key = C.KEY_FASTREVERSE
	KeySlowreverse  Key = C.KEY_SLOWREVERSE

	// Control a data application associated with the currently viewed channel,
	// e.g. teletext or data broadcast application (MHEG, MHP, HbbTV, etc.)
	KeyData             Key = C.KEY_DATA
	KeyOnscreenKeyboard Key = C.KEY_ONSCREEN_KEYBOARD
	// Electronic privacy screen control
	KeyPrivacyScreenToggle Key = C.KEY_PRIVACY_SCREEN_TOGGLE

	// Select an area of screen to be copied
	KeySelectiveScreenshot Key = C.KEY_SELECTIVE_SCREENSHOT

	// Move the focus to the next or previous user controllable element within a UI container
	KeyNextElement     Key = C.KEY_NEXT_ELEMENT
	KeyPreviousElement Key = C.KEY_PREVIOUS_ELEMENT

	// Toggle Autopilot engagement
	KeyAutopilotEngageToggle Key = C.KEY_AUTOPILOT_ENGAGE_TOGGLE

	// Shortcut Keys
	KeyMarkWaypoint     Key = C.KEY_MARK_WAYPOINT
	KeySos              Key = C.KEY_SOS
	KeyNavChart         Key = C.KEY_NAV_CHART
	KeyFishingChart     Key = C.KEY_FISHING_CHART
	KeySingleRangeRadar Key = C.KEY_SINGLE_RANGE_RADAR
	KeyDualRangeRadar   Key = C.KEY_DUAL_RANGE_RADAR
	KeyRadarOverlay     Key = C.KEY_RADAR_OVERLAY
	KeyTraditionalSonar Key = C.KEY_TRADITIONAL_SONAR
	KeyClearvuSonar     Key = C.KEY_CLEARVU_SONAR
	KeySidevuSonar      Key = C.KEY_SIDEVU_SONAR
	KeyNavInfo          Key = C.KEY_NAV_INFO
	KeyBrightnessMenu   Key = C.KEY_BRIGHTNESS_MENU

	// Some keyboards have keys which do not have a defined meaning, these keys
	// are intended to be programmed / bound to macros by the user. For most
	// keyboards with these macro-keys the key-sequence to inject, or action to
	// take, is all handled by software on the host side. So from the kernel's
	// point of view these are just normal keys.
	//
	// The KEY_MACRO# codes below are intended for such keys, which may be labeled
	// e.g. G1-G18, or S1 - S30. The KEY_MACRO# codes MUST NOT be used for keys
	// where the marking on the key does indicate a defined meaning / purpose.
	//
	// The KEY_MACRO# codes MUST also NOT be used as fallback for when no existing
	// KEY_FOO define matches the marking / purpose. In this case a new KEY_FOO
	// define MUST be added.
	KeyMacro1  Key = C.KEY_MACRO1
	KeyMacro2  Key = C.KEY_MACRO2
	KeyMacro3  Key = C.KEY_MACRO3
	KeyMacro4  Key = C.KEY_MACRO4
	KeyMacro5  Key = C.KEY_MACRO5
	KeyMacro6  Key = C.KEY_MACRO6
	KeyMacro7  Key = C.KEY_MACRO7
	KeyMacro8  Key = C.KEY_MACRO8
	KeyMacro9  Key = C.KEY_MACRO9
	KeyMacro10 Key = C.KEY_MACRO10
	KeyMacro11 Key = C.KEY_MACRO11
	KeyMacro12 Key = C.KEY_MACRO12
	KeyMacro13 Key = C.KEY_MACRO13
	KeyMacro14 Key = C.KEY_MACRO14
	KeyMacro15 Key = C.KEY_MACRO15
	KeyMacro16 Key = C.KEY_MACRO16
	KeyMacro17 Key = C.KEY_MACRO17
	KeyMacro18 Key = C.KEY_MACRO18
	KeyMacro19 Key = C.KEY_MACRO19
	KeyMacro20 Key = C.KEY_MACRO20
	KeyMacro21 Key = C.KEY_MACRO21
	KeyMacro22 Key = C.KEY_MACRO22
	KeyMacro23 Key = C.KEY_MACRO23
	KeyMacro24 Key = C.KEY_MACRO24
	KeyMacro25 Key = C.KEY_MACRO25
	KeyMacro26 Key = C.KEY_MACRO26
	KeyMacro27 Key = C.KEY_MACRO27
	KeyMacro28 Key = C.KEY_MACRO28
	KeyMacro29 Key = C.KEY_MACRO29
	KeyMacro30 Key = C.KEY_MACRO30

	// Some keyboards with the macro-keys described above have some extra keys
	// for controlling the host-side software responsible for the macro handling:
	// -A macro recording start/stop key. Note that not all keyboards which emit
	//  KEY_MACRO_RECORD_START will also emit KEY_MACRO_RECORD_STOP if
	//  KEY_MACRO_RECORD_STOP is not advertised, then KEY_MACRO_RECORD_START
	//  should be interpreted as a recording start/stop toggle;
	// -Keys for switching between different macro (pre)sets, either a key for
	//  cycling through the configured presets or keys to directly select a preset.
	KeyMacroRecordStart Key = C.KEY_MACRO_RECORD_START
	KeyMacroRecordStop  Key = C.KEY_MACRO_RECORD_STOP
	KeyMacroPresetCycle Key = C.KEY_MACRO_PRESET_CYCLE
	KeyMacroPreset1     Key = C.KEY_MACRO_PRESET1
	KeyMacroPreset2     Key = C.KEY_MACRO_PRESET2
	KeyMacroPreset3     Key = C.KEY_MACRO_PRESET3

	// Some keyboards have a buildin LCD panel where the contents are controlled
	// by the host. Often these have a number of keys directly below the LCD
	// intended for controlling a menu shown on the LCD. These keys often don't
	// have any labeling so we just name them KEY_KBD_LCD_MENU#
	KeyKbdLcdMenu1 Key = C.KEY_KBD_LCD_MENU1
	KeyKbdLcdMenu2 Key = C.KEY_KBD_LCD_MENU2
	KeyKbdLcdMenu3 Key = C.KEY_KBD_LCD_MENU3
	KeyKbdLcdMenu4 Key = C.KEY_KBD_LCD_MENU4
	KeyKbdLcdMenu5 Key = C.KEY_KBD_LCD_MENU5

	ButtonTriggerHappy   Key = C.BTN_TRIGGER_HAPPY
	ButtonTriggerHappy1  Key = C.BTN_TRIGGER_HAPPY1
	ButtonTriggerHappy2  Key = C.BTN_TRIGGER_HAPPY2
	ButtonTriggerHappy3  Key = C.BTN_TRIGGER_HAPPY3
	ButtonTriggerHappy4  Key = C.BTN_TRIGGER_HAPPY4
	ButtonTriggerHappy5  Key = C.BTN_TRIGGER_HAPPY5
	ButtonTriggerHappy6  Key = C.BTN_TRIGGER_HAPPY6
	ButtonTriggerHappy7  Key = C.BTN_TRIGGER_HAPPY7
	ButtonTriggerHappy8  Key = C.BTN_TRIGGER_HAPPY8
	ButtonTriggerHappy9  Key = C.BTN_TRIGGER_HAPPY9
	ButtonTriggerHappy10 Key = C.BTN_TRIGGER_HAPPY10
	ButtonTriggerHappy11 Key = C.BTN_TRIGGER_HAPPY11
	ButtonTriggerHappy12 Key = C.BTN_TRIGGER_HAPPY12
	ButtonTriggerHappy13 Key = C.BTN_TRIGGER_HAPPY13
	ButtonTriggerHappy14 Key = C.BTN_TRIGGER_HAPPY14
	ButtonTriggerHappy15 Key = C.BTN_TRIGGER_HAPPY15
	ButtonTriggerHappy16 Key = C.BTN_TRIGGER_HAPPY16
	ButtonTriggerHappy17 Key = C.BTN_TRIGGER_HAPPY17
	ButtonTriggerHappy18 Key = C.BTN_TRIGGER_HAPPY18
	ButtonTriggerHappy19 Key = C.BTN_TRIGGER_HAPPY19
	ButtonTriggerHappy20 Key = C.BTN_TRIGGER_HAPPY20
	ButtonTriggerHappy21 Key = C.BTN_TRIGGER_HAPPY21
	ButtonTriggerHappy22 Key = C.BTN_TRIGGER_HAPPY22
	ButtonTriggerHappy23 Key = C.BTN_TRIGGER_HAPPY23
	ButtonTriggerHappy24 Key = C.BTN_TRIGGER_HAPPY24
	ButtonTriggerHappy25 Key = C.BTN_TRIGGER_HAPPY25
	ButtonTriggerHappy26 Key = C.BTN_TRIGGER_HAPPY26
	ButtonTriggerHappy27 Key = C.BTN_TRIGGER_HAPPY27
	ButtonTriggerHappy28 Key = C.BTN_TRIGGER_HAPPY28
	ButtonTriggerHappy29 Key = C.BTN_TRIGGER_HAPPY29
	ButtonTriggerHappy30 Key = C.BTN_TRIGGER_HAPPY30
	ButtonTriggerHappy31 Key = C.BTN_TRIGGER_HAPPY31
	ButtonTriggerHappy32 Key = C.BTN_TRIGGER_HAPPY32
	ButtonTriggerHappy33 Key = C.BTN_TRIGGER_HAPPY33
	ButtonTriggerHappy34 Key = C.BTN_TRIGGER_HAPPY34
	ButtonTriggerHappy35 Key = C.BTN_TRIGGER_HAPPY35
	ButtonTriggerHappy36 Key = C.BTN_TRIGGER_HAPPY36
	ButtonTriggerHappy37 Key = C.BTN_TRIGGER_HAPPY37
	ButtonTriggerHappy38 Key = C.BTN_TRIGGER_HAPPY38
	ButtonTriggerHappy39 Key = C.BTN_TRIGGER_HAPPY39
	ButtonTriggerHappy40 Key = C.BTN_TRIGGER_HAPPY40

	KeyMax Key = C.KEY_MAX
)
