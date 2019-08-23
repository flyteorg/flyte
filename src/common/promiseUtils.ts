export function resolveAfter<T>(waitMs: number, value: T): Promise<T> {
    return new Promise(resolve => {
        setTimeout(() => resolve(value), waitMs);
    });
}

export function rejectAfter<T>(waitMs: number, reason: any): Promise<T> {
    return new Promise((resolve, reject) => {
        setTimeout(() => reject(reason), waitMs);
    });
}
