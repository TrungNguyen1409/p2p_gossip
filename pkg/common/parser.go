package common

import (
	"strconv"
)

func ParseUint8(s string) (uint8, error) {
	strconv.ParseUint(s, 10, 8)
}

func ParseUint16(s string) (uint8, error) {

}
