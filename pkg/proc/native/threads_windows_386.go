package native

import (
	"errors"
	"fmt"
	"unsafe"

	"github.com/go-delve/delve/pkg/proc"
	"github.com/go-delve/delve/pkg/proc/amd64util"
	"github.com/go-delve/delve/pkg/proc/winutil"
)

func newContext() *winutil.I386CONTEXT {
	return &winutil.I386CONTEXT{}
}

func registers(t *nativeThread) (proc.Registers, error) {
	context := newContext()

	context.SetFlags(_CONTEXT_ALL)
	err := t.getContext(context)
	if err != nil {
		return nil, err
	}

	var threadInfo _THREAD_BASIC_INFORMATION
	status := _NtQueryInformationThread(t.os.hThread, _ThreadBasicInformation, uintptr(unsafe.Pointer(&threadInfo)), uint32(unsafe.Sizeof(threadInfo)), nil)
	if !_NT_SUCCESS(status) {
		return nil, fmt.Errorf("NtQueryInformationThread failed: it returns 0x%x", status)
	}

	return winutil.NewI386Registers(context, uint64(threadInfo.TebBaseAddress)), nil
}

func (t *nativeThread) setContext(context *winutil.I386CONTEXT) error {
	return _SetThreadContext(t.os.hThread, context)
}

func (t *nativeThread) getContext(context *winutil.I386CONTEXT) error {
	return _GetThreadContext(t.os.hThread, context)
}

func (t *nativeThread) restoreRegisters(savedRegs proc.Registers) error {
	return t.setContext(savedRegs.(*winutil.I386Registers).Context)
}

func (t *nativeThread) withDebugRegisters(f func(*amd64util.DebugRegisters) error) error {
	if !enableHardwareBreakpoints {
		return errors.New("hardware breakpoints not supported")
	}

	return nil
}
