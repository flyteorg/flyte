import { CardContent } from '@material-ui/core';
import Card from '@material-ui/core/Card';
import { makeStyles, Theme } from '@material-ui/core/styles';
import ErrorOutline from '@material-ui/icons/ErrorOutline';
import classnames from 'classnames';
import { log } from 'common/log';
import { useCommonStyles } from 'components/common/styles';
import { NotFound } from 'components/NotFound/NotFound';
import { NotFoundError } from 'errors/fetchErrors';
import * as React from 'react';
import { NonIdealState } from './NonIdealState';

const useStyles = makeStyles((theme: Theme) => ({
    container: {
        alignItems: 'center',
        display: 'flex',
        height: '100%',
        justifyContent: 'center',
        minHeight: '100%',
        minWidth: '100%',
        width: '100%'
    },
    containerFixed: {
        bottom: 0,
        left: 0,
        position: 'fixed',
        right: 0,
        top: 0
    },
    errorContainer: {
        marginBottom: theme.spacing(2)
    },
    messageContainer: {
        maxWidth: theme.spacing(60)
    }
}));

interface ErrorBoundaryState {
    error?: Error;
    hasError: boolean;
}

const RenderError: React.FC<{ error: Error; fixed: boolean }> = ({
    error,
    fixed
}) => {
    const commonStyles = useCommonStyles();
    const styles = useStyles();
    const description = (
        <div className={styles.messageContainer}>
            <p>The error we received was:</p>
            <Card className={styles.errorContainer}>
                <CardContent className={commonStyles.codeBlock}>
                    {`${error}`}
                </CardContent>
            </Card>
            <p>There may be additional information in the browser console.</p>
        </div>
    );
    return (
        <div
            className={classnames(styles.container, {
                [styles.containerFixed]: fixed
            })}
        >
            <NonIdealState
                icon={ErrorOutline}
                title="Unexpected error"
                description={description}
            />
        </div>
    );
};

/** A generic error boundary which will render a NonIdealState and display
 * whatever error was thrown. `fixed` controls whether the container is
 * rendered with fixed positioning and filled to the edges of the window.
 */
export class ErrorBoundary extends React.Component<
    { fixed?: boolean },
    ErrorBoundaryState
> {
    constructor(props: object) {
        super(props);
        this.state = {
            error: undefined,
            hasError: false
        };
    }

    componentDidCatch(error: Error, info: unknown) {
        this.setState({ error });
        log.error(error, info);
    }

    render() {
        const { fixed = false } = this.props;
        if (this.state.error) {
            if (this.state.error instanceof NotFoundError) {
                return <NotFound />;
            }

            return <RenderError error={this.state.error} fixed={fixed} />;
        }
        return this.props.children;
    }
}
