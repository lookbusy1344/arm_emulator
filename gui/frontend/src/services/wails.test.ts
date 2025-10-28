import { describe, it, expect, vi, beforeEach } from 'vitest'
import { EmulatorAPI } from './wails'

// Mock window.go
const mockGo = {
  main: {
    App: {
      LoadProgramFromSource: vi.fn(),
      GetRegisters: vi.fn(),
      Step: vi.fn(),
      Continue: vi.fn(),
    },
  },
}

declare global {
  interface Window {
    go: typeof mockGo
  }
}

beforeEach(() => {
  window.go = mockGo as any
  vi.clearAllMocks()
})

describe('EmulatorAPI', () => {
  it('should load program from source', async () => {
    mockGo.main.App.LoadProgramFromSource.mockResolvedValue(null)

    await EmulatorAPI.loadProgram('MOV R0, #42', 'test.s', 0x8000)

    expect(mockGo.main.App.LoadProgramFromSource).toHaveBeenCalledWith(
      'MOV R0, #42',
      'test.s',
      0x8000
    )
  })

  it('should get registers', async () => {
    const mockRegs = {
      Registers: new Array(16).fill(0),
      CPSR: { N: false, Z: false, C: false, V: false },
      PC: 0x8000,
      Cycles: 0,
    }
    mockGo.main.App.GetRegisters.mockResolvedValue(mockRegs)

    const regs = await EmulatorAPI.getRegisters()

    expect(regs).toEqual(mockRegs)
    expect(mockGo.main.App.GetRegisters).toHaveBeenCalled()
  })

  it('should step execution', async () => {
    mockGo.main.App.Step.mockResolvedValue(null)

    await EmulatorAPI.step()

    expect(mockGo.main.App.Step).toHaveBeenCalled()
  })
})
