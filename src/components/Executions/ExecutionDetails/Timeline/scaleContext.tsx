import { Mark } from '@material-ui/core/Slider';
import * as React from 'react';
import { createContext, useContext } from 'react';
import { formatSecondsToHmsFormat } from './BarChart/utils';

const MIN_SCALE_VALUE = 60; // 1 min
const MAX_SCALE_VALUE = 3600; // 1h

const DEFAULT_MAX = MIN_SCALE_VALUE;
const DEFAULT_SCALE_FACTOR = 1;

const percentage = [0.02, 0.1, 0.4, 0.6, 0.8, 1];
const maxScaleValue = percentage.length - 1;

interface TimelineScaleState {
  scaleFactor: number;
  chartInterval: number; // value in seconds for one tic of an interval
  marks: Mark[];
  setMaxValue: (newMax: number) => void;
  setScaleFactor: (newScale: number) => void;
}

/** Use this Context to redefine Provider returns in storybooks */
export const ScaleContext = createContext<TimelineScaleState>({
  /** Default values used if ContextProvider wasn't initialized. */
  scaleFactor: DEFAULT_SCALE_FACTOR,
  chartInterval: 20,
  marks: [],
  setMaxValue: () => {
    console.error('ERROR: No ScaleContextProvider was found in parent components.');
  },
  setScaleFactor: () => {
    console.error('ERROR: No ScaleContextProvider was found in parent components.');
  },
});

/** Could be used to access the whole TimelineScaleState */
export const useScaleContext = (): TimelineScaleState => useContext(ScaleContext);

interface ScaleProviderProps {
  children?: React.ReactNode;
}

/** Should wrap "top level" component in Execution view, will build a nodeExecutions tree for specific workflow */
export const ScaleProvider = (props: ScaleProviderProps) => {
  const [maxValue, setMaxValue] = React.useState(DEFAULT_MAX);
  const [scaleFactor, setScaleFactor] = React.useState<number>(DEFAULT_SCALE_FACTOR);
  const [marks, setMarks] = React.useState<Mark[]>([]);
  const [chartInterval, setChartInterval] = React.useState<number>(1);

  React.useEffect(() => {
    const getIntervalValue = (scale: number): number => {
      const intervalValue = Math.ceil(maxValue * percentage[scale]);
      return intervalValue > 0 ? intervalValue : 1;
    };

    setChartInterval(getIntervalValue(scaleFactor));

    const newMarks: Mark[] = [];
    for (let i = 0; i < percentage.length; ++i) {
      newMarks.push({
        value: i,
        label: formatSecondsToHmsFormat(getIntervalValue(i)),
      });
    }
    setMarks(newMarks);
  }, [maxValue, scaleFactor]);

  const setMaxTimeValue = (newMax: number) => {
    // use min and max caps
    let newValue =
      newMax < MIN_SCALE_VALUE
        ? MIN_SCALE_VALUE
        : newMax > MAX_SCALE_VALUE
        ? MAX_SCALE_VALUE
        : newMax;
    // round a value to have full amount of minutes:
    newValue = Math.ceil(newValue / 60) * 60;
    setMaxValue(newValue);
  };

  const setNewScaleFactor = (newScale: number) => {
    const applyScale = newScale < 0 ? 0 : newScale > maxScaleValue ? maxScaleValue : newScale;
    setScaleFactor(applyScale);
  };

  return (
    <ScaleContext.Provider
      value={{
        scaleFactor,
        chartInterval,
        marks,
        setMaxValue: setMaxTimeValue,
        setScaleFactor: setNewScaleFactor,
      }}
    >
      {props.children}
    </ScaleContext.Provider>
  );
};
