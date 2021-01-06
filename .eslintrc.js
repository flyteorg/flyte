module.exports = {
    root: true,
    env: {
        browser: true,
        es6: true,
        node: true
    },
    parser: '@typescript-eslint/parser',
    plugins: ['@typescript-eslint', 'react-hooks', 'jest'],
    extends: [
        'eslint:recommended',
        'plugin:@typescript-eslint/recommended',
        'plugin:react-hooks/recommended',
        'plugin:jest/recommended',
        'prettier',
        'prettier/@typescript-eslint'
    ],
    rules: {
        'no-case-declarations': 'warn',
        'jest/no-mocks-import': 'off',
        'jest/valid-title': 'off',
        '@typescript-eslint/no-var-requires': 'off',
        '@typescript-eslint/ban-types': 'warn',
        '@typescript-eslint/no-empty-function': 'warn',
        '@typescript-eslint/explicit-module-boundary-types': 'off'
    }
};
