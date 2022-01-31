import * as React from 'react';
import { createMuiTheme, MuiThemeProvider } from '@material-ui/core/styles';
import Accordion from '@material-ui/core/Accordion';
import AccordionSummary from '@material-ui/core/AccordionSummary';
import AccordionDetails from '@material-ui/core/AccordionDetails';
import Card from '@material-ui/core/Card';
import CardContent from '@material-ui/core/CardContent';
import Checkbox from '@material-ui/core/Checkbox';
import FormControlLabel from '@material-ui/core/FormControlLabel';
import TextField from '@material-ui/core/TextField';
import Typography from '@material-ui/core/Typography';
import ExpandMoreIcon from '@material-ui/icons/ExpandMore';
import Form from '@rjsf/material-ui';

import { LaunchAdvancedOptionsRef } from './types';
import { flyteidl } from '@flyteorg/flyteidl/gen/pb-js/flyteidl';
import IExecutionSpec = flyteidl.admin.IExecutionSpec;
import { State } from 'xstate';
import {
    WorkflowLaunchContext,
    WorkflowLaunchEvent,
    WorkflowLaunchTypestate
} from './launchMachine';
import { useStyles } from './styles';

const muiTheme = createMuiTheme({
    props: {
        MuiTextField: {
            variant: 'outlined'
        }
    },
    overrides: {
        MuiButton: {
            label: {
                color: 'gray'
            }
        },
        MuiCard: {
            root: {
                marginBottom: 16,
                width: '100%'
            }
        }
    },
    typography: {
        h5: {
            fontSize: '.875rem',
            fontWeight: 500
        }
    }
});

interface LaunchAdvancedOptionsProps {
    state: State<
        WorkflowLaunchContext,
        WorkflowLaunchEvent,
        any,
        WorkflowLaunchTypestate
    >;
}

const isValueValid = (value: any) => {
    return value !== undefined && value !== null;
};

export const LaunchFormAdvancedInputs = React.forwardRef<
    LaunchAdvancedOptionsRef,
    LaunchAdvancedOptionsProps
>(
    (
        {
            state: {
                context: { launchPlan, ...other }
            }
        },
        ref
    ) => {
        const styles = useStyles();
        const [labelsParamData, setLabelsParamData] = React.useState({});
        const [annotationsParamData, setAnnotationsParamData] = React.useState(
            {}
        );
        const [disableAll, setDisableAll] = React.useState(false);
        const [maxParallelism, setMaxParallelism] = React.useState('');
        const [rawOutputDataConfig, setRawOutputDataConfig] = React.useState(
            ''
        );

        React.useEffect(() => {
            if (isValueValid(other.disableAll)) {
                setDisableAll(other.disableAll!);
            }
            if (isValueValid(other.maxParallelism)) {
                setMaxParallelism(`${other.maxParallelism}`);
            }
            if (
                launchPlan?.spec.rawOutputDataConfig?.outputLocationPrefix !==
                    undefined &&
                launchPlan?.spec.rawOutputDataConfig.outputLocationPrefix !==
                    null
            ) {
                setRawOutputDataConfig(
                    launchPlan.spec.rawOutputDataConfig.outputLocationPrefix
                );
            }
            const newLabels = {
                ...(other.labels?.values || {}),
                ...(launchPlan?.spec?.labels?.values || {})
            };
            const newAnnotations = {
                ...(other.annotations?.values || {}),
                ...(launchPlan?.spec?.annotations?.values || {})
            };
            setLabelsParamData(newLabels);
            setAnnotationsParamData(newAnnotations);
        }, [
            other.disableAll,
            other.maxParallelism,
            other.labels,
            other.annotations,
            launchPlan?.spec
        ]);

        React.useImperativeHandle(
            ref,
            () => ({
                getValues: () => {
                    return {
                        disableAll,
                        maxParallelism: parseInt(maxParallelism || '', 10),
                        labels: {
                            values: labelsParamData
                        },
                        annotations: {
                            values: annotationsParamData
                        }
                    } as IExecutionSpec;
                },
                validate: () => {
                    return true;
                }
            }),
            [disableAll, maxParallelism, labelsParamData, annotationsParamData]
        );

        const handleDisableAllChange = React.useCallback(() => {
            setDisableAll(prevState => !prevState);
        }, []);

        const handleMaxParallelismChange = React.useCallback(
            ({ target: { value } }) => {
                setMaxParallelism(value);
            },
            []
        );

        const handleLabelsChange = React.useCallback(({ formData }) => {
            setLabelsParamData(formData);
        }, []);

        const handleAnnotationsParamData = React.useCallback(({ formData }) => {
            setAnnotationsParamData(formData);
        }, []);

        const handleRawOutputDataConfigChange = React.useCallback(
            ({ target: { value } }) => {
                setRawOutputDataConfig(value);
            },
            []
        );

        return (
            <>
                <section title="Labels" className={styles.collapsibleSection}>
                    <Accordion>
                        <AccordionSummary
                            expandIcon={<ExpandMoreIcon />}
                            aria-controls="Labels"
                            id="labels-form"
                        >
                            <header className={styles.sectionHeader}>
                                <Typography variant="h6">Labels</Typography>
                            </header>
                        </AccordionSummary>

                        <AccordionDetails>
                            <MuiThemeProvider theme={muiTheme}>
                                <Card variant="outlined">
                                    <CardContent>
                                        <Form
                                            schema={{
                                                type: 'object',
                                                additionalProperties: true
                                            }}
                                            formData={labelsParamData}
                                            onChange={handleLabelsChange}
                                        >
                                            <div />
                                        </Form>
                                    </CardContent>
                                </Card>
                            </MuiThemeProvider>
                        </AccordionDetails>
                    </Accordion>
                </section>
                <section
                    title="Annotations"
                    className={styles.collapsibleSection}
                >
                    <Accordion>
                        <AccordionSummary
                            expandIcon={<ExpandMoreIcon />}
                            aria-controls="Annotations"
                            id="annotations-form"
                        >
                            <header className={styles.sectionHeader}>
                                <Typography variant="h6">
                                    Annotations
                                </Typography>
                            </header>
                        </AccordionSummary>

                        <AccordionDetails>
                            <MuiThemeProvider theme={muiTheme}>
                                <Card variant="outlined">
                                    <CardContent>
                                        <Form
                                            schema={{
                                                type: 'object',
                                                additionalProperties: true
                                            }}
                                            formData={annotationsParamData}
                                            onChange={
                                                handleAnnotationsParamData
                                            }
                                        >
                                            <div />
                                        </Form>
                                    </CardContent>
                                </Card>
                            </MuiThemeProvider>
                        </AccordionDetails>
                    </Accordion>
                </section>
                <section title="Enable/Disable all notifications">
                    <FormControlLabel
                        control={
                            <Checkbox
                                checked={disableAll}
                                onChange={handleDisableAllChange}
                            />
                        }
                        label="Disable all notifications"
                    />
                </section>
                <section title="Raw output data config">
                    <TextField
                        variant="outlined"
                        label="Raw output data config"
                        fullWidth
                        margin="normal"
                        value={rawOutputDataConfig}
                        onChange={handleRawOutputDataConfigChange}
                    />
                </section>
                <section title="Max parallelism">
                    <TextField
                        variant="outlined"
                        label="Max parallelism"
                        fullWidth
                        margin="normal"
                        value={maxParallelism}
                        onChange={handleMaxParallelismChange}
                    />
                </section>
            </>
        );
    }
);
