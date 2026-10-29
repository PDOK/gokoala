/// <reference types="cypress" />
// ***********************************************
// This example commands.ts shows you how to
// create various custom commands and overwrite
// existing commands.
//
// For more comprehensive examples of custom
// commands please read more here:
// https://on.cypress.io/custom-commands
// ***********************************************
//
//
// -- This is a parent command --
// Cypress.Commands.add('login', (email, password) => { ... })
//
//
// -- This is a child command --
// Cypress.Commands.add('drag', { prevSubject: 'element'}, (subject, options) => { ... })
//
//
// -- This is a dual command --
// Cypress.Commands.add('dismiss', { prevSubject: 'optional'}, (subject, options) => { ... })
//
//
// -- This will overwrite an existing command --
// Cypress.Commands.overwrite('visit', (originalFn, url, options) => { ... })
//
// declare global {
//   namespace Cypress {
//     interface Chainable {
//       login(email: string, password: string): Chainable<void>
//       drag(subject: string, options?: Partial<TypeOptions>): Chainable<Element>
//       dismiss(subject: string, options?: Partial<TypeOptions>): Chainable<Element>
//       visit(originalFn: CommandOriginalFn, url: string, options: Partial<VisitOptions>): Chainable<Element>
//     }
//   }
// }

import type { AxeResults, Result } from 'axe-core'

declare global {
  namespace Cypress {
    interface Chainable {
      checkForBrokenLinks(): Chainable<void>
      checkForBrokenImages(): Chainable<void>
      injectAxe(): Chainable<void>
      checkA11y(): Chainable<void>
    }
  }
}

// Replaces cypress-axe, which does not support Cypress 16.
Cypress.Commands.add('injectAxe', () => {
  cy.readFile('node_modules/axe-core/axe.min.js').then(source => {
    cy.window({ log: false }).then(win => {
      win.eval(source)
    })
  })
})

Cypress.Commands.add('checkA11y', () => {
  cy.window({ log: false })
    .then(win => (win as unknown as { axe: { run: (context: Document) => Promise<AxeResults> } }).axe.run(win.document))
    .then(({ violations }: AxeResults) => {
      violations.forEach((v: Result) => {
        const targets = v.nodes.flatMap(node => node.target).join(', ')
        Cypress.log({
          $el: Cypress.$(targets),
          name: 'a11y error!',
          consoleProps: () => v,
          message: `${v.id} on ${v.nodes.length} Node${v.nodes.length === 1 ? '' : 's'}`,
        })
      })
      const summary = `${violations.length} accessibility violation${violations.length === 1 ? ' was' : 's were'} detected`
      assert.equal(violations.length, 0, summary)
    })
})

Cypress.Commands.add('checkForBrokenLinks', () =>{
  cy.get('a').each(link => {
    const href = link.prop('href')
    if (href &&
      !href.includes('example.com') &&
      !href.includes('europa.eu') &&
      !href.includes('opengis.net') &&
      !href.includes('creativecommons.org/')) {
      cy.request(href)
    }
  })
})

Cypress.Commands.add('checkForBrokenImages', () => {
  const brokenImages = [];
  const c = cy.get('img')
      .each(($el, k) => {
        if ($el.prop('naturalWidth') === 0) {
          const id = $el.attr('id')
          const alt = $el.attr('alt')
          const info = `${id ? '#' + id : ''} ${alt ? alt : ''}`
          brokenImages.push(info)
          cy.log(`Broken image ${k + 1}: ${info}`)
        }
      })
  c.then(() => {
    // report all broken images at once
    if (brokenImages.length) {
      throw new Error(
          `Found ${
              brokenImages.length
          } broken images\n${brokenImages.join(', ')}`,
      )
    }
  })

})
