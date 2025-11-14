import React, { useState, KeyboardEvent } from 'react';
import { SendInput } from '../../wailsjs/go/main/App';
import './ProgramInput.css';

export const ProgramInput: React.FC = () => {
  const [input, setInput] = useState('');

  const handleSend = async () => {
    if (input.trim()) {
      try {
        await SendInput(input);
        setInput(''); // Clear input after sending
      } catch (error) {
        console.error('Failed to send input:', error);
      }
    }
  };

  const handleKeyDown = async (e: KeyboardEvent<HTMLInputElement>) => {
    if (e.key === 'Enter') {
      e.preventDefault();
      await handleSend();
    }
  };

  return (
    <div className="program-input" data-testid="program-input">
      <label htmlFor="stdin-input" className="program-input-label">
        Program Input:
      </label>
      <input
        id="stdin-input"
        type="text"
        className="program-input-field"
        value={input}
        onChange={(e) => setInput(e.target.value)}
        onKeyDown={handleKeyDown}
        placeholder="Type input for program and press Enter..."
      />
      <button
        className="program-input-send"
        onClick={handleSend}
        disabled={!input.trim()}
      >
        Send
      </button>
    </div>
  );
};
