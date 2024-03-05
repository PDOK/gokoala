describe('OGC API Tiles tests', () => {

  it('tiles page should have no a11y violations', () => {
    cy.visit('/tiles')
    cy.injectAxe()
    cy.checkA11y()
  })

  it('tiles metadata page should have no a11y violations', () => {
    cy.visit('/tiles/NetherlandsRDNewQuad')
    cy.injectAxe()
    cy.checkA11y()
  })

  it('tileMatrixSets page should have no a11y violations', () => {
    cy.visit('/tileMatrixSets')
    cy.injectAxe()
    cy.checkA11y()
  })

  it('specific tileMatrixSet (NetherlandsRDNewQuad) page should have no a11y violations', () => {
    cy.visit('/tileMatrixSets/NetherlandsRDNewQuad')
    cy.injectAxe()
    cy.checkA11y()
  })
})