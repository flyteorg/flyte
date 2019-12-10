import { Button } from '@material-ui/core';
import ErrorOutline from '@material-ui/icons/ErrorOutline';
import { NonIdealState } from 'components/common';
import * as React from 'react';

import { makeStyles, Theme } from '@material-ui/core/styles';

import { NotFound } from 'components/NotFound';
import { NotAuthorizedError, NotFoundError } from 'errors';

const useStyles = makeStyles((theme: Theme) => ({
    container: {
        margin: `${theme.spacing(2)}px 0`
    }
}));

export interface DataErrorProps {
    errorTitle: string;
    error?: Error;
    retry?: () => void;
}

/** A shared error component to be used when data fails to load. */
export const DataError: React.FC<DataErrorProps> = ({
    error,
    errorTitle,
    retry
}) => {
    const styles = useStyles();
    if (error instanceof NotFoundError) {
        return <NotFound />;
    }
    // For NotAuthorized, we will be displaying a global error.
    if (error instanceof NotAuthorizedError) {
        return null;
    }

    const description = error ? error.message : undefined;

    const action = retry ? (
        <Button variant="contained" color="primary" onClick={retry}>
            Retry
        </Button>
    ) : (
        undefined
    );
    return (
        <NonIdealState
            className={styles.container}
            description={description}
            title={errorTitle}
            icon={ErrorOutline}
        >
            {action}
        </NonIdealState>
    );
};
