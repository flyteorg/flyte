export function resolveAfter<T>(waitMs: number, value: T): Promise<T> {
  return new Promise((resolve) => {
    setTimeout(() => resolve(value), waitMs);
  });
}

export function rejectAfter<T>(waitMs: number, reason: string): Promise<T> {
  return new Promise((_resolve, reject) => {
    setTimeout(() => reject(reason), waitMs);
  });
}
