export interface Item {
    id: string;
}

export const generateItems: (c: number) => Item[] = (length: number) =>
    Array.from({ length }, (_, idx) => ({ id: `${idx}` }));
