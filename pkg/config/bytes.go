package config

import (
	"fmt"
	"regexp"
	"strings"

	"github.com/shopspring/decimal"
)

// Size represents size in bytes, and formats with a suffix.
type Size int

const (
	Byte  = Size(1)
	KByte = 1000 * Byte
	MByte = 1000 * KByte
	GByte = 1000 * MByte
	TByte = 1000 * GByte
)

var (
	fsregex = regexp.MustCompile(`(\d+\.?\d*)(tb|gb|mb|kb|b)?`)

	fsUnitMap = map[string]Size{
		"tb": TByte,
		"gb": GByte,
		"mb": MByte,
		"kb": KByte,
		"b":  1,
		"":   1,
	}
)

// SizeOf returns the cumulative len of the byte slice arguments as a Size.
func SizeOf[B ~[]byte](bs ...B) (s Size) {
	for _, b := range bs {
		s += Size(len(b))
	}
	return
}

func (b Size) MarshalText() ([]byte, error) {
	if b >= TByte {
		d := decimal.NewFromInt(int64(b)).Div(decimal.NewFromInt(int64(TByte)))
		return []byte(fmt.Sprintf("%stb", d)), nil
	} else if b >= GByte {
		d := decimal.NewFromInt(int64(b)).Div(decimal.NewFromInt(int64(GByte)))
		return []byte(fmt.Sprintf("%sgb", d)), nil
	} else if b >= MByte {
		d := decimal.NewFromInt(int64(b)).Div(decimal.NewFromInt(int64(MByte)))
		return []byte(fmt.Sprintf("%smb", d)), nil
	} else if b >= KByte {
		d := decimal.NewFromInt(int64(b)).Div(decimal.NewFromInt(int64(KByte)))
		return []byte(fmt.Sprintf("%skb", d)), nil
	}
	return []byte(fmt.Sprintf("%db", b)), nil
}

func ParseByte(s string) (b Size, err error) {
	err = b.UnmarshalText([]byte(s))
	return
}

// UnmarshalText parses a size from bs in to s.
func (b *Size) UnmarshalText(bs []byte) error {
	lc := strings.ToLower(strings.TrimSpace(string(bs)))
	matches := fsregex.FindAllStringSubmatch(lc, -1)
	if len(matches) != 1 || len(matches[0]) != 3 || fmt.Sprintf("%s%s", matches[0][1], matches[0][2]) != lc {
		return fmt.Errorf(`bad filesize expression: "%v"`, string(bs))
	}

	var (
		num  = matches[0][1]
		unit = matches[0][2]
	)

	d, err := decimal.NewFromString(num)
	if err != nil {
		return fmt.Errorf(`invalid size: "%v"`, string(bs))
	}

	u, ok := fsUnitMap[unit]
	if !ok {
		return fmt.Errorf(`bad filesize unit: "%v"`, unit)
	}

	d = d.Mul(decimal.NewFromInt(int64(u)))
	if !d.IsInteger() {
		return fmt.Errorf(`invalid size: must be whole integer in bytes: "%v"`, string(bs))
	}

	*b = Size(d.IntPart())
	return nil
}

func (b Size) String() string {
	str, _ := b.MarshalText()
	return string(str)
}
