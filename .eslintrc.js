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
        // Rules we don't want to be enabled
        'no-unused-vars': 'off', // disabled to let "@typescript-eslint/no-unused-vars" do it's job
        '@typescript-eslint/no-unused-vars': [
            'warn',
            { argsIgnorePattern: '^_' }
        ],
        'jest/no-mocks-import': 'off',

        // temporarily disabled
        '@typescript-eslint/no-var-requires': 'off', // 27
        '@typescript-eslint/ban-types': 'warn', // 62
        '@typescript-eslint/no-empty-function': 'warn', // 60
        '@typescript-eslint/explicit-module-boundary-types': 'off', // 296
        '@typescript-eslint/no-explicit-any': 'warn', // 134
        '@typescript-eslint/no-non-null-assertion': 'warn', // 59
        'react-hooks/exhaustive-deps': 'warn' // 24
    }
};
