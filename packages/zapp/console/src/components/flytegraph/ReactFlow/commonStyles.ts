import { CSSProperties } from 'react';

const positionStyle: CSSProperties = {
  bottom: '1rem',
  zIndex: 10,
  position: 'absolute',
  maxHeight: '520px',
};

export const leftPositionStyle: CSSProperties = {
  ...positionStyle,
  left: '1rem',
  width: '336px',
};

export const rightPositionStyle: CSSProperties = {
  ...positionStyle,
  right: '1rem',
  width: '150px',
};

export const graphButtonContainer: CSSProperties = {
  width: '100%',
};

export const graphButtonStyle: CSSProperties = {
  color: '#555',
  width: '100%',
};

export const popupContainerStyle: CSSProperties = {
  width: '100%',
  padding: '1rem',
  background: 'rgba(255,255,255,1)',
  border: `1px solid #ddd`,
  borderRadius: '4px',
  boxShadow: '2px 4px 10px rgba(50,50,50,.2)',
  marginBottom: '1rem',
};
