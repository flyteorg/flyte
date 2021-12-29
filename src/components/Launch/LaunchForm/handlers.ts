import * as React from 'react';
import { AuthRoleTypes, InputChangeHandler } from './types';

export function makeSwitchChangeHandler(onChange: InputChangeHandler) {
    return ({ target: { checked } }: React.ChangeEvent<HTMLInputElement>) => {
        onChange(checked);
    };
}

type StringChangeHandler = (
    value: string,
    roleType: AuthRoleTypes | null
) => void;
export function makeStringChangeHandler(
    onChange: StringChangeHandler | InputChangeHandler,
    roleType: AuthRoleTypes | null = null
) {
    return ({ target: { value } }: React.ChangeEvent<HTMLInputElement>) => {
        onChange(value, roleType);
    };
}
