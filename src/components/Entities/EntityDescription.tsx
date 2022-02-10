import { Typography } from '@material-ui/core';
import { makeStyles, Theme } from '@material-ui/core/styles';
import classnames from 'classnames';
import { useCommonStyles } from 'components/common/styles';
import { WaitForData } from 'components/common/WaitForData';
import { useNamedEntity } from 'components/hooks/useNamedEntity';
import { NamedEntityMetadata, ResourceIdentifier } from 'models/Common/types';
import * as React from 'react';
import reactLoadingSkeleton from 'react-loading-skeleton';
import { entityStrings } from './constants';
import t from './strings';

const Skeleton = reactLoadingSkeleton;

const useStyles = makeStyles((theme: Theme) => ({
    description: {
        marginTop: theme.spacing(1)
    }
}));

/** Fetches and renders the description for a given Entity (LaunchPlan,Workflow,Task) ID */
export const EntityDescription: React.FC<{
    id: ResourceIdentifier;
}> = ({ id }) => {
    const commonStyles = useCommonStyles();
    const styles = useStyles();
    const namedEntity = useNamedEntity(id);
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
                            : t(
                                  'noDescription',
                                  entityStrings[id.resourceType]
                              )}
                    </span>
                </WaitForData>
            </Typography>
        </>
    );
};
