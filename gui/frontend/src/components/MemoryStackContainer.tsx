import React, { useState } from 'react';
import { MemoryContainer } from './MemoryContainer';
import { StackView } from './StackView';

export const MemoryStackContainer: React.FC = () => {
  const [rightTab, setRightTab] = useState<'memory' | 'stack'>('memory');

  return (
    <div className="tabbed-panel">
      <div className="tabs">
        <button
          className={rightTab === 'memory' ? 'tab active' : 'tab'}
          onClick={() => setRightTab('memory')}
          data-testid="memory-tab"
        >
          Memory
        </button>
        <button
          className={rightTab === 'stack' ? 'tab active' : 'tab'}
          onClick={() => setRightTab('stack')}
          data-testid="stack-tab"
        >
          Stack
        </button>
      </div>
      <div className="tab-content">
        <div style={{ display: rightTab === 'memory' ? 'flex' : 'none', flexDirection: 'column', height: '100%' }}>
          <MemoryContainer />
        </div>
        <div style={{ display: rightTab === 'stack' ? 'flex' : 'none', flexDirection: 'column', height: '100%' }}>
          <StackView />
        </div>
      </div>
    </div>
  );
};
