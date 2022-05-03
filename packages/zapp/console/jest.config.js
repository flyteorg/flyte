/** @type {import('ts-jest/dist/types').InitialOptionsTsJest} */
const sharedConfig = require('../../../script/test/jest.base.js');

module.exports = {
  ...sharedConfig,
  testPathIgnorePatterns: ['__stories__', '.storybook', 'node_modules', 'dist', 'build'],
  setupFilesAfterEnv: ['<rootDir>/src/test/setupTests.ts'],

  modulePaths: ['<rootDir>/src'],
  roots: ['<rootDir>/src'],
  transformIgnorePatterns: ['<rootDir>/node_modules/(?!@flyteorg/flyteidl)', 'protobufjs/minimal'],

  moduleNameMapper: {
    ...sharedConfig.moduleNameMapper,
    '\\.(jpg|jpeg|png|gif|eot|otf|webp|svg|ttf|woff|woff2|mp4|webm|wav|mp3|m4a|aac|oga)$':
      '<rootDir>/assetsTransformer.js',
  },

  coveragePathIgnorePatterns: [
    ...sharedConfig.coveragePathIgnorePatterns,
    '__stories__',
    'src/components/App.tsx',
    'src/tsd',
    'src/client.tsx',
    'src/protobuf.ts',
    'src/server.ts',
  ],
};
