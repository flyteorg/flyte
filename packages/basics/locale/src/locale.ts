export const createLocalizedString =
  (strings: any = {}) =>
  (key: string, ...rest: unknown[]) => {
    const value = strings[key];
    return typeof value === 'function' ? value(...rest) : value;
  };

export const patternKey = (parent: string, pattern?: string) => {
  return `${parent}_${pattern ?? ''}`;
};
