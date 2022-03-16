import '@testing-library/jest-dom';
import { insertDefaultData } from 'mocks/insertDefaultData';
import { mockServer } from 'mocks/server';
import { obj } from './utils';

beforeAll(() => {
  insertDefaultData(mockServer);
  mockServer.listen({
    onUnhandledRequest: (req) => {
      const message = `Unexpected request: ${obj(req)}`;
      throw new Error(message);
    },
  });
});

afterEach(() => {
  mockServer.resetHandlers();
});

afterAll(() => {
  mockServer.close();
});
