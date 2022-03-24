import { Button } from '@material-ui/core';
import { makeStyles, Theme } from '@material-ui/core/styles';
import { ButtonCircularProgress } from 'components/common/ButtonCircularProgress';
import { useCommonStyles } from 'components/common/styles';
import * as React from 'react';
import { loadMoreRowGridHeight } from './constants';

const useStyles = makeStyles((theme: Theme) => ({
  loadMoreRowContainer: {
    alignItems: ' center',
    display: ' flex',
    flexDirection: 'column',
    height: theme.spacing(loadMoreRowGridHeight),
    justifyContent: ' center',
    padding: `${theme.spacing(2)}px`,
    width: '100%',
  },
  button: {
    color: theme.palette.text.secondary,
  },
}));

interface LoadMoreRowContentProps {
  lastError: string | Error | null;
  isFetching: boolean;
  style?: any;
  loadMoreRows: () => void;
}

/** Handles rendering the content below a table, which can be a "Load More"
 * button, or an error
 */
export const LoadMoreRowContent: React.FC<LoadMoreRowContentProps> = (props) => {
  const commonStyles = useCommonStyles();
  const styles = useStyles();
  const { loadMoreRows, lastError, isFetching, style } = props;

  const button = (
    <Button onClick={loadMoreRows} size="small" variant="outlined" disabled={isFetching}>
      Load More
      {isFetching ? <ButtonCircularProgress /> : null}
    </Button>
  );
  const errorContent = lastError ? (
    <div className={commonStyles.errorText}>Failed to load additional items</div>
  ) : null;

  return (
    <div className={styles.loadMoreRowContainer} style={style}>
      {errorContent}
      {button}
    </div>
  );
};
