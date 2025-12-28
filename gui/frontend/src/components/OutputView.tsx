import React, { useEffect, useState, useRef } from 'react';
import { EventsOn } from '../../wailsjs/runtime/runtime';
import './OutputView.css';

// Maximum output buffer size (1MB) to prevent memory exhaustion from long-running programs
const MAX_OUTPUT_SIZE = 1024 * 1024;

export const OutputView: React.FC = () => {
  const [output, setOutput] = useState<string>('');
  const contentRef = useRef<HTMLPreElement>(null);

  const handleOutputEvent = (data: string) => {
    setOutput(prev => {
      const newOutput = prev + data;
      // Trim from the beginning if output exceeds max size
      if (newOutput.length > MAX_OUTPUT_SIZE) {
        // Keep the last MAX_OUTPUT_SIZE characters
        return newOutput.slice(-MAX_OUTPUT_SIZE);
      }
      return newOutput;
    });

    // Auto-scroll to bottom
    setTimeout(() => {
      if (contentRef.current) {
        contentRef.current.scrollTop = contentRef.current.scrollHeight;
      }
    }, 10);
  };

  const handleClear = () => {
    setOutput('');
  };

  useEffect(() => {
    const unsubscribe = EventsOn('vm:output', handleOutputEvent);

    return () => {
      unsubscribe();
    };
  }, []);

  return (
    <div className="output-view" data-testid="output-view">
      <div className="output-toolbar">
        <button className="output-clear-btn" onClick={handleClear}>Clear</button>
      </div>
      <pre className="output-content" ref={contentRef}>
        {output || '(no output)'}
      </pre>
    </div>
  );
};
