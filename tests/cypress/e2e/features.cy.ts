describe('OGC API Features tests', () => {

  // Fix for https://github.com/cypress-io/cypress/issues/1502#issuecomment-832403402
  Cypress.on("window:before:load", () => {
    cy.state("jQuery", Cypress.$);
  });

  it('features page should have no a11y violations', () => {
    cy.visit('/collections/addresses/items')
    cy.injectAxe()
    cy.checkA11y()
  })

  it("features page should have valid HTML", () => {
    cy.visit("/collections/addresses/items");
    cy.htmlvalidate({
      exclude: ["#featuremap"], // exclude viewer
    })
  })

  it('features page should have no broken links', () => {
    cy.visit('/collections/addresses/items')
    cy.checkForBrokenLinks()
  })

  it('feature page should have no a11y violations', () => {
    cy.visit('/collections/addresses/items/4285720b-1a60-50ce-b5fd-fc1c381bda0b')
    cy.injectAxe()
    cy.checkA11y()
  })

  it("feature page should have valid HTML", () => {
    cy.visit('/collections/addresses/items/4285720b-1a60-50ce-b5fd-fc1c381bda0b');
    cy.htmlvalidate({
      exclude: ["#featuremap"], // exclude viewer
    })
  })

  it('feature page should have no broken links', () => {
    cy.visit('/collections/addresses/items/4285720b-1a60-50ce-b5fd-fc1c381bda0b')
    cy.checkForBrokenLinks()
  })

  it('schema page should have no a11y violations', () => {
    cy.visit('/collections/addresses/schema')
    cy.injectAxe()
    cy.checkA11y()
  })

  it('schema page should have valid HTML', () => {
    cy.visit('/collections/addresses/schema')
  })

  it('schema page should have no broken links', () => {
    cy.visit('/collections/addresses/schema')
    cy.checkForBrokenLinks()
  })
})