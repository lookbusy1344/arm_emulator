import React, { useState, useEffect, useRef, KeyboardEvent } from 'react';
import { ExecuteCommand } from '../../wailsjs/go/main/App';
import { EventsOn, EventsOff } from '../../wailsjs/runtime/runtime';
import './CommandInput.css';

const MAX_HISTORY = 50;

export const CommandInput: React.FC = () => {
  const [input, setInput] = useState<string>('');
  const [history, setHistory] = useState<string[]>([]);
  const [historyIndex, setHistoryIndex] = useState<number>(-1);
  const [result, setResult] = useState<string>('');
  const inputRef = useRef<HTMLInputElement>(null);

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault();

    if (!input.trim()) {
      return;
    }

    try {
      const response = await ExecuteCommand(input);
      setResult(response);

      // Add to history (avoid duplicates of consecutive commands)
      setHistory(prev => {
        const newHistory = prev[prev.length - 1] === input
          ? prev
          : [...prev, input].slice(-MAX_HISTORY);
        return newHistory;
      });

      setInput('');
      setHistoryIndex(-1);
    } catch (error) {
      setResult(`Error: ${error}`);
    }
  };

  const handleKeyDown = (e: KeyboardEvent<HTMLInputElement>) => {
    if (e.key === 'ArrowUp') {
      e.preventDefault();
      if (history.length === 0) return;

      const newIndex = historyIndex === -1
        ? history.length - 1
        : Math.max(0, historyIndex - 1);

      setHistoryIndex(newIndex);
      setInput(history[newIndex]);
    } else if (e.key === 'ArrowDown') {
      e.preventDefault();
      if (historyIndex === -1) return;

      const newIndex = historyIndex + 1;

      if (newIndex >= history.length) {
        setHistoryIndex(-1);
        setInput('');
      } else {
        setHistoryIndex(newIndex);
        setInput(history[newIndex]);
      }
    }
  };

  const handleClear = () => {
    setResult('');
  };

  // Clear on VM state change
  useEffect(() => {
    const handleStateChange = () => {
      setResult('');
    };

    EventsOn('vm:state-changed', handleStateChange);

    return () => {
      EventsOff('vm:state-changed', handleStateChange);
    };
  }, []);

  return (
    <div className="command-input">
      <form onSubmit={handleSubmit} className="command-form">
        <span className="command-prompt">&gt;</span>
        <input
          ref={inputRef}
          type="text"
          value={input}
          onChange={(e) => setInput(e.target.value)}
          onKeyDown={handleKeyDown}
          placeholder="Enter debugger command..."
          className="command-text-input"
        />
      </form>
      {result && (
        <div className="command-result-container">
          <div className="command-result-header">
            <span>Command Result</span>
            <button className="command-clear-btn" onClick={handleClear}>Clear</button>
          </div>
          <pre className="command-result">{result}</pre>
        </div>
      )}
    </div>
  );
};
