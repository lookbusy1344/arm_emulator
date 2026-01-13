import React, { useEffect, useState, useRef } from 'react';
import { EventsOn, EventsOff } from '../../wailsjs/runtime/runtime';
import { GetSourceMap, GetRegisters, ToggleBreakpoint, GetBreakpoints, GetSymbolsForAddresses } from '../../wailsjs/go/main/App';
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

      // Fetch all symbols in a single batch API call
      const addresses = sourceMap.map(entry => entry.address);
      const symbolMap = await GetSymbolsForAddresses(addresses);

      const sourceLines: SourceLine[] = sourceMap.map(entry => {
        return {
          address: entry.address,
          source: entry.line,
          hasBreakpoint: breakpointAddresses.has(entry.address),
          isCurrent: entry.address === pc,
          symbol: symbolMap[entry.address] || '',
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
    const unsubscribe = EventsOn('vm:state-changed', loadSourceData);

    return () => {
      unsubscribe();
    };
  }, []);

  return (
    <div className="source-view" data-testid="source-view" ref={containerRef}>
      <div className="source-content">
        {lines.map((line, index) => (
          <div
            key={index}
            className={`source-line ${line.isCurrent ? 'source-line-current' : ''} ${line.hasBreakpoint ? 'source-line-breakpoint' : ''}`}
            onDoubleClick={() => handleLineClick(line.address)}
          >
            <span className="source-line-number">
              {line.hasBreakpoint && <span className="breakpoint-marker" data-testid="breakpoint-indicator">●</span>}
              {line.address.toString(16).padStart(8, '0')}
            </span>
            <span className="source-line-symbol">
              {line.symbol ? `${line.symbol}:` : ''}
            </span>
            <span className="source-line-text">{line.source}</span>
          </div>
        ))}
      </div>
    </div>
  );
};
