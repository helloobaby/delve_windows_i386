//go:build 386

package winutil

import (
	"errors"
	"unsafe"

	"github.com/go-delve/delve/pkg/dwarf/op"
	"github.com/go-delve/delve/pkg/proc"
)

type FLOATING_SAVE_AREA struct {
	ControlWord   uint32
	StatusWord    uint32
	TagWord       uint32
	ErrorOffset   uint32
	ErrorSelector uint32
	DataOffset    uint32
	DataSelector  uint32
	RegisterArea  [80]byte // 8 * 10 bytes for ST(0)-ST(7)
	Spare0        uint32
}

type I386CONTEXT struct { // 这个暂时跟操作系统的数据结构保持一致
	ContextFlags uint32

	// Debug registers
	Dr0 uint32
	Dr1 uint32
	Dr2 uint32
	Dr3 uint32
	Dr6 uint32
	Dr7 uint32

	// Floating point context (FPU/MMX/SSE)
	FloatSave FLOATING_SAVE_AREA

	// Segment registers
	SegGs uint32
	SegFs uint32
	SegEs uint32
	SegDs uint32

	// General-purpose registers
	Edi uint32
	Esi uint32
	Ebx uint32
	Edx uint32
	Ecx uint32
	Eax uint32

	// Control registers
	Ebp    uint32
	Eip    uint32
	SegCs  uint32 // MUST be sanitized when read from untrusted source
	EFlags uint32
	Esp    uint32
	SegSs  uint32 // MUST be sanitized when read from untrusted source

	// Extended registers (for SSE, only present if CONTEXT_EXTENDED_REGISTERS is set)
	ExtendedRegisters [512]byte // size may vary; 512 bytes is standard for legacy compatibility
}

type I386Registers struct { // 定义为64位方便点
	eax uint64
	ebx uint64
	ecx uint64
	edx uint64
	edi uint64
	esi uint64
	ebp uint64
	esp uint64

	eip    uint64
	eflags uint64
	cs     uint64
	ds     uint64
	es     uint64
	fs     uint64
	gs     uint64
	ss     uint64
	tls    uint64

	Context *I386CONTEXT
}

func (ctx *I386CONTEXT) SetFlags(flags uint32) {
	ctx.ContextFlags = flags
}

func (ctx *I386CONTEXT) SetPC(pc uint64) {
	ctx.Eip = uint32(pc)
}

func (ctx *I386CONTEXT) SetTrap(trap bool) {
	const v = 0x100
	if trap {
		ctx.EFlags |= v
	} else {
		ctx.EFlags &= ^uint32(v)
	}
}

func (ctx *I386CONTEXT) SetReg(regNum uint64, reg *op.DwarfRegister) error {
	return errors.New("Not Implemented")
}

func NewI386Registers(context *I386CONTEXT, TebBaseAddress uint64) *I386Registers {
	regs := &I386Registers{
		eax:    uint64(context.Eax), // Convert uint32 to uint64
		ebx:    uint64(context.Ebx),
		ecx:    uint64(context.Ecx),
		edx:    uint64(context.Edx),
		edi:    uint64(context.Edi),
		esi:    uint64(context.Esi),
		ebp:    uint64(context.Ebp),
		esp:    uint64(context.Esp),
		eip:    uint64(context.Eip),
		eflags: uint64(context.EFlags), // 注意：原 AMD64 版本是 uint64(context.EFlags)，这里保持一致
		cs:     uint64(context.SegCs),
		ds:     uint64(context.SegDs), // Added DS, ES, SS as they exist in i386 context
		es:     uint64(context.SegEs),
		fs:     uint64(context.SegFs),
		gs:     uint64(context.SegGs),
		ss:     uint64(context.SegSs),
		tls:    TebBaseAddress, // TLS (Thread Local Storage) often points to TEB (Thread Environment Block) on Windows
	}

	// Note: If I386Registers has a floating-point state field (e.g., fltSave) similar to AMD64Registers,
	// you would assign it here, e.g.:
	// regs.fltSave = &context.FloatSave // Assuming FloatSaveData type exists and matches expectation

	regs.Context = context // Store the original context pointer
	return regs
}

func (r *I386Registers) PC() uint64 {
	return r.eip
}

// SP returns the stack pointer location,
// i.e. the RSP register.
func (r *I386Registers) SP() uint64 {
	return r.esp
}

func (r *I386Registers) BP() uint64 {
	return r.ebp
}

// LR returns the link register.
func (r *I386Registers) LR() uint64 {
	return 0
}

// TLS returns the value of the register
// that contains the location of the thread
// local storage segment.
func (r *I386Registers) TLS() uint64 {
	return r.tls
}

// GAddr returns the address of the G variable if it is known, 0 and false
// otherwise.
func (r *I386Registers) GAddr() (uint64, bool) {
	return 0, false
}

func (r *I386Registers) Slice(floatingPoint bool) ([]proc.Register, error) {
	return nil, errors.New("Not implemented")
}

func NewI386CONTEXT() *I386CONTEXT {
	var c *I386CONTEXT
	buf := make([]byte, unsafe.Sizeof(*c)+15)
	return (*I386CONTEXT)(unsafe.Pointer((uintptr(unsafe.Pointer(&buf[15]))) &^ 15))
}

func (r *I386Registers) Copy() (proc.Registers, error) {
	var rr I386Registers
	rr = *r
	rr.Context = NewI386CONTEXT()
	*(rr.Context) = *(r.Context)
	return &rr, nil
}
