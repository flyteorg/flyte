import { Typography } from '@material-ui/core';
import ChevronRight from '@material-ui/icons/ChevronRight';
import { SearchResult } from 'components/common';
import {
    SearchableNamedEntity,
    SearchableNamedEntityList,
    SearchableNamedEntityListProps,
    useNamedEntityListStyles
} from 'components/common/SearchableNamedEntityList';
import { useCommonStyles } from 'components/common/styles';
import * as React from 'react';

/** Renders a searchable list of Task names, with associated metadata */
export const SearchableTaskNameList: React.FC<
    Omit<SearchableNamedEntityListProps, 'renderItem'>
> = props => {
    const commonStyles = useCommonStyles();
    const listStyles = useNamedEntityListStyles();

    const renderItem = ({
        key,
        value,
        content
    }: SearchResult<SearchableNamedEntity>) => (
        <li key={key}>
            <div className={listStyles.searchResult}>
                <div className={listStyles.itemName}>
                    <div>{content}</div>
                    {!!value.metadata.description && (
                        <Typography
                            variant="body2"
                            className={commonStyles.hintText}
                        >
                            {value.metadata.description}
                        </Typography>
                    )}
                </div>
                {/* <ChevronRight className={listStyles.itemChevron} /> */}
            </div>
        </li>
    );
    return <SearchableNamedEntityList {...props} renderItem={renderItem} />;
};
