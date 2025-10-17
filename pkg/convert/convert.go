package convert

import (
	"fmt"
)

func IntPtrToStr[T int64 | int32 | int16 | int8 | int | uint64 | uint32 | uint16 | uint8 | uint](ptr *T) string {
	if ptr == nil {
		return "<nil>"
	}
	return fmt.Sprintf("%d", *ptr)
}
