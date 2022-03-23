import * as React from 'react';
import { makeStyles, Theme } from '@material-ui/core/styles';
import { storiesOf } from '@storybook/react';
import { BarChartItem } from '../BarChart';

const useStyles = makeStyles((theme: Theme) => ({
  body: {
    display: 'flex',
    alignItems: 'stretch',
    borderLeft: '1.04174px dashed #C1C1C1',
    borderRight: '1.04174px dashed #C1C1C1',
    minHeight: theme.spacing(10.5),
  },
}));

const tooltipMock = (
  <div style={{ display: 'flex', flexDirection: 'column' }}>
    <span>
      Execution Id: <strong>ei0ud0veie</strong>
    </span>
    <span>Running time: 9s</span>
    <span>Started at: 3/22/2022 8:26:46 PM UTC</span>
  </div>
);

const stories = storiesOf('Common', module);
stories.add('BarChartItem', () => {
  const styles = useStyles();
  return (
    <div className={styles.body} style={{ width: '150px' }}>
      <BarChartItem isSelected={false} value={50} color={'green'} />
      <BarChartItem isSelected={true} value={100} color={'green'} tooltip={tooltipMock} />
      <BarChartItem isSelected={false} value={75} color={'red'} tooltip={tooltipMock} />
      <BarChartItem isSelected={false} value={12} color={'red'} tooltip={tooltipMock} />
    </div>
  );
});
