export const createLocalizedString =
  (strings = {}) =>
  (key, ...rest) => {
    const value = strings[key];
    return typeof value === 'function' ? value(...rest) : value;
  };

export const patternKey = (parent: string, pattern: string) => {
  return `${parent}_${pattern}`;
};
