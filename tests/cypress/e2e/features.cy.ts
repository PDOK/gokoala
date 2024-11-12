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

  it('collection page should have no broken links', () => {
    cy.visit('/collections/addresses/items')
    cy.checkForBrokenLinks()
  })

  it('feature page should have no a11y violations', () => {
    cy.visit('/collections/addresses/items/1')
    cy.injectAxe()
    cy.checkA11y()
  })

  it("feature page should have valid HTML", () => {
    cy.visit("/collections/addresses/items/1");
    cy.htmlvalidate({
      exclude: ["#featuremap"], // exclude viewer
    })
  })

  it('feature page should have no broken links', () => {
    cy.visit('/collections/addresses/items/1')
    cy.checkForBrokenLinks()
  })
})