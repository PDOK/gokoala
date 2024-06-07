import { defineConfig } from 'cypress'
import vitePreprocessor from 'cypress-vite'

export default defineConfig({
  // e2e: {
  //   setupNodeEvents(on, config) {
  // implement node event listeners here
  //  },
  // },
  component: {
    devServer: {
      framework: 'angular',
      bundler: 'vite',
    },
    specPattern: '**/*.cy.ts',
  },

  e2e: {
    setupNodeEvents(on) {
      on('file:preprocessor', vitePreprocessor())
    },
  },
})
