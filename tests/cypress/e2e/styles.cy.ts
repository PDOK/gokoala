describe('OGC API Styles tests', () => {

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
    cy.get('a').each(link => {
      const href = link.prop('href')
      if (href && !href.includes('example.com')) {
        cy.request(href)
      }
    })
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
    cy.get('a').each(link => {
      const href = link.prop('href')
      if (href && !href.includes('example.com')) {
        cy.request(href)
      }
    })
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
    cy.get('a').each(link => {
      const href = link.prop('href')
      if (href && !href.includes('example.com')) {
        cy.request(href)
      }
    })
  })
})