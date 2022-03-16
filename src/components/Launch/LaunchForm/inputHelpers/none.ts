import { literalNone } from './constants';
import { InputHelper } from './types';

export const noneHelper: InputHelper = {
  fromLiteral: () => undefined,
  toLiteral: literalNone,
  validate: () => {},
};
