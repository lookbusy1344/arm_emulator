import React, { useEffect, useState } from 'react';
import { EventsOn, EventsOff } from '../../wailsjs/runtime/runtime';
import { GetBreakpoints, RemoveBreakpoint, GetWatchpoints, RemoveWatchpoint } from '../../wailsjs/go/main/App';
import { service } from '../../wailsjs/go/models';
import './BreakpointsView.css';

export const BreakpointsView: React.FC = () => {
  const [breakpoints, setBreakpoints] = useState<service.BreakpointInfo[]>([]);
  const [watchpoints, setWatchpoints] = useState<service.WatchpointInfo[]>([]);

  const loadBreakpoints = async () => {
    try {
      const bps = await GetBreakpoints();
      const wps = await GetWatchpoints();

      setBreakpoints(bps || []);
      setWatchpoints(wps || []);
    } catch (error) {
      console.error('Failed to load breakpoints:', error);
    }
  };

  const handleRemoveBreakpoint = async (address: number) => {
    try {
      await RemoveBreakpoint(address);
    } catch (error) {
      console.error('Failed to remove breakpoint:', error);
    }
  };

  const handleRemoveWatchpoint = async (id: number) => {
    try {
      await RemoveWatchpoint(id);
    } catch (error) {
      console.error('Failed to remove watchpoint:', error);
    }
  };

  useEffect(() => {
    loadBreakpoints();
    EventsOn('vm:state-changed', loadBreakpoints);

    return () => {
      EventsOff('vm:state-changed');
    };
  }, []);

  return (
    <div className="breakpoints-view">
      <div className="breakpoints-header">Breakpoints & Watchpoints</div>

      <div className="breakpoints-section">
        <div className="section-title">Breakpoints ({breakpoints.length})</div>
        {breakpoints.length === 0 ? (
          <div className="empty-message">No breakpoints set</div>
        ) : (
          <table className="breakpoints-table">
            <thead>
              <tr>
                <th>Address</th>
                <th>Condition</th>
                <th>Actions</th>
              </tr>
            </thead>
            <tbody>
              {breakpoints.map((bp, index) => (
                <tr key={index}>
                  <td className="bp-address">0x{bp.address.toString(16).padStart(8, '0')}</td>
                  <td className="bp-condition">{bp.condition || '(always)'}</td>
                  <td className="bp-actions">
                    <button
                      className="btn-remove"
                      onClick={() => handleRemoveBreakpoint(bp.address)}
                    >
                      Remove
                    </button>
                  </td>
                </tr>
              ))}
            </tbody>
          </table>
        )}
      </div>

      <div className="breakpoints-section">
        <div className="section-title">Watchpoints ({watchpoints.length})</div>
        {watchpoints.length === 0 ? (
          <div className="empty-message">No watchpoints set</div>
        ) : (
          <table className="breakpoints-table">
            <thead>
              <tr>
                <th>Address</th>
                <th>Type</th>
                <th>Actions</th>
              </tr>
            </thead>
            <tbody>
              {watchpoints.map((wp, index) => (
                <tr key={index}>
                  <td className="bp-address">0x{wp.address.toString(16).padStart(8, '0')}</td>
                  <td className="bp-type">{wp.type}</td>
                  <td className="bp-actions">
                    <button
                      className="btn-remove"
                      onClick={() => handleRemoveWatchpoint(wp.id)}
                    >
                      Remove
                    </button>
                  </td>
                </tr>
              ))}
            </tbody>
          </table>
        )}
      </div>
    </div>
  );
};
