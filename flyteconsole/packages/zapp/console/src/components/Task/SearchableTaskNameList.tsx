import { Typography, IconButton, Button, CircularProgress } from '@material-ui/core';
import { makeStyles, Theme } from '@material-ui/core/styles';
import ErrorOutline from '@material-ui/icons/ErrorOutline';
import classnames from 'classnames';
import { SearchResult } from 'components/common/SearchableList';
import {
  SearchableNamedEntityListProps,
  useNamedEntityListStyles,
} from 'components/common/SearchableNamedEntityList';
import { useCommonStyles } from 'components/common/styles';
import { WaitForData } from 'components/common/WaitForData';
import { NamedEntity } from 'models/Common/types';
import * as React from 'react';
import { useState } from 'react';
import { IntersectionOptions, useInView } from 'react-intersection-observer';
import reactLoadingSkeleton from 'react-loading-skeleton';
import { Link } from 'react-router-dom';
import { Routes } from 'routes/routes';
import UnarchiveOutline from '@material-ui/icons/UnarchiveOutlined';
import ArchiveOutlined from '@material-ui/icons/ArchiveOutlined';
import { useSnackbar } from 'notistack';
import { updateTaskState } from 'models/Task/api';
import { useMutation } from 'react-query';
import { FilterableNamedEntityList } from 'components/common/FilterableNamedEntityList';
import { NamedEntityState } from 'models/enums';
import { useLatestTaskVersion } from './useLatestTask';
import t from '../Executions/Tables/WorkflowExecutionTable/strings';
import { SimpleTaskInterface } from './SimpleTaskInterface';
import { isTaskArchived } from './utils';

const Skeleton = reactLoadingSkeleton;

export const showOnHoverClass = 'showOnHover';

const useStyles = makeStyles((theme: Theme) => ({
  container: {
    // All children using the showOnHover class will be hidden until
    // the mouse enters the container
    [`& .${showOnHoverClass}`]: {
      opacity: 0,
    },
    [`&:hover .${showOnHoverClass}`]: {
      opacity: 1,
    },
  },
  actionContainer: {
    display: 'flex',
    right: 0,
    top: 0,
    position: 'absolute',
    height: '100%',
  },
  archiveCheckbox: {
    whiteSpace: 'nowrap',
  },
  centeredChild: {
    alignItems: 'center',
    marginRight: 24,
  },
  confirmationButton: {
    borderRadius: 0,
    minWidth: '100px',
    minHeight: '53px',
  },
  description: {
    color: theme.palette.text.secondary,
    marginTop: theme.spacing(0.5),
    marginBottom: theme.spacing(0.5),
  },
  errorContainer: {
    // Fix icon left alignment
    marginLeft: '-2px',
  },
  interfaceContainer: {
    width: '100%',
  },
  taskName: {
    fontWeight: 'bold',
  },
}));

interface TaskNameRowProps {
  label: React.ReactNode;
  entityName: NamedEntity;
}

interface SearchableTaskNameListProps {
  names: NamedEntity[];
  onArchiveFilterChange: (showArchievedItems: boolean) => void;
  showArchived: boolean;
}

const intersectionOptions: IntersectionOptions = {
  rootMargin: '100px 0px',
  triggerOnce: true,
};

const TaskInterfaceError: React.FC = () => {
  const { flexCenter, hintText, iconRight } = useCommonStyles();
  const { errorContainer } = useStyles();
  return (
    <div className={classnames(errorContainer, flexCenter)}>
      <ErrorOutline fontSize="small" color="disabled" />
      <div className={classnames(iconRight, hintText)}>Failed to load task interface details.</div>
    </div>
  );
};

const TaskInterface: React.FC<{
  taskName: NamedEntity;
  setShowItem: (visible: boolean) => void;
}> = ({ taskName }) => {
  const styles = useStyles();
  const task = useLatestTaskVersion(taskName.id);
  return (
    <div className={styles.interfaceContainer}>
      <WaitForData {...task} errorComponent={TaskInterfaceError} loadingComponent={Skeleton}>
        {() => <SimpleTaskInterface task={task.value} />}
      </WaitForData>
    </div>
  );
};

const getArchiveIcon = (isArchived: boolean) =>
  isArchived ? <UnarchiveOutline /> : <ArchiveOutlined />;

interface SimpleTaskActionsProps {
  item: NamedEntity;
  setShowItem: (visible: boolean) => void;
}

const SimpleTaskActions: React.FC<SimpleTaskActionsProps> = ({ item, setShowItem }) => {
  const styles = useStyles();
  const { enqueueSnackbar } = useSnackbar();
  const { id } = item;
  const isArchived = isTaskArchived(item);
  const [isUpdating, setIsUpdating] = useState<boolean>(false);
  const [showConfirmation, setShowConfirmation] = useState<boolean>(false);

  const mutation = useMutation((newState: NamedEntityState) => updateTaskState(id, newState), {
    onMutate: () => setIsUpdating(true),
    onSuccess: () => {
      enqueueSnackbar(t('archiveSuccess', !isArchived), {
        variant: 'success',
      });
      setShowItem(false);
    },
    onError: () => {
      enqueueSnackbar(`${mutation.error ?? t('archiveError', !isArchived)}`, {
        variant: 'error',
      });
    },
    onSettled: () => {
      setShowConfirmation(false);
      setIsUpdating(false);
    },
  });

  const onArchiveClick = (event: React.MouseEvent) => {
    event.stopPropagation();
    event.preventDefault();
    setShowConfirmation(true);
  };

  const onConfirmArchiveClick = (event: React.MouseEvent) => {
    event.stopPropagation();
    event.preventDefault();
    mutation.mutate(
      isTaskArchived(item)
        ? NamedEntityState.NAMED_ENTITY_ACTIVE
        : NamedEntityState.NAMED_ENTITY_ARCHIVED,
    );
  };

  const onCancelClick = (event: React.MouseEvent) => {
    event.stopPropagation();
    event.preventDefault();
    setShowConfirmation(false);
  };

  const singleItemStyle = isUpdating || !showConfirmation ? styles.centeredChild : '';

  return (
    <div className={classnames(styles.actionContainer, showOnHoverClass, singleItemStyle)}>
      {isUpdating ? (
        <IconButton size="small">
          <CircularProgress size={24} />
        </IconButton>
      ) : showConfirmation ? (
        <>
          <Button
            size="medium"
            variant="contained"
            color="primary"
            className={styles.confirmationButton}
            disableElevation
            onClick={onConfirmArchiveClick}
          >
            {t('archiveAction', isArchived)}
          </Button>
          <Button
            size="medium"
            variant="contained"
            color="inherit"
            className={styles.confirmationButton}
            disableElevation
            onClick={onCancelClick}
          >
            {t('cancelAction')}
          </Button>
        </>
      ) : (
        <IconButton size="small" title={t('archiveAction', isArchived)} onClick={onArchiveClick}>
          {getArchiveIcon(isArchived)}
        </IconButton>
      )}
    </div>
  );
};

const TaskNameRow: React.FC<TaskNameRowProps> = ({ label, entityName }) => {
  const styles = useStyles();
  const listStyles = useNamedEntityListStyles();
  const [inViewRef, inView] = useInView(intersectionOptions);
  const description = entityName?.metadata?.description;
  const [showItem, setShowItem] = useState(true);

  if (!showItem) {
    return null;
  }

  return (
    <div ref={inViewRef} className={listStyles.searchResult}>
      <div className={listStyles.itemName}>
        <div className={styles.taskName}>{label}</div>
        {description && (
          <Typography variant="body2" className={styles.description}>
            {description}
          </Typography>
        )}
        {!!inView && <TaskInterface setShowItem={setShowItem} taskName={entityName} />}
        <SimpleTaskActions item={entityName} setShowItem={setShowItem} />
      </div>
    </div>
  );
};

/** Renders a searchable list of Task names, with associated metadata */
export const SearchableTaskNameList: React.FC<
  Omit<SearchableNamedEntityListProps, 'renderItem'> & SearchableTaskNameListProps
> = (props) => {
  const commonStyles = useCommonStyles();
  const styles = useStyles();
  const renderItem = ({ key, value, content }: SearchResult<NamedEntity>) => (
    <Link
      key={key}
      className={classnames(commonStyles.linkUnstyled, styles.container)}
      to={Routes.TaskDetails.makeUrl(value.id.project, value.id.domain, value.id.name)}
    >
      <TaskNameRow label={content} entityName={value} />
    </Link>
  );
  return (
    <FilterableNamedEntityList
      placeholder="Search Task Name"
      archiveCheckboxLabel="Show Only Archived Tasks"
      {...props}
      renderItem={renderItem}
    />
  );
};
