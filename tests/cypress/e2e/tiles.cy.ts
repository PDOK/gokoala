describe('OGC API Tiles tests', () => {

  it('dataset tiles page should have no a11y violations', () => {
    cy.visit('/tiles')
    cy.injectAxe()
    cy.checkA11y()
  })

  it("dataset tiles page should have valid HTML", () => {
    cy.visit("/tiles");
    cy.htmlvalidate();
  })

  it('dataset tiles metadata page should have no a11y violations', () => {
    cy.visit('/tiles/NetherlandsRDNewQuad')
    cy.injectAxe()
    cy.checkA11y()
  })

  it("dataset tiles metadata page should have valid HTML", () => {
    cy.visit("/tiles/NetherlandsRDNewQuad");
    cy.htmlvalidate();
  })

  it('geodata tiles page (collection-level) should have no a11y violations', () => {
    cy.visit('/collections/addresses/tiles')
    cy.injectAxe()
    cy.checkA11y()
  })

  it("geodata tiles page should have valid HTML", () => {
    cy.visit("/collections/addresses/tiles");
    cy.htmlvalidate();
  })

  it('geodata tiles metadata page (collection-level) should have no a11y violations', () => {
    cy.visit('/collections/addresses/tiles/NetherlandsRDNewQuad')
    cy.injectAxe()
    cy.checkA11y()
  })

  it("geodata tiles metadata page should have valid HTML", () => {
    cy.visit("/collections/addresses/tiles/NetherlandsRDNewQuad");
    cy.htmlvalidate();
  })

  it('tileMatrixSets page should have no a11y violations', () => {
    cy.visit('/tileMatrixSets')
    cy.injectAxe()
    cy.checkA11y()
  })

  it("tileMatrixSets page should have valid HTML", () => {
    cy.visit("/tileMatrixSets");
    cy.htmlvalidate();
  })

  it('specific tileMatrixSet (NetherlandsRDNewQuad) page should have no a11y violations', () => {
    cy.visit('/tileMatrixSets/NetherlandsRDNewQuad')
    cy.injectAxe()
    cy.checkA11y()
  })

  it("specific tileMatrixSet (NetherlandsRDNewQuad) page should have valid HTML", () => {
    cy.visit("/tileMatrixSets/NetherlandsRDNewQuad");
    cy.htmlvalidate();
  })
})