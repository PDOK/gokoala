import { defineConfig } from 'cypress'
import * as fixmvt from '@mapbox/mvt-fixtures'

export default defineConfig({
  // e2e: {
  //   setupNodeEvents(on, config) {
  // implement node event listeners here
  //  },
  // },
  component: {
    setupNodeEvents(on, config) {
      //  const fix1 = fixmvt.mvtf.get('043')
    },
    devServer: {
      framework: 'angular',
      bundler: 'webpack',
    },
    specPattern: '**/*.cy.ts',
  },

  e2e: {
    setupNodeEvents(on, config) {
      // implement node event listeners here
    },
  },
})
