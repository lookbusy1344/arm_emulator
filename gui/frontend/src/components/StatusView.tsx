import React, { useEffect, useState } from 'react';
import { EventsOn, EventsOff } from '../../wailsjs/runtime/runtime';
import { GetExecutionState, GetRegisters } from '../../wailsjs/go/main/App';
import './StatusView.css';

interface StatusMessage {
  type: 'info' | 'error' | 'breakpoint';
  message: string;
  timestamp: Date;
}

export const StatusView: React.FC = () => {
  const [messages, setMessages] = useState<StatusMessage[]>([]);
  const [executionState, setExecutionState] = useState<string>('');
  const [cycles, setCycles] = useState<number>(0);

  const addMessage = (type: StatusMessage['type'], message: string) => {
    setMessages(prev => [
      ...prev,
      { type, message, timestamp: new Date() }
    ].slice(-50)); // Keep last 50 messages
  };

  const loadState = async () => {
    try {
      const state = await GetExecutionState();
      const registerState = await GetRegisters();

      setExecutionState(state);
      setCycles(registerState.Cycles || 0);
    } catch (error) {
      console.error('Failed to load state:', error);
    }
  };

  useEffect(() => {
    loadState();

    const handleStateChange = () => {
      loadState();
      addMessage('info', 'VM state changed');
    };

    const handleError = (errorMsg: string) => {
      addMessage('error', errorMsg);
    };

    const handleBreakpoint = () => {
      addMessage('breakpoint', 'Breakpoint hit');
    };

    const unsubscribe1 = EventsOn('vm:state-changed', handleStateChange);
    const unsubscribe2 = EventsOn('vm:error', handleError);
    const unsubscribe3 = EventsOn('vm:breakpoint-hit', handleBreakpoint);

    return () => {
      unsubscribe1();
      unsubscribe2();
      unsubscribe3();
    };
  }, []);

  return (
    <div className="status-view" data-testid="status-view">
      <div className="status-info-bar">
        <span className="status-state" data-testid="execution-status">{executionState}</span>
        <span className="status-cycles">Cycles: {cycles}</span>
      </div>
      <div className="status-content" data-testid="status-tab-content">
        {messages.map((msg, index) => (
          <div key={index} className={`status-message status-message-${msg.type}`}>
            <span className="status-timestamp">
              {msg.timestamp.toLocaleTimeString()}
            </span>
            <span className="status-text">{msg.message}</span>
          </div>
        ))}
      </div>
    </div>
  );
};
