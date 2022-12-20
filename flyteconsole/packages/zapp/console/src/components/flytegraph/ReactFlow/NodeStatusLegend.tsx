import * as React from 'react';
import { useState, CSSProperties } from 'react';
import { Button } from '@material-ui/core';
import { nodeExecutionPhaseConstants } from 'components/Executions/constants';
import {
  graphButtonContainer,
  graphButtonStyle,
  popupContainerStyle,
  rightPositionStyle,
} from './commonStyles';
import t from './strings';
import { graphNodePhasesList } from './utils';

export const LegendItem = ({ nodeColor, text }) => {
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
    background: isNested ? nodeColor : 'none',
    border: `3px solid ${nodeColor}`,
    borderRadius: '4px',
    paddingRight: '10px',
    marginRight: '1rem',
  };
  return (
    <div style={containerStyle} data-testid="legend-item">
      <div style={colorStyle}></div>
      <div>{text}</div>
    </div>
  );
};

interface LegendProps {
  initialIsVisible?: boolean;
}

export const Legend: React.FC<LegendProps> = ({ initialIsVisible = false }) => {
  const [isVisible, setIsVisible] = useState(initialIsVisible);

  const toggleVisibility = () => {
    setIsVisible(!isVisible);
  };

  const renderLegend = () => (
    <div style={popupContainerStyle} data-testid="legend-table">
      {graphNodePhasesList.map((phase) => {
        return (
          <LegendItem
            {...nodeExecutionPhaseConstants[phase]}
            key={`gl-${nodeExecutionPhaseConstants[phase].text}`}
          />
        );
      })}
      <LegendItem nodeColor="#aaa" text="Nested" />
    </div>
  );

  return (
    <div style={rightPositionStyle}>
      <div>
        {isVisible ? renderLegend() : null}
        <div style={graphButtonContainer}>
          <Button
            style={graphButtonStyle}
            color="default"
            id="graph-show-legend"
            onClick={toggleVisibility}
            variant="contained"
            title={t('legendButton', isVisible)}
          >
            {t('legendButton', isVisible)}
          </Button>
        </div>
      </div>
    </div>
  );
};
