import { env } from 'common/env';
import {
  adminApiUrl,
  createPaginationTransformer,
  decodeProtoResponse,
  encodeProtoPayload,
} from '../utils';

jest.mock('common/env', () => ({
  env: jest.requireActual('common/env').env,
}));

const adminPrefix = 'http://admin';
const corsPrefix = '/cors_proxy';

describe('AdminEntity/utils', () => {
  describe('adminApiUrl', () => {
    beforeEach(() => {
      env.ADMIN_API_URL = adminPrefix;
      env.CORS_PROXY_PREFIX = corsPrefix;
    });

    it('uses env variable if provided', () => {
      expect(adminApiUrl('/test')).toContain(adminPrefix);
    });

    it('creates a local url if no env variable is set', () => {
      env.ADMIN_API_URL = '';
      expect(adminApiUrl('/abc').indexOf(window.location.origin)).toBe(0);
    });

    it('adds leading slash if missing', () => {
      const url = adminApiUrl('test');
      expect(url.substring(url.length - 5)).toBe('/test');
    });

    it('does not add leading slash if present', () => {
      const url = adminApiUrl('test');
      expect(url.substring(url.length - 6, url.length - 4)).not.toBe('//');
    });
  });

  describe('decodeProtoResponse', () => {
    it('provides value as a Uint8Array', () => {
      const messageType = {
        decode: jest.fn(),
      };
      const buffer = new ArrayBuffer(8);
      decodeProtoResponse(buffer, messageType);
      expect(messageType.decode).toHaveBeenCalledWith(expect.any(Uint8Array));
    });
  });

  describe('encodeProtoPayload', () => {
    it('should copy contents to a new buffer', () => {
      const data = new Uint8Array(128);
      const messageType = {
        encode: jest.fn(
          () =>
            ({
              finish: jest.fn(() => data),
            } as any),
        ),
      };
      const result = encodeProtoPayload([], messageType);
      expect(result.length).toEqual(128);
      expect(result).not.toBe(data);
    });
  });

  describe('createPaginationTransformer', () => {
    it('extracts the correct property', () => {
      const data = { items: ['a', 'b'], token: 'token' };
      const transformer = createPaginationTransformer<string, typeof data>('items');

      expect(transformer(data)).toEqual(expect.objectContaining({ entities: data.items }));
    });

    it('forwards pagination token', () => {
      const data = { items: [], token: 'token' };
      const transformer = createPaginationTransformer<string, typeof data>('items');

      expect(transformer(data)).toEqual(expect.objectContaining({ token: data.token }));
    });
  });
});
