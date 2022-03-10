import * as React from 'react';
import { Theme, Radio, RadioGroup, Slider } from '@material-ui/core';
import { makeStyles, withStyles } from '@material-ui/styles';
import FormControlLabel from '@material-ui/core/FormControlLabel';
import { COLOR_SPECTRUM } from 'components/Theme/colorSpectrum';
import { TimeZone } from './helpers';
import { useScaleContext } from './scaleContext';

function valueText(value: number) {
  return `${value}s`;
}

const useStyles = makeStyles((theme: Theme) => ({
  container: {
    borderTop: `1px solid ${theme.palette.divider}`,
    padding: '20px 24px',
    display: 'flex',
    justifyContent: 'space-between',
    alignItems: 'center'
  }
}));

const CustomSlider = withStyles({
  root: {
    color: COLOR_SPECTRUM.indigo60.color,
    height: 4,
    padding: '15px 0',
    width: 360
  },
  active: {},
  valueLabel: {
    left: 'calc(-50% + 12px)',
    color: COLOR_SPECTRUM.black.color,
    top: -22,
    '& *': {
      background: 'transparent',
      color: COLOR_SPECTRUM.black.color
    }
  },
  track: {
    height: 4
  },
  rail: {
    height: 4,
    opacity: 0.5,
    backgroundColor: COLOR_SPECTRUM.gray20.color
  },
  mark: {
    backgroundColor: COLOR_SPECTRUM.gray20.color,
    height: 8,
    width: 2,
    marginTop: -2
  },
  markLabel: {
    top: -6,
    fontSize: 12,
    color: COLOR_SPECTRUM.gray40.color
  },
  markActive: {
    opacity: 1,
    backgroundColor: 'currentColor'
  },
  marked: {
    marginBottom: 0
  }
})(Slider);

interface ExecutionTimelineFooterProps {
  onTimezoneChange?: (timezone: string) => void;
}

export const ExecutionTimelineFooter: React.FC<ExecutionTimelineFooterProps> = ({ onTimezoneChange }) => {
  const styles = useStyles();
  const [timezone, setTimezone] = React.useState(TimeZone.Local);

  const timeScale = useScaleContext();

  const handleTimezoneChange = (event: React.ChangeEvent<HTMLInputElement>) => {
    const newTimezone = (event.target as HTMLInputElement).value;
    setTimezone(newTimezone);
    if (onTimezoneChange) {
      onTimezoneChange(newTimezone);
    }
  };

  const handleTimeIntervalChange = (event, newValue) => {
    timeScale.setScaleFactor(newValue);
  };

  return (
    <div className={styles.container}>
      <CustomSlider
        value={timeScale.scaleFactor}
        onChange={handleTimeIntervalChange}
        marks={timeScale.marks}
        max={timeScale.marks.length - 1}
        ValueLabelComponent={({ children }) => <>{children}</>}
        valueLabelDisplay="on"
        getAriaValueText={valueText}
      />
      <RadioGroup row aria-label="timezone" name="timezone" value={timezone} onChange={handleTimezoneChange}>
        <FormControlLabel value={TimeZone.Local} control={<Radio />} label="Local Time" />
        <FormControlLabel value={TimeZone.UTC} control={<Radio />} label="UTC" />
      </RadioGroup>
    </div>
  );
};
