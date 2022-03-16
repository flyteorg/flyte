module.exports = {
  root: true,
  env: {
    browser: true,
    es2021: true,
  },
  parser: '@typescript-eslint/parser',
  parserOptions: {
    ecmaFeatures: {
      jsx: true,
    },
    ecmaVersion: 'latest',
    sourceType: 'module',
  },
  extends: ['plugin:react/recommended', 'airbnb', 'prettier'],
  plugins: ['react', '@typescript-eslint'],
  settings: {
    'import/resolver': {
      node: {
        extensions: ['.js', '.jsx', '.ts', '.tsx'],
      },
    },
  },
  rules: {
    // "off" or 0: turn off the rule completely; "warn" or 1;  "error" or 2
    'arrow-body-style': 'off',
    'import/extensions': 'off',
    'import/prefer-default-export': 'off',
    'react/jsx-boolean-value': 'off',
    'react/jsx-filename-extension': [2, { extensions: ['.jsx', '.tsx'] }],
    'lines-between-class-members': ['error', 'always', { exceptAfterSingleLine: true }],

    // up for discussion
    'react/function-component-definition': 'off',
    'react/destructuring-assignment': 'off',

    /**
     * temporarily off or warn
     * */
    // some setup of eslint or prettier needed
    'no-undef': 'off', // 1066
    'import/no-unresolved': 'off', // 1275
    'import/no-extraneous-dependencies': 'off', // 715 - !important
    'react/jsx-props-no-spreading': 'off', // 119

    // classic
    'no-unused-vars': 'off', // 364
    'no-use-before-define': 'warn', // 49
    'no-shadow': 'off', // 104
    'no-redeclare': 'warn', // 16
    'no-param-reassign': 'off', // 28
    'no-unused-expressions': 'warn', // 6
    'prefer-destructuring': 'off', // 34
    'max-classes-per-file': 'warn', // 2
    'no-empty-function': 'off',
    'no-useless-constructor': 'warn',
    'no-useless-computed-key': 'off',
    'no-restricted-syntax': 'off',
    'no-else-return': 'off',
    'no-console': 'warn',
    'no-plusplus': 'off',
    'no-var': 'off',
    'no-continue': 'off',
    'no-unsafe-optional-chaining': 'off',
    'no-throw-literal': 'off',
    'no-lonely-if': 'off',
    'no-useless-return': 'off',
    'no-return-await': 'off',
    'no-nested-ternary': 'off',
    'no-restricted-globals': 'off',
    'no-return-assign': 'off',
    'no-await-in-loop': 'off',
    'no-undef-init': 'off',
    'no-unneeded-ternary': 'off',
    'prefer-object-spread': 'off',
    'prefer-template': 'off',
    'default-case': 'off',
    'valid-typeof': 'off',
    'object-shorthand': 'off',
    'operator-assignment': 'off',
    'array-callback-return': 'off',
    'global-require': 'off',
    'dot-notation': 'off',
    'guard-for-in': 'off',
    'one-var': 'off',
    'vars-on-top': 'off',
    'consistent-return': 'off',
    'prefer-promise-reject-errors': 'off',
    'prefer-arrow-callback': 'off',
    'func-names': 'off',
    eqeqeq: 'warn',
    camelcase: 'warn',

    // import
    'import/no-dynamic-require': 'warn', // 1

    // react
    'react/button-has-type': 'off', // 5
    'react/jsx-no-useless-fragment': 'off', // 15
    'react/no-access-state-in-setstate': 'warn', // 2
    'react/jsx-no-bind': 'warn', // 3
    'react/prop-types': 'off',
    'react/jsx-curly-brace-presence': 'off',
    'react/self-closing-comp': 'off',
    'react/jsx-no-constructed-context-values': 'off',
    'react/no-unstable-nested-components': 'off',
    'react/no-unescaped-entities': 'off',
    'react/require-default-props': 'off',
    'react/no-unused-prop-types': 'off',
    'react/no-array-index-key': 'off',
    'react/no-unused-state': 'off',
    'react/static-property-placement': 'off',
    'react/state-in-constructor': 'off',
    'react/no-children-prop': 'off',
    'react/sort-comp': 'off',

    // jsx-a11y
    'jsx-a11y/aria-role': 'off',
    'jsx-a11y/anchor-is-valid': 'off',
    'jsx-a11y/click-events-have-key-events': 'off',
    'jsx-a11y/no-noninteractive-element-interactions': 'off',
    'jsx-a11y/no-static-element-interactions': 'off',
    'jsx-a11y/control-has-associated-label': 'off',
  },
  overrides: [
    {
      // overrides for test files
      files: ['*.spec.*', '*.test.*', 'src/**/test/*'],
      rules: {
        '@typescript-eslint/no-explicit-any': 'off',
        'import/no-extraneous-dependencies': 'off',
      },
    },
  ],
};
