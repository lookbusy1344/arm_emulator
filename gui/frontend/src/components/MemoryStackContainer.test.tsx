import { describe, it, expect, vi } from 'vitest'
import { render, screen, fireEvent } from '@testing-library/react'
import { MemoryStackContainer } from './MemoryStackContainer'

// Mock the child components
vi.mock('./MemoryContainer', () => ({
  MemoryContainer: () => <div data-testid="memory-container">Memory Container</div>
}))

vi.mock('./StackView', () => ({
  StackView: () => <div data-testid="stack-view">Stack View</div>
}))

describe('MemoryStackContainer', () => {
  it('should render Memory and Stack tabs', () => {
    render(<MemoryStackContainer />)

    expect(screen.getByTestId('memory-tab')).toBeInTheDocument()
    expect(screen.getByTestId('stack-tab')).toBeInTheDocument()
  })

  it('should default to Memory tab active', () => {
    render(<MemoryStackContainer />)

    const memoryTab = screen.getByTestId('memory-tab')
    const stackTab = screen.getByTestId('stack-tab')

    expect(memoryTab).toHaveClass('active')
    expect(stackTab).not.toHaveClass('active')
  })

  it('should show Memory content by default', () => {
    render(<MemoryStackContainer />)

    const memoryContainer = screen.getByTestId('memory-container')
    const stackView = screen.getByTestId('stack-view')

    expect(memoryContainer.parentElement).toHaveStyle({ display: 'flex' })
    expect(stackView.parentElement).toHaveStyle({ display: 'none' })
  })

  it('should switch to Stack tab when clicked', () => {
    render(<MemoryStackContainer />)

    const stackTab = screen.getByTestId('stack-tab')
    fireEvent.click(stackTab)

    const memoryContainer = screen.getByTestId('memory-container')
    const stackView = screen.getByTestId('stack-view')

    expect(memoryContainer.parentElement).toHaveStyle({ display: 'none' })
    expect(stackView.parentElement).toHaveStyle({ display: 'flex' })
  })

  it('should update active tab styling on click', () => {
    render(<MemoryStackContainer />)

    const memoryTab = screen.getByTestId('memory-tab')
    const stackTab = screen.getByTestId('stack-tab')

    // Initially Memory is active
    expect(memoryTab).toHaveClass('active')
    expect(stackTab).not.toHaveClass('active')

    // Click Stack tab
    fireEvent.click(stackTab)
    expect(memoryTab).not.toHaveClass('active')
    expect(stackTab).toHaveClass('active')

    // Click Memory tab again
    fireEvent.click(memoryTab)
    expect(memoryTab).toHaveClass('active')
    expect(stackTab).not.toHaveClass('active')
  })
})
