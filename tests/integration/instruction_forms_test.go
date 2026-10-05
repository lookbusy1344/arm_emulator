package integration

import (
	"fmt"
	"strings"
	"testing"

	"github.com/lookbusy1344/arm-emulator/loader"
	"github.com/lookbusy1344/arm-emulator/parser"
	"github.com/lookbusy1344/arm-emulator/vm"
)

// Each form assembles a few instructions through the parser, encoder and loader, runs
// them, and checks registers, flags and memory. R10 holds the address of buf.
//
//	buf+0:  0x11223344   buf+4:  0x55667788   buf+8:  0x99AABBCC   buf+12: 0xDDEEFF00
//	buf+16..buf+31: zero
type instructionForm struct {
	name    string
	code    string
	regs    map[int]uint32 // expected register values
	bufRegs map[int]uint32 // expected register values as offsets from buf
	flags   *vm.CPSR       // expected NZCV, when set
	mem     map[uint32]uint32
}

const (
	stackTop  = vm.StackSegmentStart + vm.StackSegmentSize
	setC      = "MSR CPSR_f, #0x20000000\n"
	clearNZCV = "MSR CPSR_f, #0\n"
)

func formProgram(code string) string {
	var b strings.Builder
	b.WriteString(".org 0x8000\n_start:\n    LDR R10, =buf\n")
	for line := range strings.SplitSeq(strings.TrimSpace(code), "\n") {
		if strings.HasSuffix(line, ":") {
			b.WriteString(line + "\n")
		} else {
			b.WriteString("    " + line + "\n")
		}
	}
	b.WriteString("    SWI #0\n    .ltorg\nbuf:\n" +
		"    .word 0x11223344, 0x55667788, 0x99AABBCC, 0xDDEEFF00\n" +
		"    .word 0, 0, 0, 0\n")
	return b.String()
}

func runForm(t *testing.T, code string) (*vm.VM, uint32) {
	t.Helper()
	const cycleLimit = 10000
	src := formProgram(code)
	program, err := parser.NewParser(src, "form.s").Parse()
	if err != nil {
		t.Fatalf("parse: %v\n%s", err, src)
	}
	buf, err := program.SymbolTable.Get("buf")
	if err != nil {
		t.Fatal(err)
	}
	machine := vm.NewVM()
	machine.CycleLimit = cycleLimit
	if err := machine.InitializeStack(stackTop); err != nil {
		t.Fatal(err)
	}
	if err := loader.LoadProgramIntoVM(machine, program, vm.CodeSegmentStart); err != nil {
		t.Fatalf("load: %v\n%s", err, src)
	}
	if err := machine.Run(); err != nil && machine.State != vm.StateHalted {
		t.Fatalf("run: %v\n%s", err, src)
	}
	return machine, buf
}

func TestInstructionForms(t *testing.T) {
	for _, f := range instructionForms {
		t.Run(f.name, func(t *testing.T) {
			m, buf := runForm(t, f.code)
			for r, want := range f.regs {
				if got := m.CPU.R[r]; got != want {
					t.Errorf("R%d = 0x%08X, want 0x%08X", r, got, want)
				}
			}
			for r, off := range f.bufRegs {
				if got, want := m.CPU.R[r], buf+off; got != want {
					t.Errorf("R%d = 0x%08X, want buf+%d (0x%08X)", r, got, off, want)
				}
			}
			if f.flags != nil && m.CPU.CPSR != *f.flags {
				t.Errorf("flags = %+v, want %+v", m.CPU.CPSR, *f.flags)
			}
			for off, want := range f.mem {
				got, err := m.Memory.ReadWord(buf + off)
				if err != nil || got != want {
					t.Errorf("[buf+%d] = 0x%08X (%v), want 0x%08X", off, got, err, want)
				}
			}
		})
	}
}

var instructionForms = []instructionForm{
	// Data processing, register operand
	{name: "AND", code: "MOV R1, #0xF0\nMOV R2, #0x3C\nAND R0, R1, R2", regs: map[int]uint32{0: 0x30}},
	{name: "EOR", code: "MOV R1, #0xF0\nMOV R2, #0x3C\nEOR R0, R1, R2", regs: map[int]uint32{0: 0xCC}},
	{name: "SUB", code: "MOV R1, #10\nMOV R2, #3\nSUB R0, R1, R2", regs: map[int]uint32{0: 7}},
	{name: "RSB", code: "MOV R1, #10\nMOV R2, #3\nRSB R0, R1, R2", regs: map[int]uint32{0: 0xFFFFFFF9}},
	{name: "ADD", code: "MOV R1, #10\nMOV R2, #3\nADD R0, R1, R2", regs: map[int]uint32{0: 13}},
	{name: "ADC carry set", code: "MOV R1, #10\nMOV R2, #3\n" + setC + "ADC R0, R1, R2", regs: map[int]uint32{0: 14}},
	{name: "SBC carry clear", code: "MOV R1, #10\nMOV R2, #3\n" + clearNZCV + "SBC R0, R1, R2", regs: map[int]uint32{0: 6}},
	{name: "RSC carry clear", code: "MOV R1, #10\nMOV R2, #3\n" + clearNZCV + "RSC R0, R1, R2", regs: map[int]uint32{0: 0xFFFFFFF8}},
	{name: "ORR", code: "MOV R1, #0xF0\nORR R0, R1, #0x0F", regs: map[int]uint32{0: 0xFF}},
	{name: "MOV register", code: "MOV R1, #77\nMOV R0, R1", regs: map[int]uint32{0: 77}},
	{name: "BIC", code: "MOV R1, #0xFF\nBIC R0, R1, #0x0F", regs: map[int]uint32{0: 0xF0}},
	{name: "MVN", code: "MVN R0, #0", regs: map[int]uint32{0: 0xFFFFFFFF}},

	// Compare instructions set flags without writing a register
	{name: "TST zero", code: clearNZCV + "MOV R1, #0xF0\nTST R1, #0x0F", flags: &vm.CPSR{Z: true}},
	{name: "TEQ equal", code: clearNZCV + "MOV R1, #5\nTEQ R1, #5", flags: &vm.CPSR{Z: true}},
	{name: "CMP less", code: clearNZCV + "MOV R1, #3\nCMP R1, #5", flags: &vm.CPSR{N: true}},
	{name: "CMP equal", code: clearNZCV + "MOV R1, #5\nCMP R1, #5", flags: &vm.CPSR{Z: true, C: true}},
	{name: "CMN wraps to zero", code: clearNZCV + "MVN R1, #0\nCMN R1, #1", flags: &vm.CPSR{Z: true, C: true}},

	// Rotated immediates
	{name: "immediate top byte", code: "MOV R0, #0xFF000000", regs: map[int]uint32{0: 0xFF000000}},
	{name: "immediate shifted", code: "MOV R0, #0x3FC", regs: map[int]uint32{0: 0x3FC}},
	{name: "immediate via MVN", code: "MOV R0, #-1", regs: map[int]uint32{0: 0xFFFFFFFF}},
	{name: "immediate carry out", code: clearNZCV + "MOVS R0, #0x80000000", regs: map[int]uint32{0: 0x80000000},
		flags: &vm.CPSR{N: true, C: true}},

	// Shifts by immediate
	{name: "LSL #4", code: "MOV R1, #0x81\nMOV R0, R1, LSL #4", regs: map[int]uint32{0: 0x810}},
	{name: "LSR #4", code: "MOV R1, #0x81\nMOV R0, R1, LSR #4", regs: map[int]uint32{0: 0x8}},
	{name: "ASR #4", code: "MVN R1, #0xF\nMOV R0, R1, ASR #4", regs: map[int]uint32{0: 0xFFFFFFFF}},
	{name: "ROR #4", code: "MOV R1, #0x81\nMOV R0, R1, ROR #4", regs: map[int]uint32{0: 0x10000008}},
	{name: "RRX", code: "MOV R1, #2\n" + setC + "MOV R0, R1, RRX", regs: map[int]uint32{0: 0x80000001}},
	{name: "LSR #32", code: clearNZCV + "MVN R1, #0\nMOVS R0, R1, LSR #32", regs: map[int]uint32{0: 0},
		flags: &vm.CPSR{Z: true, C: true}},
	{name: "ASR #32", code: "MOV R1, #0x80000000\nMOV R0, R1, ASR #32", regs: map[int]uint32{0: 0xFFFFFFFF}},
	{name: "shifted operand in ADD", code: "MOV R1, #3\nADD R0, R1, R1, LSL #2", regs: map[int]uint32{0: 15}},

	// Shifts by register
	{name: "LSL by register", code: "MOV R1, #1\nMOV R3, #5\nMOV R0, R1, LSL R3", regs: map[int]uint32{0: 32}},
	{name: "ASR by register", code: "MOV R1, #0x80000000\nMOV R3, #4\nMOV R0, R1, ASR R3", regs: map[int]uint32{0: 0xF8000000}},
	{name: "SUB with ROR by register", code: "MOV R1, #100\nMOV R2, #0x10\nMOV R3, #4\nSUB R0, R1, R2, ROR R3", regs: map[int]uint32{0: 99}},

	// S suffix
	{name: "MOVS zero", code: clearNZCV + "MOVS R0, #0", flags: &vm.CPSR{Z: true}},
	{name: "ADDS signed overflow", code: clearNZCV + "MVN R1, #0x80000000\nADDS R0, R1, #1",
		regs: map[int]uint32{0: 0x80000000}, flags: &vm.CPSR{N: true, V: true}},
	{name: "SUBS borrow", code: clearNZCV + "MOV R1, #0\nSUBS R0, R1, #1",
		regs: map[int]uint32{0: 0xFFFFFFFF}, flags: &vm.CPSR{N: true}},

	// Word and byte loads
	{name: "LDR base", code: "LDR R0, [R10]", regs: map[int]uint32{0: 0x11223344}},
	{name: "LDR immediate", code: "LDR R0, [R10, #4]", regs: map[int]uint32{0: 0x55667788}},
	{name: "LDR negative immediate", code: "ADD R1, R10, #8\nLDR R0, [R1, #-4]", regs: map[int]uint32{0: 0x55667788}},
	{name: "LDR register", code: "MOV R2, #8\nLDR R0, [R10, R2]", regs: map[int]uint32{0: 0x99AABBCC}},
	{name: "LDR negative register", code: "ADD R1, R10, #12\nMOV R2, #4\nLDR R0, [R1, -R2]", regs: map[int]uint32{0: 0x99AABBCC}},
	{name: "LDR scaled register", code: "MOV R2, #3\nLDR R0, [R10, R2, LSL #2]", regs: map[int]uint32{0: 0xDDEEFF00}},
	{name: "LDR pre-indexed write-back", code: "MOV R1, R10\nLDR R0, [R1, #4]!",
		regs: map[int]uint32{0: 0x55667788}, bufRegs: map[int]uint32{1: 4}},
	{name: "LDR post-indexed", code: "MOV R1, R10\nLDR R0, [R1], #8",
		regs: map[int]uint32{0: 0x11223344}, bufRegs: map[int]uint32{1: 8}},
	{name: "LDR post-indexed negative", code: "ADD R1, R10, #4\nLDR R0, [R1], #-4",
		regs: map[int]uint32{0: 0x55667788}, bufRegs: map[int]uint32{1: 0}},
	{name: "LDR post-indexed register", code: "MOV R1, R10\nMOV R2, #12\nLDR R0, [R1], R2",
		regs: map[int]uint32{0: 0x11223344}, bufRegs: map[int]uint32{1: 12}},
	{name: "LDRB", code: "LDRB R0, [R10, #1]", regs: map[int]uint32{0: 0x33}},
	{name: "LDRB post-indexed", code: "MOV R1, R10\nLDRB R0, [R1], #3",
		regs: map[int]uint32{0: 0x44}, bufRegs: map[int]uint32{1: 3}},
	{name: "LDR PC-relative label", code: "LDR R0, buf", regs: map[int]uint32{0: 0x11223344}},
	{name: "LDR literal", code: "LDR R0, =0x12345678", regs: map[int]uint32{0: 0x12345678}},
	{name: "LDR literal of label", code: "LDR R0, =buf", bufRegs: map[int]uint32{0: 0}},
	{name: "ADR", code: "ADR R0, buf", bufRegs: map[int]uint32{0: 0}},

	// Word and byte stores
	{name: "STR immediate", code: "LDR R0, =0xCAFEBABE\nSTR R0, [R10, #16]", mem: map[uint32]uint32{16: 0xCAFEBABE}},
	{name: "STR pre-indexed write-back", code: "LDR R0, =0xCAFEBABE\nMOV R1, R10\nSTR R0, [R1, #20]!",
		mem: map[uint32]uint32{20: 0xCAFEBABE}, bufRegs: map[int]uint32{1: 20}},
	{name: "STR post-indexed", code: "MOV R0, #9\nADD R1, R10, #24\nSTR R0, [R1], #-8",
		mem: map[uint32]uint32{24: 9}, bufRegs: map[int]uint32{1: 16}},
	{name: "STR PC-relative label", code: "MOV R0, #42\nSTR R0, buf", mem: map[uint32]uint32{0: 42}},
	{name: "STRB", code: "MOV R0, #0xAB\nSTRB R0, [R10, #17]", mem: map[uint32]uint32{16: 0x0000AB00}},

	// Halfword transfers
	{name: "LDRH immediate", code: "LDRH R0, [R10, #2]", regs: map[int]uint32{0: 0x1122}},
	{name: "LDRH register", code: "MOV R2, #6\nLDRH R0, [R10, R2]", regs: map[int]uint32{0: 0x5566}},
	{name: "LDRH post-indexed", code: "MOV R1, R10\nLDRH R0, [R1], #2",
		regs: map[int]uint32{0: 0x3344}, bufRegs: map[int]uint32{1: 2}},
	{name: "STRH", code: "LDR R0, =0xBEEF\nSTRH R0, [R10, #18]", mem: map[uint32]uint32{16: 0xBEEF0000}},
	{name: "STRH pre-indexed negative", code: "LDR R0, =0xBEEF\nADD R1, R10, #24\nSTRH R0, [R1, #-2]!",
		mem: map[uint32]uint32{20: 0xBEEF0000}, bufRegs: map[int]uint32{1: 22}},

	// Load and store multiple
	{name: "LDMIA", code: "LDMIA R10, {R0-R3}",
		regs: map[int]uint32{0: 0x11223344, 1: 0x55667788, 2: 0x99AABBCC, 3: 0xDDEEFF00}, bufRegs: map[int]uint32{10: 0}},
	{name: "LDMIA write-back", code: "MOV R1, R10\nLDMIA R1!, {R2, R3}",
		regs: map[int]uint32{2: 0x11223344, 3: 0x55667788}, bufRegs: map[int]uint32{1: 8}},
	{name: "LDMIB write-back", code: "MOV R1, R10\nLDMIB R1!, {R2, R3}",
		regs: map[int]uint32{2: 0x55667788, 3: 0x99AABBCC}, bufRegs: map[int]uint32{1: 8}},
	{name: "LDMDA write-back", code: "ADD R1, R10, #12\nLDMDA R1!, {R2, R3}",
		regs: map[int]uint32{2: 0x99AABBCC, 3: 0xDDEEFF00}, bufRegs: map[int]uint32{1: 4}},
	{name: "LDMDB write-back", code: "ADD R1, R10, #12\nLDMDB R1!, {R2, R3}",
		regs: map[int]uint32{2: 0x55667788, 3: 0x99AABBCC}, bufRegs: map[int]uint32{1: 4}},
	{name: "STMIA write-back", code: "ADD R1, R10, #16\nMOV R2, #0xA\nMOV R3, #0xB\nSTMIA R1!, {R2, R3}",
		mem: map[uint32]uint32{16: 0xA, 20: 0xB}, bufRegs: map[int]uint32{1: 24}},
	{name: "STMIB", code: "ADD R1, R10, #16\nMOV R2, #0xA\nMOV R3, #0xB\nSTMIB R1, {R2, R3}",
		mem: map[uint32]uint32{20: 0xA, 24: 0xB}, bufRegs: map[int]uint32{1: 16}},
	{name: "STMDA write-back", code: "ADD R1, R10, #28\nMOV R2, #0xA\nMOV R3, #0xB\nSTMDA R1!, {R2, R3}",
		mem: map[uint32]uint32{24: 0xA, 28: 0xB}, bufRegs: map[int]uint32{1: 20}},
	{name: "STMDB write-back", code: "ADD R1, R10, #28\nMOV R2, #0xA\nMOV R3, #0xB\nSTMDB R1!, {R2, R3}",
		mem: map[uint32]uint32{20: 0xA, 24: 0xB}, bufRegs: map[int]uint32{1: 20}},
	{name: "PUSH and POP", code: "MOV R2, #1\nMOV R3, #2\nPUSH {R2, R3}\nPOP {R4, R5}",
		regs: map[int]uint32{4: 1, 5: 2, vm.SP: stackTop}},
	{name: "STMFD and LDMFD", code: "MOV R2, #1\nMOV R3, #2\nSTMFD SP!, {R2, R3}\nLDMFD SP!, {R4, R5}",
		regs: map[int]uint32{4: 1, 5: 2, vm.SP: stackTop}},
	{name: "STMFA and LDMFA", code: "ADD R1, R10, #12\nMOV R2, #1\nMOV R3, #2\nSTMFA R1!, {R2, R3}\nLDMFA R1!, {R4, R5}",
		regs: map[int]uint32{4: 1, 5: 2}, bufRegs: map[int]uint32{1: 12}, mem: map[uint32]uint32{16: 1, 20: 2}},
	{name: "STMEA and LDMEA", code: "ADD R1, R10, #16\nMOV R2, #1\nMOV R3, #2\nSTMEA R1!, {R2, R3}\nLDMEA R1!, {R4, R5}",
		regs: map[int]uint32{4: 1, 5: 2}, bufRegs: map[int]uint32{1: 16}, mem: map[uint32]uint32{16: 1, 20: 2}},
	{name: "STMED and LDMED", code: "ADD R1, R10, #28\nMOV R2, #1\nMOV R3, #2\nSTMED R1!, {R2, R3}\nLDMED R1!, {R4, R5}",
		regs: map[int]uint32{4: 1, 5: 2}, bufRegs: map[int]uint32{1: 28}, mem: map[uint32]uint32{24: 1, 28: 2}},

	// Branches
	{name: "B forward", code: "B skip\nMOV R0, #1\nskip:\nMOV R1, #2", regs: map[int]uint32{0: 0, 1: 2}},
	{name: "B backward loop", code: "MOV R0, #0\nMOV R1, #5\nloop:\nADD R0, R0, R1\nSUBS R1, R1, #1\nBNE loop",
		regs: map[int]uint32{0: 15, 1: 0}},
	{name: "BL and MOV PC, LR", code: "BL fn\nMOV R1, #2\nB done\nfn:\nMOV R0, #1\nMOV PC, LR\ndone:",
		regs: map[int]uint32{0: 1, 1: 2}},
	{name: "BL and BX LR", code: "BL fn\nMOV R1, #2\nB done\nfn:\nMOV R0, #1\nBX LR\ndone:",
		regs: map[int]uint32{0: 1, 1: 2}},
	{name: "BLX register", code: "ADR R3, fn\nBLX R3\nMOV R1, #2\nB done\nfn:\nMOV R0, #1\nBX LR\ndone:",
		regs: map[int]uint32{0: 1, 1: 2}},
	{name: "BL with stacked return", code: "BL fn\nMOV R1, #2\nB done\nfn:\nPUSH {R4, LR}\nMOV R0, #1\nPOP {R4, PC}\ndone:",
		regs: map[int]uint32{0: 1, 1: 2, vm.SP: stackTop}},

	// Multiplies
	{name: "MUL", code: "MOV R1, #6\nMOV R2, #7\nMUL R0, R1, R2", regs: map[int]uint32{0: 42}},
	{name: "MLA", code: "MOV R1, #6\nMOV R2, #7\nMOV R3, #100\nMLA R0, R1, R2, R3", regs: map[int]uint32{0: 142}},
	{name: "MULS zero", code: clearNZCV + "MOV R1, #0\nMOV R2, #7\nMULS R0, R1, R2", flags: &vm.CPSR{Z: true}},
	{name: "UMULL", code: "MVN R2, #0\nMOV R3, #2\nUMULL R0, R1, R2, R3", regs: map[int]uint32{0: 0xFFFFFFFE, 1: 1}},
	{name: "UMLAL", code: "MVN R2, #0\nMOV R3, #2\nMOV R0, #1\nMOV R1, #0\nUMLAL R0, R1, R2, R3",
		regs: map[int]uint32{0: 0xFFFFFFFF, 1: 1}},
	{name: "SMULL", code: "MVN R2, #0\nMOV R3, #2\nSMULL R0, R1, R2, R3", regs: map[int]uint32{0: 0xFFFFFFFE, 1: 0xFFFFFFFF}},
	{name: "SMLAL", code: "MVN R2, #0\nMOV R3, #2\nMOV R0, #4\nMOV R1, #0\nSMLAL R0, R1, R2, R3", regs: map[int]uint32{0: 2, 1: 0}},

	// Status register transfers
	{name: "MSR immediate and MRS", code: "MSR CPSR_f, #0xF0000000\nMRS R0, CPSR",
		regs: map[int]uint32{0: 0xF0000000}, flags: &vm.CPSR{N: true, Z: true, C: true, V: true}},
	{name: "MSR register", code: "MOV R1, #0x40000000\nMSR CPSR_f, R1", flags: &vm.CPSR{Z: true}},
	{name: "MSR control field keeps flags", code: "MSR CPSR_f, #0x80000000\nMOV R1, #0\nMSR CPSR_c, R1", flags: &vm.CPSR{N: true}},

	// Conditional execution outside data processing
	{name: "LDREQ taken", code: "MSR CPSR_f, #0x40000000\nLDREQ R0, [R10]", regs: map[int]uint32{0: 0x11223344}},
	{name: "STRNE skipped", code: "MSR CPSR_f, #0x40000000\nMOV R0, #5\nSTRNE R0, [R10, #16]", mem: map[uint32]uint32{16: 0}},
	{name: "BLGT skipped", code: clearNZCV + "MOV R0, #0\nMSR CPSR_f, #0x40000000\nBLGT fn\nB done\nfn:\nMOV R0, #1\nMOV PC, LR\ndone:",
		regs: map[int]uint32{0: 0}},
	{name: "MULCS taken", code: setC + "MOV R1, #3\nMOV R2, #4\nMULCS R0, R1, R2", regs: map[int]uint32{0: 12}},
	{name: "NOP", code: "MOV R0, #3\nNOP", regs: map[int]uint32{0: 3}},
}

// refCondition is the ARM condition table, written out independently of the VM.
func refCondition(cond string, n, z, c, v bool) bool {
	switch cond {
	case "EQ":
		return z
	case "NE":
		return !z
	case "CS":
		return c
	case "CC":
		return !c
	case "MI":
		return n
	case "PL":
		return !n
	case "VS":
		return v
	case "VC":
		return !v
	case "HI":
		return c && !z
	case "LS":
		return !c || z
	case "GE":
		return n == v
	case "LT":
		return n != v
	case "GT":
		return !z && n == v
	case "LE":
		return z || n != v
	case "AL":
		return true
	}
	panic(cond)
}

// TestConditionCodes runs MOV<cond> under every combination of NZCV. Each batch
// writes R0-R7, which the harness does not use.
func TestConditionCodes(t *testing.T) {
	const (
		flagShift = 28
		batchSize = 8
	)
	conds := []string{"EQ", "NE", "CS", "CC", "MI", "PL", "VS", "VC", "HI", "LS", "GE", "LT", "GT", "LE", "AL"}
	for flags := range uint32(16) {
		n, z, c, v := flags&8 != 0, flags&4 != 0, flags&2 != 0, flags&1 != 0
		for start := 0; start < len(conds); start += batchSize {
			batch := conds[start:min(start+batchSize, len(conds))]
			var code strings.Builder
			fmt.Fprintf(&code, "MSR CPSR_f, #0x%X\n", flags<<flagShift)
			for i, cond := range batch {
				fmt.Fprintf(&code, "MOV%s R%d, #1\n", cond, i)
			}
			t.Run(fmt.Sprintf("NZCV=%04b/%s-%s", flags, batch[0], batch[len(batch)-1]), func(t *testing.T) {
				m, _ := runForm(t, code.String())
				for i, cond := range batch {
					want := refCondition(cond, n, z, c, v)
					if got := m.CPU.R[i] == 1; got != want {
						t.Errorf("MOV%s executed=%v, want %v", cond, got, want)
					}
				}
			})
		}
	}
}
