import { describe, it, expect, vi, beforeEach } from 'vitest'
import { render, screen, waitFor } from '@testing-library/react'
import userEvent from '@testing-library/user-event'
import App from '../App'

// Mock Wails API
const mockAPI = {
  LoadProgramFromSource: vi.fn(),
  GetRegisters: vi.fn(),
  Step: vi.fn(),
  Continue: vi.fn(),
  Pause: vi.fn(),
  Reset: vi.fn(),
  AddBreakpoint: vi.fn(),
  RemoveBreakpoint: vi.fn(),
  GetBreakpoints: vi.fn(),
  GetMemory: vi.fn(),
  GetSourceLine: vi.fn(),
  GetSymbols: vi.fn(),
  GetExecutionState: vi.fn(),
  IsRunning: vi.fn(),
}

beforeEach(() => {
  window.go = {
    main: {
      App: mockAPI as any,
    },
  }

  // Default mock responses
  mockAPI.GetRegisters.mockResolvedValue({
    Registers: new Array(16).fill(0),
    CPSR: { N: false, Z: false, C: false, V: false },
    PC: 0x8000,
    Cycles: 0,
  })
  mockAPI.GetExecutionState.mockResolvedValue('halted')
  mockAPI.GetBreakpoints.mockResolvedValue([])
  mockAPI.GetMemory.mockResolvedValue(new Array(256).fill(0))
  mockAPI.LoadProgramFromSource.mockResolvedValue(null)
  mockAPI.Step.mockResolvedValue(null)

  vi.clearAllMocks()
})

describe('App Integration', () => {
  it('should render main interface', async () => {
    render(<App />)

    expect(screen.getByText('ARM Emulator')).toBeInTheDocument()
    expect(screen.getByText('Load')).toBeInTheDocument()
    expect(screen.getByText('Step')).toBeInTheDocument()
    expect(screen.getByText('Run')).toBeInTheDocument()
  })

  it('should load program when Load clicked', async () => {
    const user = userEvent.setup()
    render(<App />)

    const loadButton = screen.getByText('Load')
    await user.click(loadButton)

    await waitFor(() => {
      expect(mockAPI.LoadProgramFromSource).toHaveBeenCalled()
    })
  })

  it('should step execution when Step clicked', async () => {
    const user = userEvent.setup()
    render(<App />)

    const stepButton = screen.getByText('Step')
    await user.click(stepButton)

    await waitFor(() => {
      expect(mockAPI.Step).toHaveBeenCalled()
    })
  })

  it('should display register values', async () => {
    mockAPI.GetRegisters.mockResolvedValue({
      Registers: [42, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0],
      CPSR: { N: false, Z: true, C: false, V: false },
      PC: 0x8004,
      Cycles: 123,
    })

    render(<App />)

    await waitFor(() => {
      expect(screen.getByText(/0x0000002A/)).toBeInTheDocument() // R0 = 42
      expect(screen.getByText(/0x00008004/)).toBeInTheDocument() // PC
      expect(screen.getByText(/123/)).toBeInTheDocument() // Cycles
    })
  })

  it('should show error when operation fails', async () => {
    const user = userEvent.setup()
    mockAPI.LoadProgramFromSource.mockResolvedValue('Parse error at line 1')

    render(<App />)

    const loadButton = screen.getByText('Load')
    await user.click(loadButton)

    await waitFor(() => {
      expect(screen.getByText(/Parse error at line 1/)).toBeInTheDocument()
    })
  })
})
