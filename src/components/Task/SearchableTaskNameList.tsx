import { Typography } from '@material-ui/core';
import { makeStyles, Theme } from '@material-ui/core/styles';
import ErrorOutline from '@material-ui/icons/ErrorOutline';
import * as classnames from 'classnames';
import { noDescriptionString } from 'common/constants';
import { SearchResult, WaitForData } from 'components/common';
import {
    SearchableNamedEntity,
    SearchableNamedEntityList,
    SearchableNamedEntityListProps,
    useNamedEntityListStyles
} from 'components/common/SearchableNamedEntityList';
import { useCommonStyles } from 'components/common/styles';
import { NamedEntity } from 'models';
import * as React from 'react';
import { IntersectionOptions, useInView } from 'react-intersection-observer';
import reactLoadingSkeleton from 'react-loading-skeleton';
import { SimpleTaskInterface } from './SimpleTaskInterface';
import { useLatestTaskVersion } from './useLatestTask';
const Skeleton = reactLoadingSkeleton;

const useStyles = makeStyles((theme: Theme) => ({
    description: {
        color: theme.palette.text.secondary,
        marginTop: theme.spacing(0.5)
    },
    errorContainer: {
        // Fix icon left alignment
        marginLeft: '-2px',
        marginTop: theme.spacing(1)
    },
    interfaceContainer: {
        width: '100%'
    },
    taskName: {
        fontWeight: 'bold'
    }
}));

interface TaskNameRowProps {
    label: React.ReactNode;
    entityName: NamedEntity;
}

const intersectionOptions: IntersectionOptions = {
    rootMargin: '100px 0px',
    triggerOnce: true
};

const TaskInterfaceError: React.FC = () => {
    const { flexCenter, hintText, iconRight } = useCommonStyles();
    const { errorContainer } = useStyles();
    return (
        <div className={classnames(errorContainer, flexCenter)}>
            <ErrorOutline fontSize="small" color="disabled" />
            <div className={classnames(iconRight, hintText)}>
                Failed to load task interface details.
            </div>
        </div>
    );
};

const TaskInterface: React.FC<{ taskName: NamedEntity }> = ({ taskName }) => {
    const styles = useStyles();
    const task = useLatestTaskVersion(taskName.id);
    return (
        <div className={styles.interfaceContainer}>
            <WaitForData
                {...task}
                errorComponent={TaskInterfaceError}
                loadingComponent={Skeleton}
            >
                {() => <SimpleTaskInterface task={task.value} />}
            </WaitForData>
        </div>
    );
};

// TODO:
// * Write custom error content since it will be a small area

const TaskNameRow: React.FC<TaskNameRowProps> = ({ label, entityName }) => {
    const styles = useStyles();
    const listStyles = useNamedEntityListStyles();
    const [inViewRef, inView] = useInView(intersectionOptions);
    const description = entityName.metadata.description || noDescriptionString;

    return (
        <div ref={inViewRef} className={listStyles.searchResult}>
            <div className={listStyles.itemName}>
                <div className={styles.taskName}>{label}</div>
                <Typography variant="body2" className={styles.description}>
                    {description}
                </Typography>
                {!!inView && <TaskInterface taskName={entityName} />}
            </div>
        </div>
    );
};

/** Renders a searchable list of Task names, with associated metadata */
export const SearchableTaskNameList: React.FC<
    Omit<SearchableNamedEntityListProps, 'renderItem'>
> = props => {
    const renderItem = ({
        key,
        value,
        content
    }: SearchResult<SearchableNamedEntity>) => (
        <li key={key}>
            <TaskNameRow label={content} entityName={value} />
        </li>
    );
    return <SearchableNamedEntityList {...props} renderItem={renderItem} />;
};
