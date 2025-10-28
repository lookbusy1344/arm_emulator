import React, { useState } from 'react'
import { RegisterView } from './components/RegisterView'
import { MemoryView } from './components/MemoryView'
import { useEmulator } from './hooks/useEmulator'
import './App.css'

function App() {
  const {
    registers,
    executionState,
    memory,
    memoryAddress,
    error,
    loadProgram,
    step,
    continue: continueExecution,
    pause,
    reset,
    changeMemoryAddress,
  } = useEmulator()

  const [sourceCode, setSourceCode] = useState(`; ARM Assembly Example
_start:
    MOV R0, #42
    MOV R1, #10
    ADD R2, R0, R1
    SWI #0
`)

  const handleLoadProgram = () => {
    loadProgram(sourceCode, 'program.s', 0x8000)
  }

  return (
    <div className="app">
      <header className="app-header">
        <h1 className="app-title">ARM Emulator</h1>
        <div className="app-controls">
          <button onClick={handleLoadProgram} className="btn btn-primary">
            Load
          </button>
          <button onClick={step} className="btn" disabled={executionState === 'running'}>
            Step
          </button>
          <button onClick={continueExecution} className="btn" disabled={executionState === 'running'}>
            Run
          </button>
          <button onClick={pause} className="btn" disabled={executionState !== 'running'}>
            Pause
          </button>
          <button onClick={reset} className="btn">
            Reset
          </button>
          <span className={`status status-${executionState}`}>
            {executionState.toUpperCase()}
          </span>
        </div>
      </header>

      {error && (
        <div className="error-banner">
          <strong>Error:</strong> {error}
        </div>
      )}

      <div className="app-layout">
        <div className="app-editor">
          <textarea
            className="code-editor"
            value={sourceCode}
            onChange={(e) => setSourceCode(e.target.value)}
            spellCheck={false}
          />
        </div>

        <div className="app-panels">
          <div className="panel">
            {registers && <RegisterView registers={registers} />}
          </div>

          <div className="panel">
            <MemoryView
              memory={memory}
              baseAddress={memoryAddress}
              onAddressChange={changeMemoryAddress}
            />
          </div>
        </div>
      </div>
    </div>
  )
}

export default App
