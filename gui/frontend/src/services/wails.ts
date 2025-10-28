import type {
  RegisterState,
  BreakpointInfo,
  ExecutionState,
  SymbolTable,
} from '../types/emulator'

// Wails runtime bindings
// These are injected by Wails at runtime via window.go
declare global {
  interface Window {
    go: {
      main: {
        App: {
          LoadProgramFromSource: (source: string, filename: string, entryPoint: number) => Promise<string | null>
          GetRegisters: () => Promise<RegisterState>
          Step: () => Promise<string | null>
          Continue: () => Promise<string | null>
          Pause: () => Promise<void>
          Reset: () => Promise<string | null>
          AddBreakpoint: (address: number) => Promise<string | null>
          RemoveBreakpoint: (address: number) => Promise<string | null>
          GetBreakpoints: () => Promise<BreakpointInfo[]>
          GetMemory: (address: number, size: number) => Promise<number[]>
          GetSourceLine: (address: number) => Promise<string>
          GetSymbols: () => Promise<SymbolTable>
          GetExecutionState: () => Promise<string>
          IsRunning: () => Promise<boolean>
        }
      }
    }
  }
}

/**
 * EmulatorAPI provides typed wrapper around Wails backend methods
 * All methods return promises that resolve on success or reject with error message
 */
export const EmulatorAPI = {
  /**
   * Load and parse assembly source code
   */
  async loadProgram(source: string, filename: string, entryPoint: number): Promise<void> {
    const err = await window.go.main.App.LoadProgramFromSource(source, filename, entryPoint)
    if (err) {
      throw new Error(err)
    }
  },

  /**
   * Get current register state
   */
  async getRegisters(): Promise<RegisterState> {
    return window.go.main.App.GetRegisters()
  },

  /**
   * Execute single instruction
   */
  async step(): Promise<void> {
    const err = await window.go.main.App.Step()
    if (err) {
      throw new Error(err)
    }
  },

  /**
   * Continue execution until breakpoint or halt
   */
  async continue(): Promise<void> {
    const err = await window.go.main.App.Continue()
    if (err) {
      throw new Error(err)
    }
  },

  /**
   * Pause execution
   */
  async pause(): Promise<void> {
    await window.go.main.App.Pause()
  },

  /**
   * Reset VM to initial state
   */
  async reset(): Promise<void> {
    const err = await window.go.main.App.Reset()
    if (err) {
      throw new Error(err)
    }
  },

  /**
   * Add breakpoint at address
   */
  async addBreakpoint(address: number): Promise<void> {
    const err = await window.go.main.App.AddBreakpoint(address)
    if (err) {
      throw new Error(err)
    }
  },

  /**
   * Remove breakpoint
   */
  async removeBreakpoint(address: number): Promise<void> {
    const err = await window.go.main.App.RemoveBreakpoint(address)
    if (err) {
      throw new Error(err)
    }
  },

  /**
   * Get all breakpoints
   */
  async getBreakpoints(): Promise<BreakpointInfo[]> {
    return window.go.main.App.GetBreakpoints()
  },

  /**
   * Read memory region
   */
  async getMemory(address: number, size: number): Promise<Uint8Array> {
    const data = await window.go.main.App.GetMemory(address, size)
    return new Uint8Array(data)
  },

  /**
   * Get source line for address
   */
  async getSourceLine(address: number): Promise<string> {
    return window.go.main.App.GetSourceLine(address)
  },

  /**
   * Get all symbols
   */
  async getSymbols(): Promise<SymbolTable> {
    return window.go.main.App.GetSymbols()
  },

  /**
   * Get current execution state
   */
  async getExecutionState(): Promise<ExecutionState> {
    const state = await window.go.main.App.GetExecutionState()
    return state as ExecutionState
  },

  /**
   * Check if execution is active
   */
  async isRunning(): Promise<boolean> {
    return window.go.main.App.IsRunning()
  },
}
