import { useState } from 'react';
import { Allotment } from 'allotment';
import 'allotment/dist/style.css';
import { SourceView } from './components/SourceView';
import { DisassemblyView } from './components/DisassemblyView';
import { RegisterView } from './components/RegisterView';
import { MemoryView } from './components/MemoryView';
import { StackView } from './components/StackView';
import { OutputView } from './components/OutputView';
import { StatusView } from './components/StatusView';
import { BreakpointsView } from './components/BreakpointsView';
import { CommandInput } from './components/CommandInput';
import { ExpressionEvaluator } from './components/ExpressionEvaluator';
import {
  Step,
  StepOver,
  StepOut,
  Continue,
  Pause,
  Reset
} from '../wailsjs/go/main/App';
import type { RegisterState } from './types/emulator';
import './App.css';

function App() {
  const [leftTab, setLeftTab] = useState<'source' | 'disassembly'>('source');
  const [bottomTab, setBottomTab] = useState<'output' | 'breakpoints' | 'status' | 'expressions'>('output');

  // Placeholder register state until we wire up real data
  const [registers] = useState<RegisterState>({
    Registers: Array(16).fill(0),
    CPSR: { N: false, Z: false, C: false, V: false },
    PC: 0,
    Cycles: 0
  });

  // Placeholder memory state until we wire up real data
  const [memory] = useState<Uint8Array>(new Uint8Array(256));
  const [memoryAddress, setMemoryAddress] = useState<number>(0);

  const handleStep = async () => {
    try {
      await Step();
    } catch (error) {
      console.error('Step failed:', error);
    }
  };

  const handleStepOver = async () => {
    try {
      await StepOver();
    } catch (error) {
      console.error('Step Over failed:', error);
    }
  };

  const handleStepOut = async () => {
    try {
      await StepOut();
    } catch (error) {
      console.error('Step Out failed:', error);
    }
  };

  const handleRun = async () => {
    try {
      await Continue();
    } catch (error) {
      console.error('Continue failed:', error);
    }
  };

  const handlePause = async () => {
    try {
      await Pause();
    } catch (error) {
      console.error('Pause failed:', error);
    }
  };

  const handleReset = async () => {
    try {
      await Reset();
    } catch (error) {
      console.error('Reset failed:', error);
    }
  };

  return (
    <div className="app-container">
      <Allotment vertical>
        {/* Top toolbar - fixed height */}
        <Allotment.Pane snap minSize={60} maxSize={60}>
          <div className="toolbar">
            <button onClick={() => {}}>Load</button>
            <button onClick={handleStep}>Step</button>
            <button onClick={handleStepOver}>Step Over</button>
            <button onClick={handleStepOut}>Step Out</button>
            <button onClick={handleRun}>Run</button>
            <button onClick={handlePause}>Pause</button>
            <button onClick={handleReset}>Reset</button>
          </div>
        </Allotment.Pane>

        {/* Main content area */}
        <Allotment.Pane>
          <Allotment>
            {/* Left: Source/Disassembly tabs */}
            <Allotment.Pane minSize={300} preferredSize={500}>
              <div className="tabbed-panel">
                <div className="tabs">
                  <button
                    className={leftTab === 'source' ? 'tab active' : 'tab'}
                    onClick={() => setLeftTab('source')}
                  >
                    Source
                  </button>
                  <button
                    className={leftTab === 'disassembly' ? 'tab active' : 'tab'}
                    onClick={() => setLeftTab('disassembly')}
                  >
                    Disassembly
                  </button>
                </div>
                <div className="tab-content">
                  {leftTab === 'source' && <SourceView />}
                  {leftTab === 'disassembly' && <DisassemblyView />}
                </div>
              </div>
            </Allotment.Pane>

            {/* Right: Registers/Memory/Stack */}
            <Allotment.Pane minSize={300} preferredSize={400}>
              <Allotment vertical>
                <Allotment.Pane>
                  <RegisterView registers={registers} />
                </Allotment.Pane>
                <Allotment.Pane>
                  <MemoryView
                    memory={memory}
                    baseAddress={memoryAddress}
                    onAddressChange={setMemoryAddress}
                  />
                </Allotment.Pane>
                <Allotment.Pane>
                  <StackView />
                </Allotment.Pane>
              </Allotment>
            </Allotment.Pane>
          </Allotment>
        </Allotment.Pane>

        {/* Bottom: Output/Breakpoints/Status/Expressions tabs */}
        <Allotment.Pane snap minSize={150} preferredSize={200}>
          <div className="tabbed-panel">
            <div className="tabs">
              <button
                className={bottomTab === 'output' ? 'tab active' : 'tab'}
                onClick={() => setBottomTab('output')}
              >
                Output
              </button>
              <button
                className={bottomTab === 'breakpoints' ? 'tab active' : 'tab'}
                onClick={() => setBottomTab('breakpoints')}
              >
                Breakpoints
              </button>
              <button
                className={bottomTab === 'status' ? 'tab active' : 'tab'}
                onClick={() => setBottomTab('status')}
              >
                Status
              </button>
              <button
                className={bottomTab === 'expressions' ? 'tab active' : 'tab'}
                onClick={() => setBottomTab('expressions')}
              >
                Expressions
              </button>
            </div>
            <div className="tab-content">
              {bottomTab === 'output' && <OutputView />}
              {bottomTab === 'breakpoints' && <BreakpointsView />}
              {bottomTab === 'status' && <StatusView />}
              {bottomTab === 'expressions' && <ExpressionEvaluator />}
            </div>
          </div>
        </Allotment.Pane>

        {/* Command Input - fixed at very bottom */}
        <Allotment.Pane snap minSize={40} maxSize={40}>
          <CommandInput />
        </Allotment.Pane>
      </Allotment>
    </div>
  );
}

export default App;
