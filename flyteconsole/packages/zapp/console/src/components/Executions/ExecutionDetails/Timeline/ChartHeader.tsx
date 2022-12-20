import * as React from 'react';
import * as moment from 'moment-timezone';
import makeStyles from '@material-ui/core/styles/makeStyles';
import { COLOR_SPECTRUM } from 'components/Theme/colorSpectrum';
import { useScaleContext } from './scaleContext';
import { TimeZone } from './helpers';

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
  chartTimezone: string;
  totalDurationSec: number;
  startedAt: Date;
}

export const ChartHeader = (props: HeaderProps) => {
  const styles = useStyles(props);

  const { chartInterval: chartTimeInterval, setMaxValue } = useScaleContext();
  const { startedAt, chartTimezone, totalDurationSec } = props;

  React.useEffect(() => {
    setMaxValue(props.totalDurationSec);
  }, [props.totalDurationSec, setMaxValue]);

  const labels = React.useMemo(() => {
    const len = Math.ceil(totalDurationSec / chartTimeInterval);
    const lbs = len > 0 ? new Array(len).fill('') : [];
    return lbs.map((_, idx) => {
      const time = moment.utc(new Date(startedAt.getTime() + idx * chartTimeInterval * 1000));
      return chartTimezone === TimeZone.UTC
        ? time.format('hh:mm:ss A')
        : time.local().format('hh:mm:ss A');
    });
  }, [chartTimezone, startedAt, chartTimeInterval, totalDurationSec]);

  return (
    <div className={styles.chartHeader}>
      {labels.map((label) => {
        return (
          <div className={styles.taskDurationsLabelItem} key={label}>
            {label}
          </div>
        );
      })}
    </div>
  );
};
