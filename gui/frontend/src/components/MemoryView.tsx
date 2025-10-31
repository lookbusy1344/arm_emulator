import React, { useState, useCallback, useRef, useEffect } from 'react'
import './MemoryView.css'

interface MemoryViewProps {
  memory: Uint8Array
  baseAddress: number
  onAddressChange: (address: number) => void
  highlightAddresses?: Set<number>
}

const BYTES_PER_ROW = 16

const formatHex8 = (value: number): string => {
  return value.toString(16).toUpperCase().padStart(2, '0')
}

const formatAddress = (address: number): string => {
  return `0x${address.toString(16).toUpperCase().padStart(8, '0')}`
}

const toASCII = (byte: number): string => {
  if (byte >= 32 && byte <= 126) {
    return String.fromCharCode(byte)
  }
  return '.'
}

export const MemoryView: React.FC<MemoryViewProps> = ({
  memory,
  baseAddress,
  onAddressChange,
  highlightAddresses = new Set(),
}) => {
  const [addressInput, setAddressInput] = useState(formatAddress(baseAddress))
  const previousMemoryRef = useRef<Uint8Array>(new Uint8Array(memory))
  const [changedBytes, setChangedBytes] = useState<Set<number>>(new Set())
  const memoryDumpRef = useRef<HTMLDivElement>(null)
  const changedByteRefs = useRef<Map<number, HTMLSpanElement>>(new Map())
  const prevBaseAddressRef = useRef<number>(baseAddress)

  const handleAddressSubmit = useCallback((e: React.FormEvent) => {
    e.preventDefault()

    let addr: number
    const input = addressInput.trim()

    if (input.startsWith('0x') || input.startsWith('0X')) {
      addr = parseInt(input.substring(2), 16)
    } else {
      addr = parseInt(input, 10)
    }

    if (!isNaN(addr)) {
      onAddressChange(addr)
    }
  }, [addressInput, onAddressChange])

  useEffect(() => {
    const baseChanged = prevBaseAddressRef.current !== baseAddress
    const changed = new Set<number>()

    if (!baseChanged) {
      for (let i = 0; i < memory.length; i++) {
        if (memory[i] !== previousMemoryRef.current[i]) {
          changed.add(baseAddress + i)
        }
      }
      previousMemoryRef.current = new Uint8Array(memory)
      setChangedBytes(changed)
    } else {
      // When the window jumps to a new base, don't flood with false-positive changes
      previousMemoryRef.current = new Uint8Array(memory)
      setChangedBytes(new Set())
    }

    // Prefer scrolling to an explicitly highlighted address (from last write),
    // otherwise fall back to the first changed byte within the window
    const targets: number[] = []
    if (highlightAddresses.size > 0) targets.push(...Array.from(highlightAddresses))
    if (!baseChanged && changed.size > 0) targets.push(...Array.from(changed))

    if (targets.length > 0) {
      const target = Math.min(...targets)
      setTimeout(() => {
        const element = changedByteRefs.current.get(target)
        if (element && memoryDumpRef.current) {
          element.scrollIntoView({ behavior: 'smooth', block: 'center' })
        }
      }, 0)
    }

    // Clear yellow change highlight after a short delay
    if (!baseChanged && changed.size > 0) {
      const timer = setTimeout(() => setChangedBytes(new Set()), 500)
      return () => clearTimeout(timer)
    }

    prevBaseAddressRef.current = baseAddress
  }, [memory, baseAddress, highlightAddresses])

  // Split memory into rows
  const rows: Uint8Array[] = []
  for (let i = 0; i < memory.length; i += BYTES_PER_ROW) {
    rows.push(memory.slice(i, i + BYTES_PER_ROW))
  }
  
  console.log(`MemoryView: memory.length=${memory.length}, rows.length=${rows.length}, baseAddress=0x${baseAddress.toString(16)}`)

  return (
    <div className="memory-view">
      <div className="memory-view-header">
        <div className="memory-view-title">Memory</div>
        <form onSubmit={handleAddressSubmit} className="address-input-form">
          <input
            type="text"
            className="address-input"
            value={addressInput}
            onChange={(e) => setAddressInput(e.target.value)}
            placeholder="Address (0xXXXX)"
          />
          <button type="submit" className="address-go-button">Go</button>
        </form>
      </div>

      <div className="memory-dump" ref={memoryDumpRef}>
        {rows.map((row, rowIndex) => {
          const rowAddress = baseAddress + rowIndex * BYTES_PER_ROW

          return (
            <div key={rowIndex} className="memory-row">
              <span className="memory-address">{formatAddress(rowAddress)}</span>

              <div className="memory-hex">
                {Array.from(row).map((byte, byteIndex) => {
                  const byteAddr = rowAddress + byteIndex
                  const isHighlighted = highlightAddresses.has(byteAddr)
                  const isChanged = changedBytes.has(byteAddr)

                  return (
                    <span
                      key={byteIndex}
                      ref={(el) => {
                        if (el && (isChanged || isHighlighted)) {
                          changedByteRefs.current.set(byteAddr, el)
                        }
                      }}
                      className={`memory-byte ${isHighlighted ? 'memory-byte-highlight' : ''} ${isChanged ? 'memory-byte-changed' : ''}`}
                    >
                      {formatHex8(byte)}
                    </span>
                  )
                })}
              </div>

              <span className="memory-ascii">
                {Array.from(row).map((byte) => toASCII(byte)).join('')}
              </span>
            </div>
          )
        })}
      </div>
    </div>
  )
}
