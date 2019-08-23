import * as React from 'react';
import { NodeExecutionsTableState } from './types';

export const NodeExecutionsTableContext = React.createContext<
    NodeExecutionsTableState
>({} as NodeExecutionsTableState);
