import * as React from 'react';
import { Admin } from 'flyteidl';
import { makeStyles, Theme } from '@material-ui/core/styles';
import Checkbox from '@material-ui/core/Checkbox';
import FormControl from '@material-ui/core/FormControl';
import FormControlLabel from '@material-ui/core/FormControlLabel';
import InputLabel from '@material-ui/core/InputLabel';
import MenuItem from '@material-ui/core/MenuItem';
import Select from '@material-ui/core/Select';
import Typography from '@material-ui/core/Typography';

import { qualityOfServiceTier, qualityOfServiceTierLabels } from './constants';
import { LaunchAdvancedOptionsRef } from './types';
import { flyteidl } from '@flyteorg/flyteidl/gen/pb-js/flyteidl';
import IExecutionSpec = flyteidl.admin.IExecutionSpec;
import { Grid, TextField } from '@material-ui/core';

const useStyles = makeStyles((theme: Theme) => ({
    sectionTitle: {
        marginBottom: theme.spacing(2)
    },
    sectionContainer: {
        display: 'flex',
        flexDirection: 'column'
    },
    qosContainer: {
        display: 'flex'
    },
    autoFlex: {
        flex: 1,
        display: 'flex'
    }
}));

interface LaunchAdvancedOptionsProps {
    spec: Admin.IExecutionSpec;
}

export const LaunchFormAdvancedInputs = React.forwardRef<
    LaunchAdvancedOptionsRef,
    LaunchAdvancedOptionsProps
>(({ spec }, ref) => {
    const styles = useStyles();
    const [qosTier, setQosTier] = React.useState(
        qualityOfServiceTier.UNDEFINED.toString()
    );
    const [disableAll, setDisableAll] = React.useState(false);
    const [maxParallelism, setMaxParallelism] = React.useState('');
    const [queueingBudget, setQueueingBudget] = React.useState('');

    React.useEffect(() => {
        if (spec.disableAll !== undefined && spec.disableAll !== null) {
            setDisableAll(spec.disableAll);
        }
        if (spec.maxParallelism !== undefined && spec.maxParallelism !== null) {
            setMaxParallelism(`${spec.maxParallelism}`);
        }
        if (
            spec.qualityOfService?.tier !== undefined &&
            spec.qualityOfService?.tier !== null
        ) {
            setQosTier(spec.qualityOfService.tier.toString());
        }
    }, [spec]);

    React.useImperativeHandle(
        ref,
        () => ({
            getValues: () => {
                return {
                    disableAll,
                    qualityOfService: {
                        tier: parseInt(qosTier || '0', 10)
                    },
                    maxParallelism: parseInt(maxParallelism || '', 10)
                } as IExecutionSpec;
            },
            validate: () => {
                return true;
            }
        }),
        [disableAll, qosTier, maxParallelism]
    );

    const handleQosTierChange = React.useCallback(({ target: { value } }) => {
        setQosTier(value);
    }, []);

    const handleDisableAllChange = React.useCallback(() => {
        setDisableAll(prevState => !prevState);
    }, []);

    const handleMaxParallelismChange = React.useCallback(
        ({ target: { value } }) => {
            setMaxParallelism(value);
        },
        []
    );

    const handleQueueingBudgetChange = React.useCallback(
        ({ target: { value } }) => {
            setQueueingBudget(value);
        },
        []
    );

    return (
        <>
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
            <section
                title="Quality of Service"
                className={styles.sectionContainer}
            >
                <Typography variant="h6" className={styles.sectionTitle}>
                    Quality of Service
                </Typography>
                <Grid container spacing={2}>
                    <Grid item sm={6}>
                        <FormControl variant="outlined" fullWidth>
                            <InputLabel id="quality-of-service-tier-id">
                                Quality of Service Tier
                            </InputLabel>
                            <Select
                                variant="outlined"
                                id="quality-of-service-tier-id"
                                label="Quality of Service Tier"
                                value={qosTier}
                                onChange={handleQosTierChange}
                            >
                                {Object.keys(qualityOfServiceTierLabels).map(
                                    tier => (
                                        <MenuItem
                                            key={`quality-of-service-${tier}`}
                                            value={tier}
                                        >
                                            {qualityOfServiceTierLabels[tier]}
                                        </MenuItem>
                                    )
                                )}
                            </Select>
                        </FormControl>
                    </Grid>
                    <Grid item sm={6}>
                        <TextField
                            variant="outlined"
                            label="Queueing budget"
                            fullWidth
                            value={queueingBudget}
                            onChange={handleQueueingBudgetChange}
                        />
                    </Grid>
                </Grid>
            </section>
        </>
    );
});
