import { Button } from '@material-ui/core';
import Search from '@material-ui/icons/Search';
import * as React from 'react';

import { ButtonCircularProgress, NonIdealState } from 'components/common';

const title = 'No results found';
const retry = 'Retry';

export interface NoItemsContentProps {
    loading?: boolean;
    onRetry?(): void;
}

export const NoItemsContent: React.FC<NoItemsContentProps> = ({
    loading = false,
    onRetry
}) => {
    const action =
        typeof onRetry === 'function' ? (
            <Button disabled={loading} onClick={onRetry} variant="contained">
                {retry}
                {loading && <ButtonCircularProgress />}
            </Button>
        ) : null;
    return (
        <NonIdealState icon={Search} title={title}>
            {action}
        </NonIdealState>
    );
};
