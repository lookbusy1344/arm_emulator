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

    EventsOn('vm:state-changed', () => {
      loadState();
      addMessage('info', 'VM state changed');
    });

    EventsOn('vm:error', (errorMsg: string) => {
      addMessage('error', errorMsg);
    });

    EventsOn('vm:breakpoint-hit', () => {
      addMessage('breakpoint', 'Breakpoint hit');
    });

    return () => {
      EventsOff('vm:state-changed');
      EventsOff('vm:error');
      EventsOff('vm:breakpoint-hit');
    };
  }, []);

  return (
    <div className="status-view">
      <div className="status-header">
        <span>Debugger Status</span>
        <div className="status-info">
          <span className="status-state">{executionState}</span>
          <span className="status-cycles">Cycles: {cycles}</span>
        </div>
      </div>
      <div className="status-content">
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
