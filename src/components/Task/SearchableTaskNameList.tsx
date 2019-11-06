import { Typography } from '@material-ui/core';
import { makeStyles, Theme } from '@material-ui/core/styles';
import { SearchResult, WaitForData } from 'components/common';
import {
    SearchableNamedEntity,
    SearchableNamedEntityList,
    SearchableNamedEntityListProps,
    useNamedEntityListStyles
} from 'components/common/SearchableNamedEntityList';
import { useCommonStyles } from 'components/common/styles';
import { useFetchableData } from 'components/hooks';
import { NamedEntity } from 'models';
import * as React from 'react';
import { IntersectionOptions, useInView } from 'react-intersection-observer';
import reactLoadingSkeleton from 'react-loading-skeleton';
const Skeleton = reactLoadingSkeleton;

const useStyles = makeStyles((theme: Theme) => ({
    interfaceContainer: {
        width: '100%'
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

const fakeFetchInterface = () =>
    new Promise(resolve => setTimeout(resolve, 500));

const TaskInterface: React.FC<{ taskName: NamedEntity }> = ({ taskName }) => {
    const styles = useStyles();
    const taskInterface = useFetchableData(
        {
            defaultValue: undefined,
            useCache: true,
            doFetch: fakeFetchInterface
        },
        taskName
    );
    return (
        <div className={styles.interfaceContainer}>
            <WaitForData {...taskInterface} loadingComponent={Skeleton}>
                <em>(interface data goes here)</em>
            </WaitForData>
        </div>
    );
};

// TODO:
// * Load latest task version (cache it)
// * Create component to render the interface
// * Write custom error content since it will be a small area

const TaskNameRow: React.FC<TaskNameRowProps> = ({ label, entityName }) => {
    const commonStyles = useCommonStyles();
    const listStyles = useNamedEntityListStyles();
    const [inViewRef, inView] = useInView(intersectionOptions);

    return (
        <div ref={inViewRef} className={listStyles.searchResult}>
            <div className={listStyles.itemName}>
                <div>{label}</div>
                {!!entityName.metadata.description && (
                    <Typography
                        variant="body2"
                        className={commonStyles.hintText}
                    >
                        {entityName.metadata.description}
                    </Typography>
                )}
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
