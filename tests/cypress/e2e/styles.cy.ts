describe('OGC API Styles tests', () => {

  it('styles page should have no a11y violations', () => {
    cy.visit('/styles')
    cy.injectAxe()
    cy.checkA11y()
  })

  it('styles metadata page should have no a11y violations', () => {
    cy.visit('/styles/dummy-style/metadata')
    cy.injectAxe()
    cy.checkA11y()
  })
})