import { Typography } from '@material-ui/core';
import { makeStyles, Theme } from '@material-ui/core/styles';
import * as classnames from 'classnames';
import { WaitForData } from 'components';
import { useCommonStyles } from 'components/common/styles';
import { useWorkflowNamedEntity } from 'components/hooks/useNamedEntity';
import { NamedEntityIdentifier, NamedEntityMetadata } from 'models';
import * as React from 'react';
import reactLoadingSkeleton from 'react-loading-skeleton';

const Skeleton = reactLoadingSkeleton;

const useStyles = makeStyles((theme: Theme) => ({
    description: {
        marginTop: theme.spacing(1)
    }
}));

const noDescriptionString = 'This workflow has no description.';

/** Fetches and renders the description for a given workflow ID */
export const WorkflowDescription: React.FC<{
    workflowId: NamedEntityIdentifier;
}> = ({ workflowId }) => {
    const commonStyles = useCommonStyles();
    const styles = useStyles();
    const namedEntity = useWorkflowNamedEntity(workflowId);
    const { metadata = {} as NamedEntityMetadata } = namedEntity.value;
    const hasDescription = !!metadata.description;
    return (
        <>
            <Typography variant="h6">Description</Typography>

            <Typography variant="body2" className={styles.description}>
                <WaitForData
                    {...namedEntity}
                    spinnerVariant="none"
                    loadingComponent={Skeleton}
                >
                    <span
                        className={classnames({
                            [commonStyles.hintText]: !hasDescription
                        })}
                    >
                        {hasDescription
                            ? metadata.description
                            : noDescriptionString}
                    </span>
                </WaitForData>
            </Typography>
        </>
    );
};
