package winutil

import (
	"fmt"
	"testing"
	"unsafe"
)

func Test(t *testing.T) {
	fmt.Printf("Size of i368 CONTEXT %d\n", unsafe.Sizeof(I386CONTEXT{})) // 必须是716
	fmt.Printf("Offset Eip %d\n", unsafe.Offsetof(I386CONTEXT{}.Eip))     // 必须是184
}
