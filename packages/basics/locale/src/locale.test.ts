import t, { patternKey } from './strings.mock';

describe('Basics/locale strings retirieval', () => {
  it('Simple plain string retrieved properly', () => {
    const text = t('regularString');
    expect(text).toEqual('Regular string');
  });

  it('patternKey strings retrieved properly', () => {
    let text = t(patternKey('value', 'on'));
    expect(text).toEqual('Value ON');

    text = t(patternKey('value', 'off'));
    expect(text).toEqual('Value OFF');

    text = t(patternKey('value'));
    expect(text).toEqual('Default value');
  });

  it('Localized function call, properly incorporates provided properties ', () => {
    let text = t('activeUsers', 5, 12);
    expect(text).toEqual('5 out of 12 users');

    text = t('activeUsers', 5);
    expect(text).toEqual('5 out of undefined users');
  });

  it('If string is not present - returns undefined item', () => {
    const text = t('nothingThere');
    expect(text).toEqual(undefined);
  });
});
