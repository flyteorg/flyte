import * as React from 'react';
import { useState, CSSProperties } from 'react';
import { Button } from '@material-ui/core';
import { nodePhaseColorMapping } from './utils';

export const LegendItem = ({ color, text }) => {
  /**
   * @TODO temporary check for nested graph until
   * nested functionality is deployed
   */
  const isNested = text === 'Nested';

  const containerStyle: CSSProperties = {
    display: 'flex',
    flexDirection: 'row',
    width: '100%',
    padding: '.5rem 0',
  };
  const colorStyle: CSSProperties = {
    width: '28px',
    height: '22px',
    background: isNested ? color : 'none',
    border: `3px solid ${color}`,
    borderRadius: '4px',
    paddingRight: '10px',
    marginRight: '1rem',
  };
  return (
    <div style={containerStyle}>
      <div style={colorStyle}></div>
      <div>{text}</div>
    </div>
  );
};

interface LegendProps {
  initialIsVisible?: boolean;
}

export const Legend: React.FC<LegendProps> = (props) => {
  const { initialIsVisible = false } = props;

  const [isVisible, setIsVisible] = useState(initialIsVisible);

  const positionStyle: CSSProperties = {
    bottom: '1rem',
    right: '1rem',
    zIndex: 10,
    position: 'absolute',
    width: '150px',
  };

  const buttonContainer: CSSProperties = {
    width: '100%',
    display: 'flex',
    justifyContent: 'center',
  };

  const buttonStyle: CSSProperties = {
    color: '#555',
    width: '100%',
  };

  const toggleVisibility = () => {
    setIsVisible(!isVisible);
  };

  const renderLegend = () => {
    const legendContainerStyle: CSSProperties = {
      width: '100%',
      padding: '1rem',
      background: 'rgba(255,255,255,1)',
      border: `1px solid #ddd`,
      borderRadius: '4px',
      boxShadow: '2px 4px 10px rgba(50,50,50,.2)',
      marginBottom: '1rem',
    };

    return (
      <div style={legendContainerStyle}>
        {Object.keys(nodePhaseColorMapping).map((phase) => {
          return (
            <LegendItem
              {...nodePhaseColorMapping[phase]}
              key={`gl-${nodePhaseColorMapping[phase].text}`}
            />
          );
        })}
        <LegendItem color="#aaa" text="Nested" />
      </div>
    );
  };

  return (
    <div style={positionStyle}>
      <div>
        {isVisible ? renderLegend() : null}
        <div style={buttonContainer}>
          <Button
            style={buttonStyle}
            color="default"
            id="graph-show-legend"
            onClick={toggleVisibility}
            variant="contained"
          >
            {isVisible ? 'Hide' : 'Show'} Legend
          </Button>
        </div>
      </div>
    </div>
  );
};
