import { Button, Dialog, Link, Typography } from '@material-ui/core';
import { makeStyles, Theme } from '@material-ui/core/styles';
import ArrowBack from '@material-ui/icons/ArrowBack';
import * as classnames from 'classnames';
import { navbarGridHeight } from 'common/layout';
import { useCommonStyles } from 'components/common/styles';
import { useLocationState } from 'components/hooks/useLocationState';
import { NavBarContent } from 'components/Navigation/NavBarContent';
import { interactiveTextDisabledColor, smallFontSize } from 'components/Theme';
import { Execution } from 'models';
import * as React from 'react';
import { Link as RouterLink } from 'react-router-dom';
import { Routes } from 'routes';
import { ExecutionInputsOutputsModal } from '../ExecutionInputsOutputsModal';
import { ExecutionStatusBadge } from '../ExecutionStatusBadge';
import { TerminateExecutionButton } from '../TerminateExecution';
import { executionIsTerminal } from '../utils';
import { RelaunchExecutionForm } from './RelaunchExecutionForm';

const useStyles = makeStyles((theme: Theme) => {
    return {
        actions: {
            alignItems: 'center',
            display: 'flex',
            justifyContent: 'flex-end',
            flex: '1 0 auto',
            height: '100%',
            marginLeft: theme.spacing(2)
        },
        backLink: {
            color: 'inherit',
            marginRight: theme.spacing(1)
        },
        container: {
            alignItems: 'center',
            display: 'flex',
            flex: '1 1 auto',
            maxWidth: '100%'
        },
        titleContainer: {
            alignItems: 'center',
            display: 'flex',
            flex: '0 1 auto',
            flexDirection: 'column',
            maxHeight: theme.spacing(navbarGridHeight),
            overflow: 'hidden'
        },
        inputsOutputsLink: {
            color: interactiveTextDisabledColor
        },
        actionButton: {
            marginLeft: theme.spacing(2)
        },
        title: {
            flex: '0 1 auto',
            marginLeft: theme.spacing(2)
        },
        version: {
            flex: '0 1 auto',
            overflow: 'hidden'
        }
    };
});

/** Renders information about a given Execution into the NavBar */
export const ExecutionDetailsAppBarContent: React.FC<{
    execution: Execution;
}> = ({ execution }) => {
    const commonStyles = useCommonStyles();
    const styles = useStyles();
    const [showInputsOutputs, setShowInputsOutputs] = React.useState(false);
    const [showRelaunchForm, setShowRelaunchForm] = React.useState(false);

    const { domain, name, project } = execution.id;
    const { phase, workflowId } = execution.closure;

    const {
        backLink = Routes.WorkflowDetails.makeUrl(
            workflowId.project,
            workflowId.domain,
            workflowId.name
        )
    } = useLocationState();

    let modalContent: JSX.Element | null = null;
    if (showInputsOutputs) {
        const onClose = () => setShowInputsOutputs(false);
        modalContent = (
            <ExecutionInputsOutputsModal
                execution={execution}
                onClose={onClose}
            />
        );
    }
    const onClickShowInputsOutputs = () => setShowInputsOutputs(true);
    const onClickRelaunch = () => setShowRelaunchForm(true);
    const onCloseRelaunch = () => setShowRelaunchForm(false);

    const actionContent = executionIsTerminal(execution) ? (
        <Button
            variant="outlined"
            color="primary"
            className={classnames(
                styles.actionButton,
                commonStyles.buttonWhiteOutlined
            )}
            onClick={onClickRelaunch}
            size="small"
        >
            Relaunch
        </Button>
    ) : (
        <TerminateExecutionButton className={styles.actionButton} />
    );

    return (
        <>
            <NavBarContent>
                <div className={styles.container}>
                    <RouterLink className={styles.backLink} to={backLink}>
                        <ArrowBack />
                    </RouterLink>
                    <ExecutionStatusBadge phase={phase} type="workflow" />
                    <div className={styles.titleContainer}>
                        <Typography
                            variant="body1"
                            className={classnames(
                                styles.title,
                                commonStyles.textWrapped
                            )}
                        >
                            <span>
                                {`${project}/${domain}/${workflowId.name}/`}
                                <strong>{`${name}`}</strong>
                            </span>
                        </Typography>
                    </div>
                    <div className={styles.actions}>
                        <Link
                            className={styles.inputsOutputsLink}
                            component="button"
                            onClick={onClickShowInputsOutputs}
                            variant="body1"
                        >
                            View Inputs &amp; Outputs
                        </Link>
                        {actionContent}
                    </div>
                </div>
                <Dialog
                    scroll="paper"
                    maxWidth="sm"
                    fullWidth={true}
                    open={showRelaunchForm}
                >
                    <RelaunchExecutionForm
                        execution={execution}
                        onClose={onCloseRelaunch}
                    />
                </Dialog>
            </NavBarContent>
            {modalContent}
        </>
    );
};
