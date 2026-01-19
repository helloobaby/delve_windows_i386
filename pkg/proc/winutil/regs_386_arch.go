//go:build 386

package winutil

import (
	"encoding/binary"
	"fmt"
	"unsafe"

	"github.com/go-delve/delve/pkg/dwarf/op"
	"github.com/go-delve/delve/pkg/dwarf/regnum"
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

type I386CONTEXT struct { // 这个必须跟操作系统的C数据结构保持一致
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
	var p *uint32

	switch regNum {
	case regnum.I386_Eax:
		p = &ctx.Eax
	case regnum.I386_Ecx:
		p = &ctx.Ecx
	case regnum.I386_Edx:
		p = &ctx.Edx
	case regnum.I386_Ebx:
		p = &ctx.Ebx
	case regnum.I386_Esp:
		p = &ctx.Esp
	case regnum.I386_Ebp:
		p = &ctx.Ebp
	case regnum.I386_Esi:
		p = &ctx.Esi
	case regnum.I386_Edi:
		p = &ctx.Edi
	case regnum.I386_Eip:
		p = &ctx.Eip
	case regnum.I386_Eflags:
		p = &ctx.EFlags
	case regnum.I386_Cs:
		p = &ctx.SegCs
	case regnum.I386_Ds:
		p = &ctx.SegDs
	case regnum.I386_Es:
		p = &ctx.SegEs
	case regnum.I386_Fs:
		p = &ctx.SegFs
	case regnum.I386_Gs:
		p = &ctx.SegGs
	case regnum.I386_Ss:
		p = &ctx.SegSs
	}
	if p != nil {
		// 校验字节长度，32位寄存器应该是 4 字节
		// 注意：有些 DWARF 实现可能会传 8 字节，这里取低 32 位
		if reg.Bytes != nil && len(reg.Bytes) != 4 && len(reg.Bytes) != 8 {
			return fmt.Errorf("wrong number of bytes for register %d (%d)", regNum, len(reg.Bytes))
		}
		*p = uint32(reg.Uint64Val)
		return nil
	}

	// 处理 XMM 寄存器 (I386 通常有 8 个 XMM 寄存器: XMM0-XMM7)
	if regNum >= regnum.I386_XMM0 && regNum <= regnum.I386_XMM0+7 {
		reg.FillBytes()
		if len(reg.Bytes) > 16 {
			return fmt.Errorf("too many bytes when setting register XMM%d", regNum-regnum.I386_XMM0)
		}
		// 计算在 ExtendedRegisters 中的偏移
		// 注意：Windows I386CONTEXT 的浮点寄存器通常存在 ExtendedRegisters 字段中
		// 具体的偏移量依赖于具体的结构体定义，以下是通用写法：
		idx := (regNum - regnum.I386_XMM0) * 16
		copy(ctx.ExtendedRegisters[idx:], reg.Bytes)
		return nil
	}

	return fmt.Errorf("can not set register %d (unsupported or read-only)", regNum)
}

// 从系统标准的CONTEXT结构体转化
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
		tls:    TebBaseAddress,         // TLS (Thread Local Storage) often points to TEB (Thread Environment Block) on Windows
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
	var regs = []struct {
		k string
		v uint64
	}{
		{"Eip", r.eip},
		{"Esp", r.esp},
		{"Eax", r.eax},
		{"Ebx", r.ebx},
		{"Ecx", r.ecx},
		{"Edx", r.edx},
		{"Edi", r.edi},
		{"Esi", r.esi},
		{"Ebp", r.ebp},
		{"Eflags", r.eflags},
		{"TLS", r.tls},
	}

	outlen := len(regs)
	// 判断是否存在 Context 来提取浮点寄存器
	hasFloat := floatingPoint && r.Context != nil
	if hasFloat {
		// 8个 ST 寄存器 + 8个 XMM 寄存器 + 状态寄存器
		outlen += 8 + 8 + 8
	}

	out := make([]proc.Register, 0, outlen)
	for _, reg := range regs {
		out = proc.AppendUint64Register(out, reg.k, reg.v)
	}

	if hasFloat {
		// 从 Context.ExtendedRegisters 中解析 FXSAVE 结构
		// Windows I386CONTEXT 的 ExtendedRegisters 字段通常是 [512]byte
		// 格式遵循 x86 FXSAVE 标准

		f := r.Context.ExtendedRegisters // 假设类型是 [512]byte

		// 辅助函数：从字节数组提取 uint16/uint32
		get16 := func(off int) uint64 { return uint64(binary.LittleEndian.Uint16(f[off : off+2])) }
		get32 := func(off int) uint64 { return uint64(binary.LittleEndian.Uint32(f[off : off+4])) }

		out = proc.AppendUint64Register(out, "CW", get16(0))
		out = proc.AppendUint64Register(out, "SW", get16(2))
		out = proc.AppendUint64Register(out, "TW", uint64(f[4])) // Tag Word
		out = proc.AppendUint64Register(out, "FOP", get16(6))
		out = proc.AppendUint64Register(out, "FIP", get32(8))
		out = proc.AppendUint64Register(out, "FDP", get32(16))
		out = proc.AppendUint64Register(out, "MXCSR", get32(24))
		out = proc.AppendUint64Register(out, "MXCSR_MASK", get32(28))

		// 8个 ST 寄存器 (每个16字节，实际有效通常是10字节)
		for i := 0; i < 8; i++ {
			off := 32 + (i * 16)
			out = proc.AppendBytesRegister(out, fmt.Sprintf("ST(%d)", i), f[off:off+10])
		}

		// 8个 XMM 寄存器 (每个16字节)
		for i := 0; i < 8; i++ {
			off := 160 + (i * 16)
			out = proc.AppendBytesRegister(out, fmt.Sprintf("XMM%d", i), f[off:off+16])
		}
	}

	return out, nil
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
