describe('OGC API Common tests', () => {

  // Fix for https://github.com/cypress-io/cypress/issues/1502#issuecomment-832403402
  Cypress.on("window:before:load", () => {
    cy.state("jQuery", Cypress.$);
  });

  it('landing page should have no a11y violations', () => {
    cy.visit('/')
    cy.injectAxe()
    cy.checkA11y()
  })

  it("landing page should have valid HTML", () => {
    cy.visit("/");
    cy.htmlvalidate();
  })

  it('landing page should have headings for all expected children (openapi, conformance, etc)', () => {
    cy.visit('/')
    cy.get('.card-header.h5').should("have.length", 6)
  })

  it('landing page should have no broken links', () => {
    cy.visit('/')
    cy.checkForBrokenLinks()
  })

  // disabled since it has two violations in the 3rd party swagger-ui component (outside our control)
  it.skip('openapi page should have no a11y violations', () => {
    cy.visit('/api')
    cy.injectAxe()
    cy.checkA11y()
  })

  it("openapi page should have valid HTML", () => {
    cy.visit("/api");
    cy.htmlvalidate();
  })

  it('openapi page should have no broken links', () => {
    cy.visit('/api')
    cy.checkForBrokenLinks()
  })

  it('conformance page should have no a11y violations', () => {
    cy.visit('/conformance')
    cy.injectAxe()
    cy.checkA11y()
  })

  it("conformance page should have valid HTML", () => {
    cy.visit("/conformance");
    cy.htmlvalidate();
  })

  // Here we also check ogc.org pages, so this test may fail if ogc webpage is down...
  it('conformance page should have no broken links', () => {
    cy.visit('/conformance')
    cy.checkForBrokenLinks()
  })

  it('collections page should have no a11y violations', () => {
    cy.visit('/collections')
    cy.injectAxe()
    cy.checkA11y()
  })

  it("collections page should have valid HTML", () => {
    cy.visit("/collections");
    cy.htmlvalidate();
  })

  it('collections page should have no broken links', () => {
    cy.visit('/collections')
    cy.checkForBrokenLinks()
  })

  it('collection page should have no a11y violations', () => {
    cy.visit('/collections/addresses')
    cy.injectAxe()
    cy.checkA11y()
  })

  it("collection page should have valid HTML", () => {
    cy.visit("/collections/addresses");
    cy.htmlvalidate();
  })

  it('collection page should have no broken links', () => {
    cy.visit('/collections/addresses')
    cy.checkForBrokenLinks()
  })
})