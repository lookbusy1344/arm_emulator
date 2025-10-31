import React, { useEffect, useState, useRef } from 'react';
import { EventsOn, EventsOff } from '../../wailsjs/runtime/runtime';
import { GetStack, GetRegisters } from '../../wailsjs/go/main/App';
import { service } from '../../wailsjs/go/models';
import './StackView.css';

interface StackEntryView extends service.StackEntry {
  isSP: boolean;
  isChanged: boolean;
}

export const StackView: React.FC = () => {
  const [entries, setEntries] = useState<StackEntryView[]>([]);
  const previousValuesRef = useRef<Map<number, number>>(new Map());

  const loadStack = async () => {
    try {
      const registerState = await GetRegisters();
      const sp = registerState.Registers[13]; // R13 is SP
      const stackData = await GetStack(0, 16);

      const entriesWithSP = stackData.map(entry => ({
        ...entry,
        isSP: entry.address === sp,
        isChanged: previousValuesRef.current.has(entry.address) && 
                   previousValuesRef.current.get(entry.address) !== entry.value,
      }));

      const newValues = new Map<number, number>();
      stackData.forEach(entry => newValues.set(entry.address, entry.value));
      previousValuesRef.current = newValues;

      setEntries(entriesWithSP);
    } catch (error) {
      console.error('Failed to load stack:', error);
    }
  };

  useEffect(() => {
    loadStack();
    EventsOn('vm:state-changed', loadStack);

    return () => {
      EventsOff('vm:state-changed', loadStack);
    };
  }, []);

  return (
    <div className="stack-view">
      <div className="stack-header">Stack</div>
      <div className="stack-content">
        {entries.map((entry, index) => (
          <div key={index} className={`stack-entry ${entry.isSP ? 'stack-entry-sp' : ''}`}>
            {entry.isSP && <span className="sp-marker">→</span>}
            <span className="stack-address">
              {entry.address.toString(16).padStart(8, '0')}
            </span>
            <span className={`stack-value ${entry.isChanged ? 'stack-value-changed' : ''}`}>
              {entry.value.toString(16).padStart(8, '0')}
            </span>
            {entry.symbol && (
              <span className="stack-symbol">{entry.symbol}</span>
            )}
          </div>
        ))}
      </div>
    </div>
  );
};
