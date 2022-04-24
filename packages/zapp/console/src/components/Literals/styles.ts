import { makeStyles, Theme } from '@material-ui/core/styles';

export const useLiteralStyles = makeStyles((theme: Theme) => ({
  nestedContainer: {
    marginLeft: theme.spacing(1),
  },
  labelValueContainer: {
    display: 'flex',
    flexDirection: 'row',
  },
  valueLabel: {
    color: theme.palette.grey[500],
    marginRight: theme.spacing(1),
  },
}));
