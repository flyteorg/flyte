import * as React from 'react';
import makeStyles from '@material-ui/core/styles/makeStyles';
import { COLOR_SPECTRUM } from 'components/Theme/colorSpectrum';

interface StyleProps {
  chartWidth: number;
  labelInterval: number;
}

const useStyles = makeStyles((_theme) => ({
  chartHeader: (props: StyleProps) => ({
    height: 41,
    display: 'flex',
    alignItems: 'center',
    width: `${props.chartWidth}px`,
  }),
  taskDurationsLabelItem: (props: StyleProps) => ({
    fontSize: 12,
    fontFamily: 'Open Sans',
    fontWeight: 'bold',
    color: COLOR_SPECTRUM.gray40.color,
    paddingLeft: 10,
    width: `${props.labelInterval}px`,
  }),
}));

interface HeaderProps extends StyleProps {
  labels: string[];
}

export const ChartHeader = (props: HeaderProps) => {
  const styles = useStyles(props);

  return (
    <div className={styles.chartHeader}>
      {props.labels.map((label, idx) => {
        return (
          <div className={styles.taskDurationsLabelItem} key={`duration-tick-${label}-${idx}`}>
            {label}
          </div>
        );
      })}
    </div>
  );
};
