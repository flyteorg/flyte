import { makeStyles, Theme } from '@material-ui/core';
import { headerGridHeight } from 'components/Tables/constants';
import {
  headerFontFamily,
  listhoverColor,
  nestedListColor,
  smallFontSize,
  tableHeaderColor,
  tablePlaceholderColor,
} from 'components/Theme/constants';
import { nodeExecutionsTableColumnWidths, workflowVersionsTableColumnWidths } from './constants';

export const selectedClassName = 'selected';

// NOTE: The order of these `makeStyles` calls is important, as it determines
// specificity in the browser. The execution table styles are overridden by
// the columns styles in some cases. So the column styles should be defined
// last.
export const useExecutionTableStyles = makeStyles((theme: Theme) => ({
  borderBottom: {
    borderBottom: `1px solid ${theme.palette.divider}`,
  },
  borderTop: {
    borderTop: `1px solid ${theme.palette.divider}`,
  },
  childrenContainer: {
    backgroundColor: nestedListColor,
    minHeight: theme.spacing(7),
  },
  childGroupLabel: {
    borderWidth: '2px',
    padding: `${theme.spacing(2)}px 0`,
  },
  errorContainer: {
    padding: `0 ${theme.spacing(8)}px ${theme.spacing(2)}px`,
    '$childrenContainer &': {
      paddingTop: theme.spacing(2),
      paddingLeft: theme.spacing(2),
    },
  },
  expander: {
    alignItems: 'center',
    display: 'flex',
    justifyContent: 'center',
    marginLeft: theme.spacing(-4),
    marginRight: theme.spacing(1),
    width: theme.spacing(3),
  },
  headerColumn: {
    marginRight: theme.spacing(1),
    minWidth: 0,
    '&:first-of-type': {
      marginLeft: theme.spacing(2),
    },
  },
  headerColumnVersion: {
    width: theme.spacing(4),
  },
  headerColumnName: {
    fontSize: smallFontSize,
    fontWeight: 'bold',
    textTransform: 'uppercase',
  },
  headerRow: {
    alignItems: 'center',
    borderBottom: `4px solid ${theme.palette.divider}`,
    borderTop: `1px solid ${theme.palette.divider}`,
    color: tableHeaderColor,
    display: 'flex',
    fontFamily: headerFontFamily,
    flexDirection: 'row',
    height: theme.spacing(headerGridHeight),
  },
  logLink: {
    '&:not(:first-child)': {
      borderLeft: `1px solid ${theme.palette.divider}`,
      marginLeft: theme.spacing(1),
      paddingLeft: theme.spacing(1),
    },
  },
  logLinksContainer: {
    display: 'flex',
    flexDirection: 'row',
  },
  noRowsContent: {
    color: tablePlaceholderColor,
    margin: `${theme.spacing(5)}px auto`,
    textAlign: 'center',
  },
  row: {
    display: 'flex',
    flexDirection: 'column',
    justifyContent: 'center',
    '&:hover': {
      backgroundColor: listhoverColor,
    },
    [`&.${selectedClassName}`]: {
      backgroundColor: listhoverColor,
    },
  },
  clickableRow: {
    cursor: 'pointer',
  },
  rowContent: {},
  rowColumns: {
    alignItems: 'center',
    display: 'flex',
    flexDirection: 'row',
  },
  rowColumn: {
    marginRight: theme.spacing(1),
    minWidth: 0,
    paddingBottom: theme.spacing(1),
    paddingTop: theme.spacing(1),
    textOverflow: 'ellipsis',
    whiteSpace: 'nowrap',
  },
  scrollContainer: {
    flex: '1 1 0',
    overflowY: 'scroll',
    paddingBottom: theme.spacing(3),
  },
  tableContainer: {
    display: 'flex',
    flexDirection: 'column',
  },
}));

export const nameColumnLeftMarginGridWidth = 6;
export const useColumnStyles = makeStyles((theme: Theme) => ({
  columnName: {
    flexGrow: 1,
    // We want this to fluidly expand into whatever available space,
    // so no minimum width.
    flexBasis: 0,
    overflow: 'hidden',
    '&:first-of-type': {
      marginLeft: theme.spacing(nameColumnLeftMarginGridWidth),
    },
  },
  columnNodeId: {
    flexBasis: nodeExecutionsTableColumnWidths.nodeId,
  },
  columnType: {
    flexBasis: nodeExecutionsTableColumnWidths.type,
    textTransform: 'capitalize',
  },
  columnStatus: {
    display: 'flex',
    flexBasis: nodeExecutionsTableColumnWidths.phase,
  },
  columnStartedAt: {
    flexBasis: nodeExecutionsTableColumnWidths.startedAt,
  },
  columnDuration: {
    flexBasis: nodeExecutionsTableColumnWidths.duration,
    textAlign: 'right',
  },
  columnLogs: {
    flexBasis: nodeExecutionsTableColumnWidths.logs,
    marginLeft: theme.spacing(4),
    marginRight: theme.spacing(2),
  },
  selectedExecutionName: {
    fontWeight: 'bold',
  },
}));

export const useWorkflowVersionsColumnStyles = makeStyles(() => ({
  columnRadioButton: {
    width: workflowVersionsTableColumnWidths.radio,
  },
  columnName: {
    flexBasis: workflowVersionsTableColumnWidths.name,
    whiteSpace: 'normal',
    flexGrow: 1,
  },
  columnCreatedAt: {
    flexBasis: workflowVersionsTableColumnWidths.createdAt,
  },
}));
