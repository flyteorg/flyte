import { Button } from '@material-ui/core';
import { makeStyles, Theme } from '@material-ui/core/styles';
import { useCommonStyles } from 'components/common/styles';
import { isLoadingState } from 'components/hooks/fetchMachine';
import { FetchableState } from 'components/hooks/types';
import * as React from 'react';
import { ButtonCircularProgress } from '../common';
import { loadMoreRowGridHeight } from './constants';

export const useStyles = makeStyles((theme: Theme) => ({
    loadMoreRowContainer: {
        alignItems: ' center',
        display: ' flex',
        flexDirection: 'column',
        height: theme.spacing(loadMoreRowGridHeight),
        justifyContent: ' center',
        padding: `${theme.spacing(2)}px`,
        width: '100%'
    },
    button: {
        color: theme.palette.text.secondary
    }
}));

export interface LoadMoreRowContentProps {
    className?: string;
    lastError: string | Error | null;
    state: FetchableState<any>;
    style?: any;
    loadMoreRows: () => void;
}

/** Handles rendering the content below a table, which can be a "Load More"
 * button, or an error
 */
export const LoadMoreRowContent: React.FC<LoadMoreRowContentProps> = props => {
    const commonStyles = useCommonStyles();
    const styles = useStyles();
    const { loadMoreRows, lastError, state, style } = props;

    const button = (
        <Button onClick={loadMoreRows} size="small" variant="outlined">
            Load More
            {isLoadingState(state) ? <ButtonCircularProgress /> : null}
        </Button>
    );
    const errorContent = lastError ? (
        <div className={commonStyles.errorText}>
            Failed to load additional items
        </div>
    ) : null;

    return (
        <div className={styles.loadMoreRowContainer} style={style}>
            {errorContent}
            {button}
        </div>
    );
};
