import * as React from 'react';
import { FormControlLabel, Checkbox, FormGroup } from '@material-ui/core';
import { makeStyles, Theme } from '@material-ui/core/styles';
import { MultiSelectForm } from 'components/common/MultiSelectForm';
import { SearchInputForm } from 'components/common/SearchInputForm';
import { SingleSelectForm } from 'components/common/SingleSelectForm';
import { FilterPopoverButton } from 'components/Tables/filters/FilterPopoverButton';
import {
    FilterState,
    MultiFilterState,
    SearchFilterState,
    SingleFilterState,
    BooleanFilterState
} from './filters/types';

const useStyles = makeStyles((theme: Theme) => ({
    container: {
        alignItems: 'center',
        display: 'flex',
        flexDirection: 'row',
        height: theme.spacing(7),
        minHeight: theme.spacing(7),
        paddingLeft: theme.spacing(1),
        width: '100%'
    },
    filterButton: {
        marginLeft: theme.spacing(1)
    },
    checkbox: {
        marginLeft: theme.spacing(1)
    }
}));

const RenderFilter: React.FC<{ filter: FilterState }> = ({ filter }) => {
    const searchFilterState = filter as SearchFilterState;
    switch (filter.type) {
        case 'single':
            return <SingleSelectForm {...(filter as SingleFilterState<any>)} />;
        case 'multi':
            return (
                <MultiSelectForm {...(filter as MultiFilterState<any, any>)} />
            );
        case 'search':
            return (
                <SearchInputForm
                    {...searchFilterState}
                    defaultValue={searchFilterState.value}
                />
            );
        default:
            return null;
    }
};

/** Renders the set of filter buttons relevant to a table of WorkflowExecutions:
 * Status, Version, Start Time, Duration
 * The state for this component is generated externally by `useExecutionFiltersState` and passed in.
 * This allows for the consuming code to have direct access to the
 * current filters without relying on complicated callback arrangements
 */
export const ExecutionFilters: React.FC<{
    filters: (FilterState | BooleanFilterState)[];
    chartIds?: string[];
    clearCharts?: () => void;
    showArchived?: boolean;
    onArchiveFilterChange?: (showArchievedItems: boolean) => void;
}> = ({
    filters,
    chartIds,
    clearCharts,
    showArchived,
    onArchiveFilterChange
}) => {
    const styles = useStyles();

    filters = filters.map(filter => {
        const onChangeFunc = filter.onChange;
        filter.onChange = value => {
            if (clearCharts) clearCharts();
            if (onChangeFunc) onChangeFunc(value);
        };
        return filter;
    });

    return (
        <div className={styles.container}>
            {filters.map((filter: any) => {
                if (filter.hidden) {
                    return null;
                }
                if (filter.type === 'boolean') {
                    const handleChange = (
                        event: React.ChangeEvent<HTMLInputElement>
                    ) => filter.setActive(event.target.checked);

                    return (
                        <FormControlLabel
                            key={filter.label}
                            data-testid="checkbox"
                            control={
                                <Checkbox
                                    checked={filter.active}
                                    onChange={handleChange}
                                />
                            }
                            className={styles.checkbox}
                            label={filter.label}
                        />
                    );
                }
                const renderContent = () => <RenderFilter filter={filter} />;
                return (
                    <FilterPopoverButton
                        {...filter.button}
                        active={filter.active}
                        key={filter.label}
                        onReset={filter.onReset}
                        buttonText={filter.label}
                        className={styles.filterButton}
                        renderContent={renderContent}
                    />
                );
            })}
            {chartIds && chartIds.length > 0 && (
                <FilterPopoverButton
                    open={false}
                    active={true}
                    renderContent={() => <></>}
                    className={styles.filterButton}
                    buttonText="Clear Manually Selected Executions"
                    onReset={clearCharts}
                    key="charts"
                />
            )}
            {!!onArchiveFilterChange && (
                <FormGroup>
                    <FormControlLabel
                        control={
                            <Checkbox
                                value={showArchived}
                                onChange={(_, checked) =>
                                    onArchiveFilterChange(checked)
                                }
                            />
                        }
                        className={styles.checkbox}
                        label="Show archived executions"
                    />
                </FormGroup>
            )}
        </div>
    );
};
