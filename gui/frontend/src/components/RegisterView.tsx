import React from 'react'
import type { RegisterState } from '../types/emulator'
import './RegisterView.css'

interface RegisterViewProps {
  registers: RegisterState
  changedRegisters?: Set<number>
}

const formatHex32 = (value: number): string => {
  return `0x${value.toString(16).toUpperCase().padStart(8, '0')}`
}

export const RegisterView: React.FC<RegisterViewProps> = ({
  registers,
  changedRegisters = new Set(),
}) => {
  const { Registers, CPSR, PC, Cycles } = registers

  return (
    <div className="register-view">
      <h3 className="register-view-title">Registers</h3>

      {/* General Purpose Registers */}
      <div className="register-grid">
        {Registers.map((value, index) => (
          <div
            key={index}
            className={`register-row ${changedRegisters.has(index) ? 'register-changed' : ''}`}
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
