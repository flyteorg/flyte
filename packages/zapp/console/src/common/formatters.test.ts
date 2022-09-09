import { millisecondsToHMS } from 'common/formatters';
import { unknownValueString, zeroSecondsString } from './constants';

describe('millisecondsToHMS', () => {
  it('should display unknown if the value is negative', () => {
    expect(millisecondsToHMS(-1)).toEqual(unknownValueString);
  });

  it('should display 0s if the value is 0', () => {
    expect(millisecondsToHMS(0)).toEqual(zeroSecondsString);
  });

  it('should display in `ms` format if the value is < 1000', () => {
    expect(millisecondsToHMS(999)).toEqual('999ms');
  });

  it('should display in `h` format if the value is >= 86400000 (24h)', () => {
    expect(millisecondsToHMS(86400001)).toEqual('24h');
  });

  it('should display in `h m` format if the value is < 86400000 (24h)', () => {
    expect(millisecondsToHMS(86340000)).toEqual('23h 59m');
  });

  it('should display in `h m s` format if the value is < 86400000 (24h)', () => {
    expect(millisecondsToHMS(86399999)).toEqual('23h 59m 59s');
  });

  it('should display in `m s` format if the value is < 3600000 (1h)', () => {
    expect(millisecondsToHMS(3599999)).toEqual('59m 59s');
  });

  it('should display in `m` format if the value is < 3600000 (1h) and seconds < 1', () => {
    expect(millisecondsToHMS(3540000)).toEqual('59m');
  });

  it('should display in `s` format if the value is < 60000 (1m)', () => {
    expect(millisecondsToHMS(59999)).toEqual('59s');
  });
});
