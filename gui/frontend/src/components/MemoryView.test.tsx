import { describe, it, expect } from 'vitest'
import { render, screen } from '@testing-library/react'
import userEvent from '@testing-library/user-event'
import { MemoryView } from './MemoryView'

describe('MemoryView', () => {
  const mockMemory = new Uint8Array([
    0x48, 0x65, 0x6c, 0x6c, 0x6f, 0x20, 0x57, 0x6f,
    0x72, 0x6c, 0x64, 0x21, 0x00, 0x00, 0x00, 0x00,
  ])

  it('should render memory dump in hex', () => {
    render(
      <MemoryView
        memory={mockMemory}
        baseAddress={0x8000}
        onAddressChange={() => {}}
      />
    )

    // Should show hex values
    expect(screen.getByText(/48/)).toBeInTheDocument()
    expect(screen.getByText(/65/)).toBeInTheDocument()
  })

  it('should display ASCII representation', () => {
    render(
      <MemoryView
        memory={mockMemory}
        baseAddress={0x8000}
        onAddressChange={() => {}}
      />
    )

    // Should show "Hello World!"
    expect(screen.getByText(/Hello World/)).toBeInTheDocument()
  })

  it('should show base address', () => {
    render(
      <MemoryView
        memory={mockMemory}
        baseAddress={0x8000}
        onAddressChange={() => {}}
      />
    )

    expect(screen.getByText(/0x00008000/i)).toBeInTheDocument()
  })

  it('should call onAddressChange when address input changes', async () => {
    const user = userEvent.setup()
    let newAddress = 0

    render(
      <MemoryView
        memory={mockMemory}
        baseAddress={0x8000}
        onAddressChange={(addr) => { newAddress = addr }}
      />
    )

    const input = screen.getByPlaceholderText(/address/i)
    await user.clear(input)
    await user.type(input, '0x9000')
    await user.keyboard('{Enter}')

    expect(newAddress).toBe(0x9000)
  })

  it('should format addresses in rows', () => {
    // Create 32 bytes to test multiple rows
    const largerMemory = new Uint8Array(32)
    render(
      <MemoryView
        memory={largerMemory}
        baseAddress={0x8000}
        onAddressChange={() => {}}
      />
    )

    // First row should start with base address
    expect(screen.getByText('0x00008000')).toBeInTheDocument()
    // Second row should be +0x10 (16 bytes per row)
    expect(screen.getByText('0x00008010')).toBeInTheDocument()
  })
})
