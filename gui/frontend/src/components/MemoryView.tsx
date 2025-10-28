import React, { useState, useCallback } from 'react'
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

  // Split memory into rows
  const rows: Uint8Array[] = []
  for (let i = 0; i < memory.length; i += BYTES_PER_ROW) {
    rows.push(memory.slice(i, i + BYTES_PER_ROW))
  }

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

      <div className="memory-dump">
        {rows.map((row, rowIndex) => {
          const rowAddress = baseAddress + rowIndex * BYTES_PER_ROW

          return (
            <div key={rowIndex} className="memory-row">
              <span className="memory-address">{formatAddress(rowAddress)}</span>

              <div className="memory-hex">
                {Array.from(row).map((byte, byteIndex) => {
                  const byteAddr = rowAddress + byteIndex
                  const isHighlighted = highlightAddresses.has(byteAddr)

                  return (
                    <span
                      key={byteIndex}
                      className={`memory-byte ${isHighlighted ? 'memory-byte-highlight' : ''}`}
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
