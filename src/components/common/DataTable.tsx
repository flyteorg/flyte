import { makeStyles, Theme } from '@material-ui/core/styles';
import * as React from 'react';
import {
  Paper,
  Table,
  TableBody,
  TableCell,
  TableContainer,
  TableHead,
  TableRow,
} from '@material-ui/core';
import { COLOR_SPECTRUM } from 'components/Theme/colorSpectrum';

const useStyles = makeStyles((theme: Theme) => ({
  headerCell: {
    padding: theme.spacing(1, 0, 1, 0),
    color: COLOR_SPECTRUM.gray40.color,
  },
  cell: {
    padding: theme.spacing(1, 0, 1, 0),
    minWidth: '100px',
  },
  cellLeft: {
    padding: theme.spacing(1, 1, 1, 0),
    minWidth: '100px',
  },
}));

export interface DataTableProps {
  data: { [k: string]: string };
}

export const DataTable: React.FC<DataTableProps> = ({ data }) => {
  const styles = useStyles();

  return (
    <TableContainer component={Paper}>
      <Table>
        <TableHead>
          <TableRow>
            <TableCell className={styles.headerCell}>Key</TableCell>
            <TableCell className={styles.headerCell}>Value</TableCell>
          </TableRow>
        </TableHead>
        <TableBody>
          {Object.keys(data).map((key) => (
            <TableRow key={key}>
              <TableCell className={styles.cellLeft}>{key}</TableCell>
              <TableCell className={styles.cell}>{data[key]}</TableCell>
            </TableRow>
          ))}
        </TableBody>
      </Table>
    </TableContainer>
  );
};
