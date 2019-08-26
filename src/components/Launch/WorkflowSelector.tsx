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
import { FetchableData } from 'components/hooks';
import { NamedEntityIdentifier, WorkflowId } from 'models';
import * as React from 'react';

const useStyles = makeStyles((theme: Theme) => ({
    container: {
        flexGrow: 1,
        position: 'relative'
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
    searchResults: FetchableData<WorkflowSelectorOption[]>;
    searchValue?: string;
    selectedItem?: WorkflowSelectorOption;
    workflowId: NamedEntityIdentifier;
    onSearchStringChanged(newValue: string): void;
    onSelectionChanged(newSelection: WorkflowSelectorOption): void;
}

function useWorkflowSelectorState(props: WorkflowSelectorProps) {
    const {
        options,
        searchResults,
        searchValue = '',
        selectedItem,
        workflowId,
        onSearchStringChanged,
        onSelectionChanged
    } = props;
    const [isOpen, setIsOpen] = React.useState(false);
    const [items, setItems] = React.useState(options);
    const searchActive = searchValue.length > 0;
    let inputValue = '';

    if (searchActive) {
        inputValue = searchValue;
    } else if (selectedItem) {
        inputValue = selectedItem.name;
    }

    const onChange = ({
        target: { value }
    }: React.ChangeEvent<HTMLInputElement>) => onSearchStringChanged(value);

    const selectItem = (item: WorkflowSelectorOption) => {
        setIsOpen(false);
        onSelectionChanged(item);
    };

    return {
        inputValue,
        isOpen,
        items,
        onChange,
        selectItem,
        setIsOpen
    };
}

export const WorkflowSelector: React.FC<WorkflowSelectorProps> = props => {
    const styles = useStyles();
    const {
        inputValue,
        isOpen,
        items,
        onChange,
        selectItem,
        setIsOpen
    } = useWorkflowSelectorState(props);

    const handleClickShowOptions = () => {
        setIsOpen(!isOpen);
    };

    const handleMouseDownShowOptions = (
        event: React.MouseEvent<HTMLButtonElement>
    ) => {
        event.preventDefault();
    };

    return (
        <div className={styles.container}>
            <TextField
                fullWidth={true}
                InputProps={{
                    endAdornment: (
                        <InputAdornment position="end">
                            <IconButton
                                edge="end"
                                onClick={handleClickShowOptions}
                                onMouseDown={handleMouseDownShowOptions}
                                size="small"
                            >
                                {isOpen ? <ExpandLess /> : <ExpandMore />}
                            </IconButton>
                        </InputAdornment>
                    )
                }}
                label="Workflow Version"
                onChange={onChange}
                value={inputValue}
                variant="outlined"
            />
            {isOpen ? (
                <Paper className={styles.paper} elevation={1}>
                    {items.map(item => {
                        const onClick = () => selectItem(item);
                        return (
                            <MenuItem
                                onClick={onClick}
                                key={item.id}
                                // selected={isHighlighted}
                                component="div"
                                // style={{
                                //     fontWeight: isSelected ? 500 : 400
                                // }}
                            >
                                {item.name}
                            </MenuItem>
                        );
                    })}
                </Paper>
            ) : null}
        </div>
    );
};
