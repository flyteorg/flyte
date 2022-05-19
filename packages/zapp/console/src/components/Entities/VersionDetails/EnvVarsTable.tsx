import * as React from 'react';
import {
  Typography,
  Table,
  TableBody,
  TableCell,
  TableContainer,
  TableHead,
  TableRow,
  Paper,
} from '@material-ui/core';
import { makeStyles, Theme } from '@material-ui/core/styles';
import { Core } from 'flyteidl';
import { COLOR_SPECTRUM } from 'components/Theme/colorSpectrum';
import t from '../strings';

const useStyles = makeStyles((theme: Theme) => ({
  container: {
    marginBottom: theme.spacing(1),
    ['& .MuiTableCell-sizeSmall']: {
      paddingLeft: 0,
    },
  },
  headerText: {
    color: COLOR_SPECTRUM.gray25.color,
  },
}));

interface EnvVarsTableProps {
  rows: Core.IKeyValuePair[];
}

export default function EnvVarsTable({ rows }: EnvVarsTableProps) {
  const styles = useStyles();

  if (!rows || rows.length == 0) {
    return <Typography>{t('empty')}</Typography>;
  }
  return (
    <TableContainer className={styles.container} component={Paper}>
      <Table size="small" aria-label="a dense table">
        <TableHead>
          <TableRow>
            <TableCell>
              <Typography className={styles.headerText} variant="h4">
                {t('key')}
              </Typography>
            </TableCell>
            <TableCell>
              <Typography className={styles.headerText} variant="h4">
                {t('value')}
              </Typography>
            </TableCell>
          </TableRow>
        </TableHead>
        <TableBody>
          {rows.map((row) => (
            <TableRow key={row.key}>
              <TableCell component="th" scope="row">
                {row.key}
              </TableCell>
              <TableCell component="th" scope="row">
                {row.value}
              </TableCell>
            </TableRow>
          ))}
        </TableBody>
      </Table>
    </TableContainer>
  );
}
