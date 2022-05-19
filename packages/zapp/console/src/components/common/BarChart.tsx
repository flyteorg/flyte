import * as React from 'react';
import { makeStyles, Theme } from '@material-ui/core/styles';
import { smallFontSize } from 'components/Theme/constants';
import { COLOR_SPECTRUM } from 'components/Theme/colorSpectrum';
import { Tooltip, Typography, Zoom } from '@material-ui/core';

const useStyles = makeStyles((theme: Theme) => ({
  chartContainer: {
    paddingLeft: theme.spacing(1),
    paddingRight: theme.spacing(3),
    paddingTop: theme.spacing(1),
    minHeight: '135px',
  },
  title: {
    marginTop: theme.spacing(2),
    paddingBottom: theme.spacing(1),
    paddingLeft: theme.spacing(1),
    borderBottom: `1px solid ${theme.palette.divider}`,
  },
  wrapper: {
    display: 'flex',
    flexDirection: 'column',
    marginBottom: theme.spacing(2.5),
  },
  header: {
    display: 'flex',
    justifyContent: 'space-between',
    marginBottom: theme.spacing(0.75),
    fontSize: smallFontSize,
    color: COLOR_SPECTRUM.gray40.color,
  },
  body: {
    display: 'flex',
    alignItems: 'stretch',
    borderLeft: '1.04174px dashed #C1C1C1',
    borderRight: '1.04174px dashed #C1C1C1',
    minHeight: theme.spacing(10.5),
  },
  item: {
    flex: 1,
    display: 'flex',
    flexDirection: 'column',
    justifyContent: 'flex-end',
    alignItems: 'center',
    '&:last-child': {
      marginRight: 0,
    },
  },
  itemBar: {
    borderRadius: 2,
    marginRight: theme.spacing(0.25),
    minHeight: theme.spacing(0.75),
    cursor: 'pointer',
    width: '80%',
    marginLeft: '10%',
  },
}));

interface BarChartData {
  value: number;
  color: string;
  metadata?: any;
  tooltip?: React.ReactChild;
}

interface BarChartItemProps extends BarChartData {
  onClick?: () => void;
  isSelected: boolean;
}

interface BarChartProps {
  title: string;
  data: BarChartData[];
  startDate?: string;
  onClickItem?: (item: any) => void;
  chartIds: string[];
}

/**
 * Display individual chart item for the BarChart component
 * @param value
 * @param color
 * @constructor
 */
export const BarChartItem: React.FC<BarChartItemProps> = ({
  value,
  color,
  isSelected,
  tooltip,
  onClick,
}) => {
  const styles = useStyles();

  const content = (
    <div
      className={styles.itemBar}
      style={{
        backgroundColor: color,
        height: `${value}%`,
        opacity: isSelected ? '100%' : '50%',
      }}
      onClick={onClick}
    />
  );

  return (
    <div className={styles.item}>
      {tooltip ? (
        <Tooltip title={<>{tooltip}</>} TransitionComponent={Zoom}>
          {content}
        </Tooltip>
      ) : (
        content
      )}
    </div>
  );
};

/**
 * Display information as bar chart with value and color
 * @param data
 * @param startDate
 * @constructor
 */
export const BarChart: React.FC<BarChartProps> = ({
  title,
  chartIds,
  data,
  startDate,
  onClickItem,
}) => {
  const styles = useStyles();

  const maxHeight = React.useMemo(() => {
    return Math.max(...data.map((x) => Math.log2(x.value)));
  }, [data]);

  const handleClickItem = React.useCallback(
    (item) => () => {
      if (onClickItem) {
        onClickItem(item);
      }
    },
    [onClickItem],
  );

  return (
    <div>
      <Typography className={styles.title} variant="h3">
        {title}
      </Typography>
      <div className={styles.chartContainer}>
        <div className={styles.wrapper}>
          <div className={styles.header}>
            <span>{startDate}</span>
            <span>Most Recent</span>
          </div>
          <div className={styles.body}>
            {data.map((item, index) => (
              <BarChartItem
                value={(Math.log2(item.value) / maxHeight) * 100}
                color={item.color}
                tooltip={item.tooltip}
                onClick={handleClickItem(item)}
                key={`bar-chart-item-${index}`}
                isSelected={
                  chartIds.length === 0
                    ? true
                    : item.metadata && chartIds.includes(item.metadata.name)
                }
              />
            ))}
          </div>
        </div>
      </div>
    </div>
  );
};
