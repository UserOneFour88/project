import js from '@eslint/js';
import globals from 'globals';
import { defineConfig } from 'eslint/config';
import parserTs from '@typescript-eslint/parser';
import pluginTs from '@typescript-eslint/eslint-plugin';
import pluginReact from 'eslint-plugin-react';
import pluginReactHooks from 'eslint-plugin-react-hooks';
import pluginImport from 'eslint-plugin-import';
import pluginUnusedImports from 'eslint-plugin-unused-imports';
import pluginSort from 'eslint-plugin-simple-import-sort';
import pluginA11y from 'eslint-plugin-jsx-a11y';
import pluginStorybook from 'eslint-plugin-storybook';

export default defineConfig([
  {
    ignores: [
      'dist/**',
      'node_modules/**',
      '.husky/**',
      '.storybook/**',
      '*.config.js',
      '*.config.mjs',
      '.eslintrc.js',
      '.eslintrc.json',
      'jest.config.js',
      'webpack.config.js',
      'postcss.config.js',
      'tailwind.config.js',
    ],
  },
  js.configs.recommended,
  {
    files: ['**/*.{ts,tsx,js,jsx,mjs,cjs}'],
    languageOptions: {
      parser: parserTs,
      parserOptions: {
        ecmaVersion: 'latest',
        sourceType: 'module',
        ecmaFeatures: { jsx: true },
      },
      globals: {
        ...globals.browser,
        ...globals.node,
      },
    },
    plugins: {
      '@typescript-eslint': pluginTs,
      react: pluginReact,
      'react-hooks': pluginReactHooks,
      import: pluginImport,
      'unused-imports': pluginUnusedImports,
      'simple-import-sort': pluginSort,
      'jsx-a11y': pluginA11y,
      storybook: pluginStorybook,
    },
    settings: {
      react: {
        version: 'detect',
      },
    },
    rules: {
      ...pluginReact.configs.recommended.rules,
      ...pluginReactHooks.configs.recommended.rules,
      ...pluginA11y.configs.recommended.rules,
      ...pluginTs.configs.recommended.rules,
      'react/react-in-jsx-scope': 'off',
      'react/prop-types': 'off',
      '@typescript-eslint/no-unused-vars': 'off',
      'unused-imports/no-unused-imports': 'error',
      'unused-imports/no-unused-vars': [
        'warn',
        {
          vars: 'all',
          varsIgnorePattern: '^_',
          args: 'after-used',
          argsIgnorePattern: '^_',
        },
      ],
      'simple-import-sort/imports': 'error',
      'simple-import-sort/exports': 'error',
      'import/no-unresolved': 'off',
      'no-console': process.env.NODE_ENV === 'production' ? 'warn' : 'off',
      'no-debugger': process.env.NODE_ENV === 'production' ? 'warn' : 'off',
      'no-multiple-empty-lines': ['error', { max: 1, maxEOF: 0 }],
    },
  },
  {
    files: ['**/*.stories.{js,jsx,ts,tsx}', '.storybook/**/*.{js,jsx,ts,tsx}'],
    rules: {
      ...pluginStorybook.configs.recommended.rules,
    },
  },
]);
