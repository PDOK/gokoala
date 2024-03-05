describe('OGC API Features tests', () => {

  it('features page should have no a11y violations', () => {
    cy.visit('/collections/addresses/items')
    cy.injectAxe()
    cy.checkA11y()
  })

  it('feature page should have no a11y violations', () => {
    cy.visit('/collections/addresses/items/1')
    cy.injectAxe()
    cy.checkA11y()
  })
})