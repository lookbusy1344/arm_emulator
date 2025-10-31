import React, { useState, useEffect, useCallback } from 'react'
import { MemoryView } from './MemoryView'
import { GetMemory, GetLastMemoryWrite } from '../../wailsjs/go/main/App'
import { EventsOn, EventsOff } from '../../wailsjs/runtime/runtime'

const MEMORY_WINDOW_SIZE = 256

export const MemoryContainer: React.FC = () => {
  const [memory, setMemory] = useState<Uint8Array>(new Uint8Array(MEMORY_WINDOW_SIZE))
  const [baseAddress, setBaseAddress] = useState<number>(0x8000)

  const loadMemory = useCallback(async (address: number) => {
    try {
      const data = await GetMemory(address, MEMORY_WINDOW_SIZE)
      setMemory(new Uint8Array(data))
      setBaseAddress(address)
    } catch (error) {
      console.error('Failed to load memory:', error)
    }
  }, [])

  const checkMemoryWrite = useCallback(async (): Promise<number | null> => {
    try {
      const result = await GetLastMemoryWrite()
      if (result.hasWrite) {
        // Align to 16-byte boundary for display
        const alignedAddress = Math.floor(result.address / 16) * 16
        return alignedAddress
      }
      return null
    } catch (error) {
      console.error('Failed to check memory write:', error)
      return null
    }
  }, [])

  useEffect(() => {
    loadMemory(baseAddress)
  }, [])

  useEffect(() => {
    const handleStateChange = async () => {
      const writeAddress = await checkMemoryWrite()
      if (writeAddress !== null) {
        // Navigate to the write address
        await loadMemory(writeAddress)
      } else {
        // Just reload current view
        await loadMemory(baseAddress)
      }
    }

    EventsOn('vm:state-changed', handleStateChange)

    return () => {
      EventsOff('vm:state-changed')
    }
  }, [baseAddress, checkMemoryWrite, loadMemory])

  return (
    <MemoryView
      memory={memory}
      baseAddress={baseAddress}
      onAddressChange={loadMemory}
    />
  )
}
