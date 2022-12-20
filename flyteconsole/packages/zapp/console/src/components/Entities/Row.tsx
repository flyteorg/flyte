import * as React from 'react';
import { Typography } from '@material-ui/core';
import { makeStyles, Theme } from '@material-ui/core/styles';
import { COLOR_SPECTRUM } from 'components/Theme/colorSpectrum';

const useStyles = makeStyles((theme: Theme) => ({
  row: {
    display: 'flex',
    marginBottom: theme.spacing(1),
  },
  title: {
    width: 100,
    color: COLOR_SPECTRUM.gray25.color,
  },
}));

interface MyProps {
  children?: React.ReactNode;
  title: String;
}
export const Row: React.FC<MyProps> = (props) => {
  const styles = useStyles();

  return (
    <div className={styles.row}>
      <div className={styles.title}>
        <Typography>{props.title}</Typography>
      </div>
      <div>{props.children}</div>
    </div>
  );
};
