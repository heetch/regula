require('babel-register');
const path = require('path');

module.exports = {
  root: true,
  env: {
    node: true,
    mocha: true,
  },
  extends: [
    'plugin:vue/essential',
    '@vue/airbnb',
  ],
  rules: {
    'no-console': process.env.NODE_ENV === 'production' ? 'error' : 'off',
    'no-debugger': process.env.NODE_ENV === 'production' ? 'error' : 'off',
    "import/extensions": ['error', "ignorePackages"]
  },
  parserOptions: {
    parser: 'babel-eslint',
  },
  settings: {
    'import/resolver': {
      alias: [
        ['@', path.join(__dirname, '/src')],
      ]
    }
  }
};
