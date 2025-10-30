import React, { useEffect, useState, useRef } from 'react';
import { EventsOn, EventsOff } from '../../wailsjs/runtime/runtime';
import { GetSourceMap, GetRegisters, ToggleBreakpoint, GetBreakpoints, GetSymbolForAddress } from '../../wailsjs/go/main/App';
import './SourceView.css';

interface SourceLine {
  address: number;
  source: string;
  hasBreakpoint: boolean;
  isCurrent: boolean;
  symbol: string;
}

export const SourceView: React.FC = () => {
  const [lines, setLines] = useState<SourceLine[]>([]);
  const [currentPC, setCurrentPC] = useState<number>(0);
  const containerRef = useRef<HTMLDivElement>(null);

  const loadSourceData = async () => {
    try {
      const sourceMap = await GetSourceMap();
      const registerState = await GetRegisters();
      const breakpoints = await GetBreakpoints();

      const pc = registerState.PC;
      setCurrentPC(pc);

      // Convert source map to sorted array
      const breakpointAddresses = new Set(breakpoints.map(bp => bp.address));

      // Fetch all symbols in parallel
      const entries = Object.entries(sourceMap);
      const symbolPromises = entries.map(([addrStr]) =>
        GetSymbolForAddress(parseInt(addrStr))
      );
      const symbols = await Promise.all(symbolPromises);

      const sourceLines: SourceLine[] = entries.map(([addrStr, source], index) => {
        const address = parseInt(addrStr);
        return {
          address,
          source,
          hasBreakpoint: breakpointAddresses.has(address),
          isCurrent: address === pc,
          symbol: symbols[index] || '',
        };
      });

      // Sort by address
      sourceLines.sort((a, b) => a.address - b.address);

      setLines(sourceLines);

      // Auto-scroll to current PC
      setTimeout(() => scrollToCurrentLine(), 100);
    } catch (error) {
      console.error('Failed to load source data:', error);
    }
  };

  const scrollToCurrentLine = () => {
    if (containerRef.current) {
      const currentLine = containerRef.current.querySelector('.source-line-current');
      if (currentLine) {
        currentLine.scrollIntoView({ behavior: 'smooth', block: 'center' });
      }
    }
  };

  const handleLineClick = async (address: number) => {
    try {
      await ToggleBreakpoint(address);
      // State will update via event
    } catch (error) {
      console.error('Failed to toggle breakpoint:', error);
    }
  };

  useEffect(() => {
    // Initial load
    loadSourceData();

    // Subscribe to VM state changes
    EventsOn('vm:state-changed', loadSourceData);

    return () => {
      EventsOff('vm:state-changed');
    };
  }, []);

  return (
    <div className="source-view" ref={containerRef}>
      <div className="source-header">Source Code</div>
      <div className="source-content">
        {lines.map((line, index) => (
          <div
            key={index}
            className={`source-line ${line.isCurrent ? 'source-line-current' : ''} ${line.hasBreakpoint ? 'source-line-breakpoint' : ''}`}
            onClick={() => handleLineClick(line.address)}
          >
            <span className="source-line-number">
              {line.hasBreakpoint && <span className="breakpoint-marker">●</span>}
              {line.address.toString(16).padStart(8, '0')}
            </span>
            {line.symbol && (
              <span className="source-line-symbol">{line.symbol}:</span>
            )}
            <span className="source-line-text">{line.source}</span>
          </div>
        ))}
      </div>
    </div>
  );
};
