import { Typography } from '@material-ui/core';
import ChevronRight from '@material-ui/icons/ChevronRight';
import { SearchResult } from 'components/common';
import { useCommonStyles } from 'components/common/styles';
import {
    SearchableNamedEntity,
    SearchableNamedEntityList,
    SearchableNamedEntityListProps,
    useNamedEntityListStyles
} from 'components/Workflow/SearchableNamedEntityList';
import * as React from 'react';

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
