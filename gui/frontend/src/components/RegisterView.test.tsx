import { describe, it, expect } from 'vitest'
import { render, screen } from '@testing-library/react'
import { RegisterView } from './RegisterView'
import type { RegisterState } from '../types/emulator'

describe('RegisterView', () => {
  const mockRegisters: RegisterState = {
    Registers: [42, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0],
    CPSR: { N: true, Z: false, C: false, V: false },
    PC: 0x8000,
    Cycles: 123,
  }

  it('should render all registers', () => {
    render(<RegisterView registers={mockRegisters} />)

    // Check R0-R15 labels exist
    for (let i = 0; i < 16; i++) {
      expect(screen.getByText(`R${i}`)).toBeInTheDocument()
    }
  })

  it('should display register values in hex', () => {
    render(<RegisterView registers={mockRegisters} />)

    // R0 should show 0x0000002A (42 in hex)
    expect(screen.getByText(/0x0000002A/i)).toBeInTheDocument()
  })

  it('should display CPSR flags', () => {
    render(<RegisterView registers={mockRegisters} />)

    expect(screen.getByText('N: 1')).toBeInTheDocument()
    expect(screen.getByText('Z: 0')).toBeInTheDocument()
    expect(screen.getByText('C: 0')).toBeInTheDocument()
    expect(screen.getByText('V: 0')).toBeInTheDocument()
  })

  it('should highlight changed registers', () => {
    const changedRegs = new Set([0, 15])
    const { container } = render(
      <RegisterView registers={mockRegisters} changedRegisters={changedRegs} />
    )

    // R0 and PC (R15) should have highlight class
    const highlighted = container.querySelectorAll('.register-changed')
    expect(highlighted).toHaveLength(2)
  })

  it('should display PC value', () => {
    render(<RegisterView registers={mockRegisters} />)

    expect(screen.getByText(/0x00008000/i)).toBeInTheDocument()
  })

  it('should display cycle count', () => {
    render(<RegisterView registers={mockRegisters} />)

    expect(screen.getByText(/123/)).toBeInTheDocument()
  })
})
