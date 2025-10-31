import React, { useState, useEffect } from 'react';
import { EvaluateExpression } from '../../wailsjs/go/main/App';
import { EventsOn, EventsOff } from '../../wailsjs/runtime/runtime';
import './ExpressionEvaluator.css';

interface EvaluationResult {
  expression: string;
  value?: number;
  error?: string;
  timestamp: number;
}

const MAX_RESULTS = 10;

export const ExpressionEvaluator: React.FC = () => {
  const [input, setInput] = useState<string>('');
  const [results, setResults] = useState<EvaluationResult[]>([]);

  const handleEvaluate = async () => {
    if (!input.trim()) {
      return;
    }

    try {
      const value = await EvaluateExpression(input);

      setResults(prev => {
        const newResult: EvaluationResult = {
          expression: input,
          value,
          timestamp: Date.now()
        };
        return [newResult, ...prev].slice(0, MAX_RESULTS);
      });

      setInput('');
    } catch (error) {
      setResults(prev => {
        const newResult: EvaluationResult = {
          expression: input,
          error: String(error),
          timestamp: Date.now()
        };
        return [newResult, ...prev].slice(0, MAX_RESULTS);
      });
    }
  };

  const handleSubmit = (e: React.FormEvent) => {
    e.preventDefault();
    handleEvaluate();
  };

  const handleKeyDown = (e: React.KeyboardEvent<HTMLInputElement>) => {
    if (e.key === 'Enter') {
      e.preventDefault();
      handleEvaluate();
    }
  };

  const handleClear = () => {
    setResults([]);
  };

  // Clear on VM state change
  useEffect(() => {
    const handleStateChange = () => {
      setResults([]);
    };

    const unsubscribe = EventsOn('vm:state-changed', handleStateChange);

    return () => {
      unsubscribe();
    };
  }, []);

  return (
    <div className="expression-evaluator">
      <form onSubmit={handleSubmit} className="expression-form">
        <input
          type="text"
          value={input}
          onChange={(e) => setInput(e.target.value)}
          onKeyDown={handleKeyDown}
          placeholder="Enter expression (e.g., R0, PC, R1+4, 0x8000)..."
          className="expression-input"
        />
        <button type="submit" className="expression-eval-btn">
          Evaluate
        </button>
        <button type="button" className="expression-clear-btn" onClick={handleClear}>Clear</button>
      </form>

      <div className="expression-results">
        {results.length === 0 ? (
          <div className="expression-no-results">No evaluations yet</div>
        ) : (
          results.map((result, index) => (
            <div
              key={result.timestamp}
              className={`expression-result-item ${result.error ? 'error' : ''}`}
            >
              <div className="expression-result-header">
                <span className="expression-result-index">#{results.length - index}</span>
                <span className="expression-result-expr">{result.expression}</span>
              </div>
              <div className="expression-result-value">
                {result.error ? (
                  <span className="expression-error">{result.error}</span>
                ) : (
                  <>
                    <span className="expression-hex">0x{result.value?.toString(16).toUpperCase().padStart(8, '0')}</span>
                    <span className="expression-decimal">({result.value})</span>
                  </>
                )}
              </div>
            </div>
          ))
        )}
      </div>
    </div>
  );
};
