import React, { useEffect, useState, useRef } from 'react';
import { EventsOn, EventsOff } from '../../wailsjs/runtime/runtime';
import './OutputView.css';

export const OutputView: React.FC = () => {
  const [output, setOutput] = useState<string>('');
  const contentRef = useRef<HTMLPreElement>(null);

  const handleOutputEvent = (data: string) => {
    setOutput(prev => prev + data);

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
    <div className="output-view">
      <div className="output-toolbar">
        <button className="output-clear-btn" onClick={handleClear}>Clear</button>
      </div>
      <pre className="output-content" ref={contentRef}>
        {output || '(no output)'}
      </pre>
    </div>
  );
};
