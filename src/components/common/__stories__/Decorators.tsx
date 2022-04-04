import * as React from 'react';

const PANEL_VIEW = '420px';

export function PanelViewDecorator(Story: any) {
  return (
    <div style={{ margin: '8px' }}>
      <div style={{ display: 'flex', justifyContent: 'space-between' }}>
        <span>Panel View</span>
        <span>{PANEL_VIEW}</span>
      </div>
      <div style={{ border: 'solid 1px lightblue', width: PANEL_VIEW }}>
        <Story />
      </div>
    </div>
  );
}
