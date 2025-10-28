export interface RegisterState {
  Registers: number[]  // [16]uint32 from Go
  CPSR: CPSRState
  PC: number           // uint32
  Cycles: number       // uint64 (may lose precision for very large values)
}

export interface CPSRState {
  N: boolean  // Negative
  Z: boolean  // Zero
  C: boolean  // Carry
  V: boolean  // Overflow
}

export interface BreakpointInfo {
  Address: number   // uint32
  Enabled: boolean
}

export interface WatchpointInfo {
  Address: number   // uint32
  Type: string      // "read" | "write" | "readwrite"
  Enabled: boolean
}

export type ExecutionState = 'running' | 'halted' | 'breakpoint' | 'error'

export interface SymbolTable {
  [name: string]: number  // uint32
}
