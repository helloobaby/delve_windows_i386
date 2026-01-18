package native

import (
	"github.com/go-delve/delve/pkg/proc/winutil"
)

const (
	_CONTEXT_i386            = 0x00010000
	_CONTEXT_CONTROL         = (_CONTEXT_i386 | 0x1)
	_CONTEXT_INTEGER         = (_CONTEXT_i386 | 0x2)
	_CONTEXT_SEGMENTS        = (_CONTEXT_i386 | 0x4)
	_CONTEXT_FLOATING_POINT  = (_CONTEXT_i386 | 0x8)
	_CONTEXT_DEBUG_REGISTERS = (_CONTEXT_i386 | 0x10)
	_CONTEXT_FULL            = (_CONTEXT_CONTROL | _CONTEXT_INTEGER | _CONTEXT_SEGMENTS)
	_CONTEXT_ALL             = (_CONTEXT_CONTROL | _CONTEXT_INTEGER | _CONTEXT_SEGMENTS | _CONTEXT_FLOATING_POINT | _CONTEXT_DEBUG_REGISTERS)
)

type _CONTEXT = winutil.I386CONTEXT

type _DEBUG_EVENT struct {
	DebugEventCode uint32
	ProcessId      uint32
	ThreadId       uint32
	U              [84]byte
}
