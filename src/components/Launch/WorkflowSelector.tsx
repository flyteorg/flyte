import {
    IconButton,
    InputAdornment,
    MenuItem,
    Paper,
    TextField
} from '@material-ui/core';
import { makeStyles, Theme } from '@material-ui/core/styles';
import ExpandLess from '@material-ui/icons/ExpandLess';
import ExpandMore from '@material-ui/icons/ExpandMore';
import { escapeKeyListener } from 'components/common/keyboardEvents';
import { useCommonStyles } from 'components/common/styles';
import { FetchFn, useFetchableData } from 'components/hooks';
import { useDebouncedValue } from 'components/hooks/useDebouncedValue';
import { NamedEntityIdentifier, WorkflowId } from 'models';
import * as React from 'react';

const minimumQuerySize = 3;
const searchDebounceTimeMs = 500;

const useStyles = makeStyles((theme: Theme) => ({
    container: {
        flexGrow: 1,
        position: 'relative'
    },
    menuItem: {
        display: 'flex',
        justifyContent: 'space-between'
    },
    paper: {
        border: `1px solid ${theme.palette.divider}`,
        left: 0,
        marginTop: theme.spacing(0.5),
        position: 'absolute',
        right: 0,
        zIndex: 1
    }
}));

export interface WorkflowSelectorOption {
    id: string;
    data: WorkflowId;
    name: string;
    description?: string;
}

export interface WorkflowSelectorProps {
    options: WorkflowSelectorOption[];
    selectedItem?: WorkflowSelectorOption;
    workflowId: NamedEntityIdentifier;
    fetchSearchResults: FetchFn<WorkflowSelectorOption[], string>;
    onSelectionChanged(newSelection: WorkflowSelectorOption): void;
}

function useWorkflowSelectorState({
    fetchSearchResults,
    options,
    selectedItem,
    workflowId,
    onSelectionChanged
}: WorkflowSelectorProps) {
    const [rawSearchValue, setSearchValue] = React.useState('');
    const debouncedSearchValue = useDebouncedValue(
        rawSearchValue,
        searchDebounceTimeMs
    );

    const [isExpanded, setIsExpanded] = React.useState(false);
    const [focused, setFocused] = React.useState(false);

    const searchResults = useFetchableData<WorkflowSelectorOption[], string>(
        {
            defaultValue: [],
            autoFetch: debouncedSearchValue.length > minimumQuerySize,
            debugName: 'WorkflowSelector Search',
            doFetch: fetchSearchResults
        },
        debouncedSearchValue
    );
    const items = focused ? searchResults.value : options;

    let inputValue = '';
    if (focused) {
        inputValue = rawSearchValue;
    } else if (selectedItem) {
        inputValue = selectedItem.name;
    }

    const onBlur = () => {
        setFocused(false);
    };

    const onFocus = () => {
        setFocused(true);
    };

    const onClickTextInput = () => {
        if (!focused) {
            setSearchValue('');
            setIsExpanded(false);
        }
    };

    const onChange = ({
        target: { value }
    }: React.ChangeEvent<HTMLInputElement>) => {
        setSearchValue(value);
    };

    const selectItem = (item: WorkflowSelectorOption) => {
        console.log(item.id);
        onSelectionChanged(item);
        setSearchValue('');
        setFocused(false);
        setIsExpanded(false);
    };

    const showSearchResults =
        (searchResults.hasLoaded || searchResults.loading) && focused;
    const showList = showSearchResults || isExpanded;

    return {
        inputValue,
        isExpanded,
        items,
        onBlur,
        onChange,
        onClickTextInput,
        onFocus,
        searchResults,
        selectItem,
        setIsExpanded,
        showList
    };
}

const preventBubble = (event: React.MouseEvent<any>) => {
    event.preventDefault();
};

/** Combines a dropdown selector of default options with a searchable text input
 * that will fetch results using a provided function.
 */
export const WorkflowSelector: React.FC<WorkflowSelectorProps> = props => {
    const styles = useStyles();
    const commonStyles = useCommonStyles();
    const {
        inputValue,
        isExpanded,
        items,
        onBlur,
        onChange,
        onClickTextInput,
        onFocus,
        selectItem,
        setIsExpanded,
        showList
    } = useWorkflowSelectorState(props);
    const inputRef = React.useRef<HTMLInputElement>();

    const blurInput = () => {
        if (inputRef.current) {
            inputRef.current.blur();
        }
    };

    const handleClickShowOptions = () => {
        blurInput();
        setIsExpanded(!isExpanded);
    };

    return (
        <div className={styles.container}>
            <TextField
                inputRef={inputRef}
                fullWidth={true}
                inputProps={{
                    onClick: onClickTextInput
                }}
                InputProps={{
                    onBlur,
                    onFocus,
                    onKeyDown: escapeKeyListener(blurInput),
                    endAdornment: (
                        <InputAdornment position="end">
                            <IconButton
                                edge="end"
                                onClick={handleClickShowOptions}
                                onMouseDown={preventBubble}
                                size="small"
                            >
                                {isExpanded ? <ExpandLess /> : <ExpandMore />}
                            </IconButton>
                        </InputAdornment>
                    )
                }}
                label="Workflow Version"
                onChange={onChange}
                value={inputValue}
                variant="outlined"
            />
            {showList ? (
                <Paper className={styles.paper} elevation={1}>
                    {items.map(item => {
                        const onClick = () => {
                            selectItem(item);
                            blurInput();
                        };
                        return (
                            <MenuItem
                                className={styles.menuItem}
                                onClick={onClick}
                                onMouseDown={preventBubble}
                                key={item.id}
                                // selected={isHighlighted}
                                component="div"
                                // style={{
                                //     fontWeight: isSelected ? 500 : 400
                                // }}
                            >
                                {item.name}
                                <span className={commonStyles.hintText}>
                                    {item.description}
                                </span>
                            </MenuItem>
                        );
                    })}
                </Paper>
            ) : null}
        </div>
    );
};
