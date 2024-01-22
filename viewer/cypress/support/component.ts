// ***********************************************************
// This example support/component.ts is processed and
// loaded automatically before your test files.
//
// This is a great place to put global configuration and
// behavior that modifies Cypress.
//
// You can change the location of this file or turn off
// automatically serving support files with the
// 'supportFile' configuration option.
//
// You can read more here:
// https://on.cypress.io/configuration
// ***********************************************************

// Import commands.js using ES2015 syntax:
import './commands'

// Alternatively you can use CommonJS syntax:
// require('./commands')

import { mount } from 'cypress/angular'
import { Options as AxeOptions, configureAxe, injectAxe } from 'cypress-axe'
import * as axe from 'axe-core'

// Augment the Cypress namespace to include type definitions for
// your custom command.
// Alternatively, can be defined in cypress/support/component.d.ts
// with a <reference path="./component" /> at the top of your spec.

declare const checkA11y: (
  context?: string | Node | axe.ContextObject | undefined,
  options?: AxeOptions | undefined,
  violationCallback?: ((violations: axe.Result[]) => void) | undefined,
  skipFailures?: boolean
) => void
declare global {
  // eslint-disable-next-line @typescript-eslint/no-namespace
  namespace Cypress {
    interface Chainable {
      mount: typeof mount
      injectAxe: typeof injectAxe
      configureAxe: typeof configureAxe
      checkA11y: typeof checkA11y
    }
  }
}

Cypress.Commands.add('mount', mount)

// Example use:
// cy.mount(MyComponent)
