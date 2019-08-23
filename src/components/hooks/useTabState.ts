import { useState } from 'react';

export function useTabState(
    tabs: { [k: string]: string },
    defaultValue: string
) {
    const [value, setValue] = useState(defaultValue);
    const onChange = (event: any, tabId: string) => setValue(tabId);

    return {
        onChange,
        value
    };
}
