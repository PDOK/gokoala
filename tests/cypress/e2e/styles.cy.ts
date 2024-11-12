describe('OGC API Styles tests', () => {

  // Fix for https://github.com/cypress-io/cypress/issues/1502#issuecomment-832403402
  Cypress.on("window:before:load", () => {
    cy.state("jQuery", Cypress.$);
  });

  it('styles page should have no a11y violations', () => {
    cy.visit('/styles')
    cy.injectAxe()
    cy.checkA11y()
  })

  it("styles page should have valid HTML", () => {
    cy.visit("/styles");
    cy.htmlvalidate();
  })

  it('styles page should have no broken links', () => {
    cy.visit('/styles')
    cy.checkForBrokenLinks()
  })

  it('style page should have no a11y violations', () => {
    cy.visit('/styles/dummy-style')
    cy.injectAxe()
    cy.checkA11y()
  })

  it("style page should have valid HTML", () => {
    cy.visit("/styles/dummy-style");
    cy.htmlvalidate();
  })

  it('style page should have no broken links', () => {
    cy.visit('/styles/dummy-style')
    cy.checkForBrokenLinks()
  })

  it('styles metadata page should have no a11y violations', () => {
    cy.visit('/styles/dummy-style/metadata')
    cy.injectAxe()
    cy.checkA11y()
  })

  it("styles metadata page should have valid HTML", () => {
    cy.visit("/styles/dummy-style/metadata");
    cy.htmlvalidate();
  })

  it('styles metadata page should have no broken links', () => {
    cy.visit('/styles/dummy-style/metadata')
    cy.checkForBrokenLinks()
  })
})