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
    const changed = new Set<number>()
    for (let i = 0; i < memory.length; i++) {
      if (memory[i] !== previousMemoryRef.current[i]) {
        changed.add(baseAddress + i)
      }
    }
    
    previousMemoryRef.current = new Uint8Array(memory)
    setChangedBytes(changed)
    
    if (changed.size > 0) {
      // Scroll first changed byte into view
      const firstChanged = Math.min(...Array.from(changed))
      setTimeout(() => {
        const element = changedByteRefs.current.get(firstChanged)
        if (element && memoryDumpRef.current) {
          element.scrollIntoView({ behavior: 'smooth', block: 'center' })
        }
      }, 0)
      
      const timer = setTimeout(() => setChangedBytes(new Set()), 500)
      return () => clearTimeout(timer)
    }
  }, [memory, baseAddress])

  // Split memory into rows
  const rows: Uint8Array[] = []
  for (let i = 0; i < memory.length; i += BYTES_PER_ROW) {
    rows.push(memory.slice(i, i + BYTES_PER_ROW))
  }
  
  console.log(`MemoryView: memory.length=${memory.length}, rows.length=${rows.length}, baseAddress=0x${baseAddress.toString(16)}`)

  return (
    <div className="memory-view">
      <div className="memory-view-header">
        <h3 className="memory-view-title">Memory</h3>
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
                        if (el && isChanged) {
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
