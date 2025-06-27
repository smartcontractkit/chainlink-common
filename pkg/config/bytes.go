package config

import (
	"fmt"
	"regexp"
	"strconv"
	"strings"
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

func (b Size) MarshalText() ([]byte, error) {
	if b >= TByte {
		return []byte(fmt.Sprintf("%.2ftb", float64(b)/float64(TByte))), nil
	} else if b >= GByte {
		return []byte(fmt.Sprintf("%.2fgb", float64(b)/float64(GByte))), nil
	} else if b >= MByte {
		return []byte(fmt.Sprintf("%.2fmb", float64(b)/float64(MByte))), nil
	} else if b >= KByte {
		return []byte(fmt.Sprintf("%.2fkb", float64(b)/float64(KByte))), nil
	}
	return []byte(fmt.Sprintf("%db", b)), nil
}

func ParseByte(s string) (b Size, err error) {
	err = b.UnmarshalText([]byte(s))
	return
}

// UnmarshalText parses a file size from bs in to s.
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

	value, err := strconv.ParseFloat(num, 64)
	if err != nil {
		return fmt.Errorf(`bad filesize value: "%v"`, string(bs))
	}

	u, ok := fsUnitMap[unit]
	if !ok {
		return fmt.Errorf(`bad filesize unit: "%v"`, unit)
	}

	*b = Size(value * float64(u))
	return nil
}

func (b Size) String() string {
	str, _ := b.MarshalText()
	return string(str)
}
