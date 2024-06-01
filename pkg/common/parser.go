package common

import (
	"strconv"
)

func ParseUint8(s string) (uint8, error) {
	parseUint, err := strconv.ParseUint(s, 10, 8)
	if err != nil {
		return 0, err
	}

	return uint8(parseUint), nil
}

func ParseUint16(s string) (uint16, error) {
	parseUint, err := strconv.ParseUint(s, 10, 16)
	if err != nil {
		return 0, err
	}

	return uint16(parseUint), nil
}
