describe('OGC API Features tests', () => {

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
    cy.get('a').each(link => {
      const href = link.prop('href')
      if (href && !href.includes('example.com')) {
        cy.request(href)
      }
    })
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
    cy.get('a').each(link => {
      const href = link.prop('href')
      if (href && !href.includes('example.com')) {
        cy.request(href)
      }
    })
  })
})