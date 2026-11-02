import eslint from '@eslint/js'
import tseslint from 'typescript-eslint'
import angular from 'angular-eslint'
import prettier from 'eslint-plugin-prettier/recommended'
import rxjs from '@smarttools/eslint-plugin-rxjs'

export default tseslint.config(
  {
    ignores: ['dist/**', '.angular/**', 'node_modules/**'],
  },
  {
    files: ['**/*.ts'],
    extends: [
      eslint.configs.recommended,
      ...tseslint.configs.recommended,
      ...angular.configs.tsRecommended,
      prettier,
      rxjs.configs.recommended,
    ],
    languageOptions: {
      parserOptions: {
        project: './tsconfig.json',
      },
    },
    processor: angular.processInlineTemplates,
    rules: {
      'no-console': ['error'],
      'prettier/prettier': 'error',
      '@angular-eslint/directive-selector': ['error', { type: 'attribute', prefix: 'app', style: 'camelCase' }],
      '@angular-eslint/component-selector': ['error', { type: 'element', prefix: 'app', style: 'kebab-case' }],
    },
  },
  {
    files: ['**/*.html'],
    extends: [...angular.configs.templateRecommended, ...angular.configs.templateAccessibility, prettier],
    rules: {
      'prettier/prettier': ['error', { parser: 'angular', singleQuote: true, semi: false }],
    },
  }
)
