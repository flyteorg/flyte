import * as React from 'react';
import { InputChangeHandler } from './types';

export function makeSwitchChangeHandler(onChange: InputChangeHandler) {
    return ({ target: { checked } }: React.ChangeEvent<HTMLInputElement>) => {
        onChange(checked);
    };
}

type StringChangeHandler = (value: string) => void;
export function makeStringChangeHandler(
    onChange: InputChangeHandler | StringChangeHandler
) {
    return ({ target: { value } }: React.ChangeEvent<HTMLInputElement>) => {
        onChange(value);
    };
}
