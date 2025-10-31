import React, { useState, useEffect, useCallback } from 'react'
import { MemoryView } from './MemoryView'
import { GetMemory, GetLastMemoryWrite } from '../../wailsjs/go/main/App'
import { EventsOn, EventsOff } from '../../wailsjs/runtime/runtime'

const MEMORY_WINDOW_SIZE = 256

export const MemoryContainer: React.FC = () => {
  const [memory, setMemory] = useState<Uint8Array>(new Uint8Array(MEMORY_WINDOW_SIZE))
  const [baseAddress, setBaseAddress] = useState<number>(0x8000)

  const loadMemory = useCallback(async (address: number) => {
    const msg = `MemoryContainer: loadMemory called with address 0x${address.toString(16)}`
    console.log(msg)
    try {
      const data = await GetMemory(address, MEMORY_WINDOW_SIZE)
      console.log(`MemoryContainer: received data, type=${typeof data}, length=${data.length}, isArray=${Array.isArray(data)}`)
      
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
      
      console.log(`MemoryContainer: created Uint8Array, length=${uint8Array.length}`)
      setMemory(uint8Array)
      setBaseAddress(address)
    } catch (error) {
      console.error('MemoryContainer: Failed to load memory:', error)
      // Don't update state on error - keep showing previous valid memory
    }
  }, [])

  const checkMemoryWrite = useCallback(async (): Promise<number | null> => {
    console.log('MemoryContainer: checkMemoryWrite called')
    try {
      const result = await GetLastMemoryWrite()
      console.log(`MemoryContainer: GetLastMemoryWrite returned address=0x${result.address.toString(16)}, hasWrite=${result.hasWrite}`)
      if (result.hasWrite) {
        // Align to 16-byte boundary for display
        const alignedAddress = Math.floor(result.address / 16) * 16
        console.log(`MemoryContainer: aligned address to 0x${alignedAddress.toString(16)}`)
        
        // Only navigate if the write is outside the current visible window
        const currentStart = baseAddress
        const currentEnd = baseAddress + MEMORY_WINDOW_SIZE
        if (result.address < currentStart || result.address >= currentEnd) {
          console.log(`MemoryContainer: write at 0x${result.address.toString(16)} is outside window [0x${currentStart.toString(16)}-0x${currentEnd.toString(16)}], will navigate`)
          return alignedAddress
        }
        console.log(`MemoryContainer: write at 0x${result.address.toString(16)} is within current window, reloading`)
      }
      return null
    } catch (error) {
      console.error('Failed to check memory write:', error)
      return null
    }
  }, [baseAddress])

  useEffect(() => {
    console.log('MemoryContainer: initial useEffect, loading memory at 0x' + baseAddress.toString(16))
    loadMemory(baseAddress)
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [])

  useEffect(() => {
    const handleStateChange = async () => {
      console.log('MemoryContainer: vm:state-changed event received')
      // Just reload current view - don't auto-navigate on memory writes
      // This avoids issues with writes near segment boundaries
      console.log(`MemoryContainer: reloading current view at 0x${baseAddress.toString(16)}`)
      await loadMemory(baseAddress)
      console.log('MemoryContainer: handleStateChange completed')
    }

    console.log('MemoryContainer: setting up vm:state-changed listener')
    EventsOn('vm:state-changed', handleStateChange)

    return () => {
      console.log('MemoryContainer: cleaning up vm:state-changed listener')
      EventsOff('vm:state-changed')
    }
  }, [baseAddress, loadMemory])

  return (
    <MemoryView
      memory={memory}
      baseAddress={baseAddress}
      onAddressChange={loadMemory}
    />
  )
}
