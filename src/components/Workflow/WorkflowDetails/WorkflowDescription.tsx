import { Typography } from '@material-ui/core';
import * as classnames from 'classnames';
import * as React from 'react';

import { makeStyles, Theme } from '@material-ui/core/styles';
import { useCommonStyles } from 'components/common/styles';

const useStyles = makeStyles((theme: Theme) => ({
    description: {
        marginTop: theme.spacing(1)
    }
}));

const noDescriptionString = 'This workflow has no description.';

export const WorkflowDescription: React.FC<{ descriptionString: string }> = ({
    descriptionString
}) => {
    const commonStyles = useCommonStyles();
    const styles = useStyles();
    const hasDescription = descriptionString.length > 0;
    return (
        <>
            <Typography variant="h6">Description</Typography>
            <Typography
                variant="body2"
                className={classnames(styles.description, {
                    [commonStyles.hintText]: !hasDescription
                })}
            >
                {hasDescription ? descriptionString : noDescriptionString}
            </Typography>
        </>
    );
};
