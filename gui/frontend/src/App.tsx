import { useState, useEffect, useCallback } from 'react';
import { Allotment } from 'allotment';
import 'allotment/dist/style.css';
import { SourceView } from './components/SourceView';
import { DisassemblyView } from './components/DisassemblyView';
import { RegisterView } from './components/RegisterView';
import { MemoryStackContainer } from './components/MemoryStackContainer';
import { OutputView } from './components/OutputView';
import { StatusView } from './components/StatusView';
import { BreakpointsView } from './components/BreakpointsView';
import { CommandInput } from './components/CommandInput';
import { ExpressionEvaluator } from './components/ExpressionEvaluator';
import { ProgramInput } from './components/ProgramInput';
import {
  Step,
  StepOver,
  StepOut,
  Continue,
  Pause,
  Reset,
  LoadProgramFromFile,
  GetRegisters,
  ToggleBreakpoint
} from '../wailsjs/go/main/App';
import './App.css';

function App() {
  const [leftTab, setLeftTab] = useState<'source' | 'disassembly'>('source');
  const [bottomTab, setBottomTab] = useState<'output' | 'breakpoints' | 'status' | 'expressions'>('output');

  const handleStep = useCallback(async () => {
    try {
      await Step();
    } catch (error) {
      console.error('Step failed:', error);
    }
  }, []);

  const handleStepOver = useCallback(async () => {
    try {
      await StepOver();
    } catch (error) {
      console.error('Step Over failed:', error);
    }
  }, []);

  const handleStepOut = useCallback(async () => {
    try {
      await StepOut();
    } catch (error) {
      console.error('Step Out failed:', error);
    }
  }, []);

  const handleRun = useCallback(async () => {
    try {
      await Continue();
    } catch (error) {
      console.error('Continue failed:', error);
    }
  }, []);

  const handlePause = useCallback(async () => {
    try {
      await Pause();
    } catch (error) {
      console.error('Pause failed:', error);
    }
  }, []);

  const handleReset = useCallback(async () => {
    try {
      await Reset();
    } catch (error) {
      console.error('Reset failed:', error);
    }
  }, []);

  const handleLoad = useCallback(async () => {
    try {
      await LoadProgramFromFile();
    } catch (error) {
      console.error('Load failed:', error);
    }
  }, []);

  const handleToggleBreakpoint = useCallback(async () => {
    try {
      const regs = await GetRegisters();
      await ToggleBreakpoint(regs.PC);
    } catch (error) {
      console.error('Toggle breakpoint failed:', error);
    }
  }, []);

  // Keyboard shortcuts matching TUI
  useEffect(() => {
    const handleKeyDown = (e: KeyboardEvent) => {
      // Ignore if typing in input fields
      if (e.target instanceof HTMLInputElement || e.target instanceof HTMLTextAreaElement) {
        return;
      }

      switch (e.key) {
        case 'F5':
          e.preventDefault();
          handleRun();
          break;
        case 'F9':
          e.preventDefault();
          handleToggleBreakpoint();
          break;
        case 'F10':
          e.preventDefault();
          handleStepOver();
          break;
        case 'F11':
          e.preventDefault();
          handleStep();
          break;
      }
    };

    window.addEventListener('keydown', handleKeyDown);
    return () => window.removeEventListener('keydown', handleKeyDown);
  }, [handleRun, handleToggleBreakpoint, handleStepOver, handleStep]);

  return (
    <div className="app-container">
      <Allotment vertical>
        {/* Top toolbar - fixed height */}
        <Allotment.Pane snap minSize={60} maxSize={60}>
          <div className="toolbar" data-testid="toolbar">
            <button onClick={handleLoad}>Load</button>
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

            {/* Right: Registers/Memory+Stack */}
            <Allotment.Pane minSize={300} preferredSize={400}>
              <Allotment vertical>
                <Allotment.Pane>
                  <RegisterView />
                </Allotment.Pane>
                <Allotment.Pane>
                  <MemoryStackContainer />
                </Allotment.Pane>
              </Allotment>
            </Allotment.Pane>
          </Allotment>
        </Allotment.Pane>

        {/* Bottom: Output/Breakpoints/Status/Expressions tabs */}
        <Allotment.Pane snap minSize={120} preferredSize={130} maxSize={140}>
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
              <div style={{ display: bottomTab === 'output' ? 'flex' : 'none', flexDirection: 'column', height: '100%' }}>
                <OutputView />
              </div>
              <div style={{ display: bottomTab === 'breakpoints' ? 'flex' : 'none', flexDirection: 'column', height: '100%' }} data-testid="breakpoints-tab-content">
                <BreakpointsView />
              </div>
              <div style={{ display: bottomTab === 'status' ? 'flex' : 'none', flexDirection: 'column', height: '100%' }}>
                <StatusView />
              </div>
              <div style={{ display: bottomTab === 'expressions' ? 'flex' : 'none', flexDirection: 'column', height: '100%' }}>
                <ExpressionEvaluator />
              </div>
            </div>
          </div>
        </Allotment.Pane>

        {/* Program Input - for guest program stdin */}
        <Allotment.Pane snap minSize={50} preferredSize={50}>
          <ProgramInput />
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
