module.exports = {
  globals: {
    'ts-jest': {
      babelConfig: true,
    },
  },
  verbose: false,
  moduleFileExtensions: ['ts', 'tsx', 'js', 'json', 'node'],
  modulePaths: ['<rootDir>/src'],
  roots: ['<rootDir>/src'],
  testPathIgnorePatterns: [
    '__stories__',
    '<rootDir>/.storybook',
    '<rootDir>/node_modules',
    '<rootDir>/dist',
    '<rootDir>/build',
  ],
  transform: {
    '^.+\\.(j|t)sx?$': 'ts-jest',
  },
  transformIgnorePatterns: ['<rootDir>/node_modules/(?!@flyteorg/flyteidl)'],
  moduleNameMapper: {
    '\\.(jpg|jpeg|png|gif|eot|otf|webp|svg|ttf|woff|woff2|mp4|webm|wav|mp3|m4a|aac|oga)$':
      '<rootDir>/assetsTransformer.js',
  },
  coverageDirectory: '.coverage',
  collectCoverageFrom: ['**/*.{ts,tsx}', '!**/*/*.stories.tsx'],
  coveragePathIgnorePatterns: [
    '__stories__',
    '<rootDir>/.storybook',
    '<rootDir>/node_modules',
    '<rootDir>/dist',
    '<rootDir>/build',
    '<rootDir>/src/tsd',
    '<rootDir>/.eslintrc.js',
    '\\.config.js$',
  ],
  coverageReporters: ['text', 'json', 'html'],
  clearMocks: true,
  setupFiles: ['<rootDir>/node_modules/regenerator-runtime/runtime'],
  setupFilesAfterEnv: ['<rootDir>/src/test/setupTests.ts'],
  preset: 'ts-jest',
};
