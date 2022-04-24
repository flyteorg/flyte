import { makeStyles, Theme } from '@material-ui/core';
import { workflowExecutionsTableColumnWidths } from '../constants';

export const useStyles = makeStyles((theme: Theme) => ({
  cellName: {
    paddingLeft: theme.spacing(1),
  },
  columnName: {
    flexGrow: 1,
    flexBasis: workflowExecutionsTableColumnWidths.name,
    whiteSpace: 'normal',
  },
  columnLastRun: {
    flexBasis: workflowExecutionsTableColumnWidths.lastRun,
  },
  columnStatus: {
    flexBasis: workflowExecutionsTableColumnWidths.phase,
  },
  columnStartedAt: {
    flexBasis: workflowExecutionsTableColumnWidths.startedAt,
  },
  columnDuration: {
    flexBasis: workflowExecutionsTableColumnWidths.duration,
    textAlign: 'right',
  },
  columnActions: {
    flexBasis: workflowExecutionsTableColumnWidths.actions,
    marginLeft: theme.spacing(2),
    marginRight: theme.spacing(2),
    textAlign: 'right',
  },
  rightMargin: {
    marginRight: theme.spacing(1),
  },
  confirmationButton: {
    borderRadius: 0,
    // make the button responsive, so the button won't overflow
    width: '50%',
    minHeight: '53px',
    // cancel margins that are coming from table row style
    marginTop: theme.spacing(-1),
    marginBottom: theme.spacing(-1),
  },
  actionContainer: {
    transition: theme.transitions.create('opacity', {
      duration: theme.transitions.duration.shorter,
      easing: theme.transitions.easing.easeInOut,
    }),
  },
  actionProgress: {
    width: '100px', // same as confirmationButton size
    textAlign: 'center',
  },
}));
