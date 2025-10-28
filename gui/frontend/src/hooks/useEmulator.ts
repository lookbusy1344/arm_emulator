import { useState, useCallback, useEffect } from 'react'
import { EmulatorAPI } from '../services/wails'
import type { RegisterState, ExecutionState, BreakpointInfo } from '../types/emulator'

export const useEmulator = () => {
  const [registers, setRegisters] = useState<RegisterState | null>(null)
  const [executionState, setExecutionState] = useState<ExecutionState>('halted')
  const [breakpoints, setBreakpoints] = useState<BreakpointInfo[]>([])
  const [memory, setMemory] = useState<Uint8Array>(new Uint8Array(256))
  const [memoryAddress, setMemoryAddress] = useState(0x8000)
  const [error, setError] = useState<string | null>(null)

  // Load program from source
  const loadProgram = useCallback(async (source: string, filename: string, entryPoint: number) => {
    try {
      await EmulatorAPI.loadProgram(source, filename, entryPoint)
      await refreshState()
      setError(null)
    } catch (err) {
      setError(err instanceof Error ? err.message : String(err))
    }
  }, [])

  // Step single instruction
  const step = useCallback(async () => {
    try {
      await EmulatorAPI.step()
      await refreshState()
      setError(null)
    } catch (err) {
      setError(err instanceof Error ? err.message : String(err))
    }
  }, [])

  // Continue execution
  const continueExecution = useCallback(async () => {
    try {
      setExecutionState('running')
      await EmulatorAPI.continue()
      await refreshState()
      setError(null)
    } catch (err) {
      setError(err instanceof Error ? err.message : String(err))
    }
  }, [])

  // Pause execution
  const pause = useCallback(async () => {
    try {
      await EmulatorAPI.pause()
      await refreshState()
      setError(null)
    } catch (err) {
      setError(err instanceof Error ? err.message : String(err))
    }
  }, [])

  // Reset VM
  const reset = useCallback(async () => {
    try {
      await EmulatorAPI.reset()
      await refreshState()
      setError(null)
    } catch (err) {
      setError(err instanceof Error ? err.message : String(err))
    }
  }, [])

  // Add breakpoint
  const addBreakpoint = useCallback(async (address: number) => {
    try {
      await EmulatorAPI.addBreakpoint(address)
      const bps = await EmulatorAPI.getBreakpoints()
      setBreakpoints(bps)
      setError(null)
    } catch (err) {
      setError(err instanceof Error ? err.message : String(err))
    }
  }, [])

  // Remove breakpoint
  const removeBreakpoint = useCallback(async (address: number) => {
    try {
      await EmulatorAPI.removeBreakpoint(address)
      const bps = await EmulatorAPI.getBreakpoints()
      setBreakpoints(bps)
      setError(null)
    } catch (err) {
      setError(err instanceof Error ? err.message : String(err))
    }
  }, [])

  // Change memory view address
  const changeMemoryAddress = useCallback(async (address: number) => {
    setMemoryAddress(address)
    try {
      const mem = await EmulatorAPI.getMemory(address, 256)
      setMemory(mem)
      setError(null)
    } catch (err) {
      setError(err instanceof Error ? err.message : String(err))
    }
  }, [])

  // Refresh all state from backend
  const refreshState = useCallback(async () => {
    try {
      const [regs, state, bps, mem] = await Promise.all([
        EmulatorAPI.getRegisters(),
        EmulatorAPI.getExecutionState(),
        EmulatorAPI.getBreakpoints(),
        EmulatorAPI.getMemory(memoryAddress, 256),
      ])

      setRegisters(regs)
      setExecutionState(state)
      setBreakpoints(bps)
      setMemory(mem)
    } catch (err) {
      setError(err instanceof Error ? err.message : String(err))
    }
  }, [memoryAddress])

  // Initial load
  useEffect(() => {
    refreshState()
  }, [])

  return {
    registers,
    executionState,
    breakpoints,
    memory,
    memoryAddress,
    error,
    loadProgram,
    step,
    continue: continueExecution,
    pause,
    reset,
    addBreakpoint,
    removeBreakpoint,
    changeMemoryAddress,
    refreshState,
  }
}
