package native

import (
	"fmt"
	"testing"
	"unsafe"
)

type _DEBUG_EVENT_64 struct {
	DebugEventCode uint32
	ProcessId      uint32
	ThreadId       uint32
	_              uint32 // to align Union properly
	U              [160]byte
}

type _DEBUG_EVENT_32 struct {
	DebugEventCode uint32
	ProcessId      uint32
	ThreadId       uint32
	U              [84]byte
}

func TestProcess(t *testing.T) {
	var de _DEBUG_EVENT_64
	fmt.Printf("_DEBUG_EVENT_64 Size %d\n", unsafe.Sizeof(de))

	var de2 _DEBUG_EVENT_32
	fmt.Printf("_DEBUG_EVENT_32 Size %d\n", unsafe.Sizeof(de2))

}
