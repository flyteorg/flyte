import { makeStyles, Theme, Typography } from '@material-ui/core';
import FormControlLabel from '@material-ui/core/FormControlLabel';
import Checkbox from '@material-ui/core/Checkbox';
import * as React from 'react';
import { Protobuf } from 'flyteidl';
import { useStyles } from './styles';
import { LaunchInterruptibleInputRef } from './types';
import { formStrings } from './constants';

export const useInterruptibleStyles = makeStyles((theme: Theme) => ({
  labelIndeterminate: {
    color: theme.palette.grey[500],
  },
}));

const isValueValid = (value: any) => {
  return value !== undefined && value !== null;
};

interface LaunchInterruptibleInputProps {
  initialValue?: Protobuf.IBoolValue | null;
}

export const LaunchInterruptibleInputImpl: React.ForwardRefRenderFunction<
  LaunchInterruptibleInputRef,
  LaunchInterruptibleInputProps
> = (props, ref) => {
  // interruptible stores the override to enable/disable the setting for an execution
  const [interruptible, setInterruptible] = React.useState(false);
  // indeterminate tracks whether the interruptible flag is unspecified/indeterminate (true) or an override has been selected (false)
  const [indeterminate, setIndeterminate] = React.useState(true);

  React.useEffect(() => {
    if (isValueValid(props.initialValue) && isValueValid(props.initialValue!.value)) {
      setInterruptible(() => props.initialValue!.value!);
      setIndeterminate(() => false);
    } else {
      setInterruptible(() => false);
      setIndeterminate(() => true);
    }
  }, [props.initialValue?.value]);

  const handleInputChange = React.useCallback(() => {
    if (indeterminate) {
      setInterruptible(() => true);
      setIndeterminate(() => false);
    } else if (interruptible) {
      setInterruptible(() => false);
      setIndeterminate(() => false);
    } else {
      setInterruptible(() => false);
      setIndeterminate(() => true);
    }
  }, [interruptible, indeterminate]);

  React.useImperativeHandle(
    ref,
    () => ({
      getValue: () => {
        if (indeterminate) {
          return null;
        }

        return Protobuf.BoolValue.create({ value: interruptible });
      },
      validate: () => true,
    }),
    [interruptible, indeterminate],
  );

  const styles = useStyles();
  const interruptibleStyles = useInterruptibleStyles();

  const getInterruptibleLabel = () => {
    if (indeterminate) {
      return (
        <Typography
          className={interruptibleStyles.labelIndeterminate}
        >{`${formStrings.interruptible} (no override)`}</Typography>
      );
    } else if (interruptible) {
      return <Typography>{`${formStrings.interruptible} (enabled)`}</Typography>;
    }
    return <Typography>{`${formStrings.interruptible} (disabled)`}</Typography>;
  };

  return (
    <section>
      <header className={styles.sectionHeader}>
        <Typography variant="h6">Override interruptible flag</Typography>
        <Typography variant="body2">
          Overrides the interruptible flag of a workflow for a single execution, allowing it to be
          forced on or off. If no value was selected, the workflow's default will be used.
        </Typography>
      </header>
      <section title={formStrings.interruptible}>
        <FormControlLabel
          control={
            <Checkbox
              checked={interruptible}
              indeterminate={indeterminate}
              onChange={handleInputChange}
            />
          }
          label={getInterruptibleLabel()}
        />
      </section>
    </section>
  );
};

export const LaunchInterruptibleInput = React.forwardRef(LaunchInterruptibleInputImpl);
