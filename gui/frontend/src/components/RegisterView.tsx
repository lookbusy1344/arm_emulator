import React, { useEffect, useState } from 'react'
import type { RegisterState } from '../types/emulator'
import { GetRegisters } from '../../wailsjs/go/main/App'
import { EventsOn, EventsOff } from '../../wailsjs/runtime/runtime'
import './RegisterView.css'

interface RegisterViewProps {
  registers?: RegisterState
  changedRegisters?: Set<number>
}

const formatHex32 = (value: number): string => {
  return `0x${value.toString(16).toUpperCase().padStart(8, '0')}`
}

export const RegisterView: React.FC<RegisterViewProps> = ({
  registers: externalRegisters,
  changedRegisters = new Set(),
}) => {
  const [registers, setRegisters] = useState<RegisterState | null>(externalRegisters || null)
  const [previousRegisters, setPreviousRegisters] = useState<number[]>(Array(16).fill(0))
  const [highlightedRegs, setHighlightedRegs] = useState<Set<number>>(new Set())

  const loadRegisters = async () => {
    try {
      const regs = await GetRegisters()
      
      // Track which registers changed
      const changed = new Set<number>()
      if (registers) {
        regs.Registers.forEach((val, idx) => {
          if (val !== previousRegisters[idx]) {
            changed.add(idx)
          }
        })
      }
      
      setPreviousRegisters([...regs.Registers])
      setHighlightedRegs(changed)
      setRegisters(regs)
      
      // Clear highlights after 1 second
      if (changed.size > 0) {
        setTimeout(() => setHighlightedRegs(new Set()), 1000)
      }
    } catch (error) {
      console.error('Failed to load registers:', error)
    }
  }

  useEffect(() => {
    loadRegisters()
    EventsOn('vm:state-changed', loadRegisters)
    return () => {
      EventsOff('vm:state-changed')
    }
  }, [])

  if (!registers) {
    return <div className="register-view">Loading...</div>
  }

  const { Registers, CPSR, PC, Cycles } = registers
  const effectiveHighlights = changedRegisters.size > 0 ? changedRegisters : highlightedRegs

  return (
    <div className="register-view">
      <h3 className="register-view-title">Registers</h3>

      {/* General Purpose Registers */}
      <div className="register-grid">
        {Registers.map((value, index) => (
          <div
            key={index}
            className={`register-row ${effectiveHighlights.has(index) ? 'register-changed' : ''}`}
          >
            <span className="register-name">R{index}</span>
            <span className="register-value">{formatHex32(value)}</span>
            <span className="register-decimal">({value})</span>
          </div>
        ))}
      </div>

      {/* CPSR Flags */}
      <div className="cpsr-section">
        <h4 className="cpsr-title">CPSR Flags</h4>
        <div className="cpsr-flags">
          <span className={`flag ${CPSR.N ? 'flag-set' : 'flag-clear'}`}>
            N: {CPSR.N ? '1' : '0'}
          </span>
          <span className={`flag ${CPSR.Z ? 'flag-set' : 'flag-clear'}`}>
            Z: {CPSR.Z ? '1' : '0'}
          </span>
          <span className={`flag ${CPSR.C ? 'flag-set' : 'flag-clear'}`}>
            C: {CPSR.C ? '1' : '0'}
          </span>
          <span className={`flag ${CPSR.V ? 'flag-set' : 'flag-clear'}`}>
            V: {CPSR.V ? '1' : '0'}
          </span>
        </div>
      </div>

      {/* Special Registers */}
      <div className="special-registers">
        <div className="register-row">
          <span className="register-name">PC</span>
          <span className="register-value">{formatHex32(PC)}</span>
        </div>
        <div className="register-row">
          <span className="register-name">Cycles</span>
          <span className="register-value">{Cycles}</span>
        </div>
      </div>
    </div>
  )
}
