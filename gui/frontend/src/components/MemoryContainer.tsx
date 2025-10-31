import React, { useState, useEffect, useCallback } from 'react'
import { MemoryView } from './MemoryView'
import { GetMemory, GetLastMemoryWrite } from '../../wailsjs/go/main/App'
import { EventsOn, EventsOff } from '../../wailsjs/runtime/runtime'

const MEMORY_WINDOW_SIZE = 256

export const MemoryContainer: React.FC = () => {
  const [memory, setMemory] = useState<Uint8Array>(new Uint8Array(MEMORY_WINDOW_SIZE))
  const [baseAddress, setBaseAddress] = useState<number>(0x8000)
  const [highlightAddresses, setHighlightAddresses] = useState<Set<number>>(new Set())

  const loadMemory = useCallback(async (address: number) => {
    try {
      const data = await GetMemory(address, MEMORY_WINDOW_SIZE)

      let uint8Array: Uint8Array
      if (typeof data === 'string') {
        // Wails is returning base64 encoded string, decode it
        const binaryString = atob(data)
        uint8Array = new Uint8Array(binaryString.length)
        for (let i = 0; i < binaryString.length; i++) {
          uint8Array[i] = binaryString.charCodeAt(i)
        }
      } else {
        // If it's already an array, use it directly
        uint8Array = new Uint8Array(data)
      }

      setMemory(uint8Array)
      setBaseAddress(address)
    } catch (error) {
      console.error('MemoryContainer: Failed to load memory:', error)
      // Don't update state on error - keep showing previous valid memory
    }
  }, [])

  const checkMemoryWrite = useCallback(async (): Promise<{ navigateTo: number | null, writeAddress: number | null }> => {
    try {
      const result = await GetLastMemoryWrite()
      if (result.hasWrite) {
        // Highlight the exact byte that was written
        setHighlightAddresses(new Set([result.address]))
        setTimeout(() => setHighlightAddresses(new Set()), 1000)

        // Align to 16-byte boundary for display
        const alignedAddress = Math.floor(result.address / 16) * 16

        // Only navigate if the write is outside the current visible window
        const currentStart = baseAddress
        const currentEnd = baseAddress + MEMORY_WINDOW_SIZE
        if (result.address < currentStart || result.address >= currentEnd) {
          return { navigateTo: alignedAddress, writeAddress: result.address }
        }
        return { navigateTo: null, writeAddress: result.address }
      }
      // No write, clear any previous highlight
      setHighlightAddresses(new Set())
      return { navigateTo: null, writeAddress: null }
    } catch (error) {
      console.error('Failed to check memory write:', error)
      return { navigateTo: null, writeAddress: null }
    }
  }, [baseAddress])

  useEffect(() => {
    loadMemory(baseAddress)
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [])

  useEffect(() => {
    const handleStateChange = async () => {
      const decision = await checkMemoryWrite()
      if (decision.navigateTo !== null) {
        await loadMemory(decision.navigateTo)
      } else {
        await loadMemory(baseAddress)
      }
    }

    const unsubscribe = EventsOn('vm:state-changed', handleStateChange)

    return () => {
      unsubscribe()
    }
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [])

  return (
    <MemoryView
      memory={memory}
      baseAddress={baseAddress}
      onAddressChange={loadMemory}
      highlightAddresses={highlightAddresses}
    />
  )
}
