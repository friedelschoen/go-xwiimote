package miidata

import (
	"encoding/binary"
	"unicode/utf16"
)

const (
	slotOffset = 0x08
	slotSize   = 0x4A
)

type Mii struct {
	ParadeMii bool

	Invalid    bool
	Female     bool
	BirthMonth int
	BirthDay   int
	FavColor   int
	IsFavorite bool

	Name string

	Height int
	Weight int

	ID int

	SystemID0 int
	SystemID1 int
	SystemID2 int
	SystemID3 int

	FaceShape     int
	SkinColor     int
	FacialFeature int

	MingleOff  bool
	Downloaded bool

	HairType  int
	HairColor int
	HairPart  int

	EyebrowType         int
	EyebrowRotation     int
	EyebrowColor        int
	EyebrowSize         int
	EyebrowVertPos      int
	EyebrowHorizSpacing int

	EyeType int

	EyeRotation     int
	EyeVertPos      int
	EyeColor        int
	EyeSize         int
	EyeHorizSpacing int

	NoseType    int
	NoseSize    int
	NoseVertPos int

	LipType    int
	LipColor   int
	LipSize    int
	LipVertPos int

	GlassesType    int
	GlassesColor   int
	GlassesSize    int
	GlassesVertPos int

	MustacheType    int
	BeardType       int
	FacialHairColor int
	MustacheSize    int
	MustacheVertPos int

	MoleOn       bool
	MoleSize     int
	MoleVertPos  int
	MoleHorizPos int

	CreatorName string
}

func utf16name(payload []byte) string {
	// 10 uint16 code units
	var cu [10]uint16
	for i := 0; i < 10; i++ {
		cu[i] = binary.BigEndian.Uint16(payload[i*2 : i*2+2])
	}
	// trim at first 0x0000
	n := 0
	for n < len(cu) && cu[n] != 0 {
		n++
	}
	return string(utf16.Decode(cu[:n]))
}

func DecodeMii(payload []byte) (out Mii) {
	if len(payload) < 0x4A {
		return out
	}

	// Helpers: decode big-endian words and bitfields MSB-first.
	u16 := func(off int) int { return int(binary.BigEndian.Uint16(payload[off : off+2])) }
	u32 := func(off int) int { return int(binary.BigEndian.Uint32(payload[off : off+4])) }

	/* addr 0x00-0x01 */
	w0 := u16(0x00)
	out.Invalid = ((w0 >> 15) & 0x1) != 0
	out.Female = ((w0 >> 14) & 0x1) != 0
	out.BirthMonth = (w0 >> 10) & 0xF
	out.BirthDay = (w0 >> 5) & 0x1F
	out.FavColor = (w0 >> 1) & 0xF
	out.IsFavorite = (w0 & 0x1) != 0

	/* addr 0x02-0x15 */
	out.Name = utf16name(payload[0x02:0x16])

	/* addr 0x16-0x17 */
	out.Height = int(payload[0x16])
	out.Weight = int(payload[0x17])

	/* addr 0x18-0x1B */
	out.ID = u32(0x18)

	/* addr 0x1C-0x1F */
	out.SystemID0 = int(payload[0x1C])
	out.SystemID1 = int(payload[0x1D])
	out.SystemID2 = int(payload[0x1E])
	out.SystemID3 = int(payload[0x1F])

	/* addr 0x20-0x21 */
	w1 := u16(0x20)
	out.FaceShape = (w1 >> 13) & 0x7
	out.SkinColor = (w1 >> 10) & 0x7
	out.FacialFeature = (w1 >> 6) & 0xF
	// unknown0: (w1 >> 3) & 0x7
	out.MingleOff = ((w1 >> 2) & 0x1) != 0
	// unknown1: (w1 >> 1) & 0x1
	out.Downloaded = (w1 & 0x1) != 0

	/* addr 0x22-0x23 */
	w2 := u16(0x22)
	out.HairType = (w2 >> 9) & 0x7F
	out.HairColor = (w2 >> 6) & 0x7
	out.HairPart = (w2 >> 5) & 0x1
	// unknown2: w2 & 0x1F

	/* addr 0x24-0x27 */
	d0 := u32(0x24)
	out.EyebrowType = (d0 >> 27) & 0x1F
	// unknown3: (d0 >> 26) & 0x1
	out.EyebrowRotation = (d0 >> 22) & 0xF
	// unknown4: (d0 >> 16) & 0x3F
	out.EyebrowColor = (d0 >> 13) & 0x7
	out.EyebrowSize = (d0 >> 9) & 0xF
	out.EyebrowVertPos = (d0 >> 4) & 0x1F
	out.EyebrowHorizSpacing = d0 & 0xF

	/* addr 0x28-0x2B */
	d1 := u32(0x28)
	out.EyeType = (d1 >> 26) & 0x3F
	// unknown5: (d1 >> 24) & 0x3
	out.EyeRotation = (d1 >> 21) & 0x7
	out.EyeVertPos = (d1 >> 16) & 0x1F
	out.EyeColor = (d1 >> 13) & 0x7
	// unknown6: (d1 >> 12) & 0x1
	out.EyeSize = (d1 >> 9) & 0x7
	out.EyeHorizSpacing = (d1 >> 5) & 0xF
	// unknown7: d1 & 0x1F

	/* addr 0x2C-0x2D */
	w3 := u16(0x2C)
	out.NoseType = (w3 >> 12) & 0xF
	out.NoseSize = (w3 >> 8) & 0xF
	out.NoseVertPos = (w3 >> 3) & 0x1F
	// unknown8: w3 & 0x7

	/* addr 0x2E-0x2F */
	w4 := u16(0x2E)
	out.LipType = (w4 >> 11) & 0x1F
	out.LipColor = (w4 >> 9) & 0x3
	out.LipSize = (w4 >> 5) & 0xF
	out.LipVertPos = w4 & 0x1F

	/* addr 0x30-0x31 */
	w5 := u16(0x30)
	out.GlassesType = (w5 >> 12) & 0xF
	out.GlassesColor = (w5 >> 9) & 0x7
	// unknown9: (w5 >> 8) & 0x1
	out.GlassesSize = (w5 >> 5) & 0x7
	out.GlassesVertPos = w5 & 0x1F

	/* addr 0x32-0x33 */
	w6 := u16(0x32)
	out.MustacheType = (w6 >> 14) & 0x3
	out.BeardType = (w6 >> 12) & 0x3
	out.FacialHairColor = (w6 >> 9) & 0x7
	out.MustacheSize = (w6 >> 5) & 0xF
	out.MustacheVertPos = w6 & 0x1F

	/* addr 0x34-0x35 */
	w7 := u16(0x34)
	out.MoleOn = ((w7 >> 15) & 0x1) != 0
	out.MoleSize = (w7 >> 11) & 0xF
	out.MoleVertPos = (w7 >> 6) & 0x1F
	out.MoleHorizPos = (w7 >> 1) & 0x1F
	// unknown: w7 & 0x1

	/* addr 0x36-0x3f */
	out.CreatorName = utf16name(payload[0x36:0x40])

	return out
}

func DecodeBlock(block []byte) []Mii {
	parade := binary.BigEndian.Uint16(block[4:6])
	var result [10]Mii
	var n int
	for i := range result {
		result[n] = DecodeMii(block[slotOffset+i*slotSize : slotOffset+(i+1)*slotSize])
		if result[n].Name == "" {
			continue
		}
		result[n].ParadeMii = parade&(1<<i) != 0
		n++
	}
	return result[:n]
}
