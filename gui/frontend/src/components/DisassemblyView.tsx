import React, { useEffect, useState } from 'react';
import { EventsOn, EventsOff } from '../../wailsjs/runtime/runtime';
import { GetDisassembly, GetRegisters, GetBreakpoints, ToggleBreakpoint } from '../../wailsjs/go/main/App';
import { service } from '../../wailsjs/go/models';
import './DisassemblyView.css';

interface DisassemblyLineView extends service.DisassemblyLine {
  hasBreakpoint: boolean;
  isCurrent: boolean;
}

export const DisassemblyView: React.FC = () => {
  const [lines, setLines] = useState<DisassemblyLineView[]>([]);

  const loadDisassembly = async () => {
    try {
      const registerState = await GetRegisters();
      const breakpoints = await GetBreakpoints();
      const pc = registerState.PC;

      // Get disassembly around PC
      const startAddr = Math.max(0, pc - 20 * 4);
      const disasm = await GetDisassembly(startAddr, 15);

      const breakpointAddresses = new Set(breakpoints.map(bp => bp.address));

      const linesWithState = disasm.map(line => ({
        ...line,
        hasBreakpoint: breakpointAddresses.has(line.address),
        isCurrent: line.address === pc,
      }));

      setLines(linesWithState);
    } catch (error) {
      console.error('Failed to load disassembly:', error);
    }
  };

  const handleLineClick = async (address: number) => {
    try {
      await ToggleBreakpoint(address);
    } catch (error) {
      console.error('Failed to toggle breakpoint:', error);
    }
  };

  useEffect(() => {
    loadDisassembly();
    EventsOn('vm:state-changed', loadDisassembly);

    return () => {
      EventsOff('vm:state-changed');
    };
  }, []);

  return (
    <div className="disassembly-view">
      <div className="disassembly-header">Disassembly</div>
      <div className="disassembly-content">
        {lines.map((line, index) => (
          <div
            key={index}
            className={`disasm-line ${line.isCurrent ? 'disasm-line-current' : ''} ${line.hasBreakpoint ? 'disasm-line-breakpoint' : ''}`}
            onClick={() => handleLineClick(line.address)}
          >
            <span className="disasm-address">
              {line.hasBreakpoint && <span className="breakpoint-marker">●</span>}
              {line.address.toString(16).padStart(8, '0')}
            </span>
            <span className="disasm-opcode">
              {line.opcode.toString(16).padStart(8, '0')}
            </span>
            {line.symbol && (
              <span className="disasm-symbol">{line.symbol}:</span>
            )}
          </div>
        ))}
      </div>
    </div>
  );
};
