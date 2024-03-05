describe('accessibility (a11y) tests', () => {
  it('landing page no detectable a11y violations', () => {
    cy.visit('/')
    cy.injectAxe()
    cy.checkA11y()
  })

  // disabled since it has two violations in the 3rd party swagger-ui component (outside our control)
  xit('openapi no detectable a11y violations', () => {
    cy.visit('/api')
    cy.injectAxe()
    cy.checkA11y()
  })

  it('conformance page no detectable a11y violations', () => {
    cy.visit('/conformance')
    cy.injectAxe()
    cy.checkA11y()
  })

  it('collections page no detectable a11y violations', () => {
    cy.visit('/collections')
    cy.injectAxe()
    cy.checkA11y()
  })

  it('collection page no detectable a11y violations', () => {
    cy.visit('/collections/addresses')
    cy.injectAxe()
    cy.checkA11y()
  })

  it('features page no detectable a11y violations', () => {
    cy.visit('/collections/addresses/items')
    cy.injectAxe()
    cy.checkA11y()
  })

  it('feature page no detectable a11y violations', () => {
    cy.visit('/collections/addresses/items/1')
    cy.injectAxe()
    cy.checkA11y()
  })

  it('styles page no detectable a11y violations', () => {
    cy.visit('/styles')
    cy.injectAxe()
    cy.checkA11y()
  })

  it('styles metadata page no detectable a11y violations', () => {
    cy.visit('/styles/dummy-style/metadata')
    cy.injectAxe()
    cy.checkA11y()
  })

  it('tiles page no detectable a11y violations', () => {
    cy.visit('/tiles')
    cy.injectAxe()
    cy.checkA11y()
  })

  it('tiles metadata page no detectable a11y violations', () => {
    cy.visit('/tiles/NetherlandsRDNewQuad')
    cy.injectAxe()
    cy.checkA11y()
  })

  it('tileMatrixSets page no detectable a11y violations', () => {
    cy.visit('/tileMatrixSets')
    cy.injectAxe()
    cy.checkA11y()
  })

  it('specific tileMatrixSet (NetherlandsRDNewQuad) page no detectable a11y violations', () => {
    cy.visit('/tileMatrixSets/NetherlandsRDNewQuad')
    cy.injectAxe()
    cy.checkA11y()
  })
})