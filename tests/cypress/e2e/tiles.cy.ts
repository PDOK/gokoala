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

  it('dataset tiles page should have no broken links', () => {
    cy.visit('/tiles')
    cy.get('a').each(link => {
      const href = link.prop('href')
      if (href && !href.includes('example.com')) {
        cy.request(href)
      }
    })
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

  it('dataset tiles metadata page should have no broken links', () => {
    cy.visit('/tiles/NetherlandsRDNewQuad')
    cy.get('a').each(link => {
      const href = link.prop('href')
      if (href && !href.includes('example.com')) {
        cy.request(href)
      }
    })
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

  it('geodata tiles page should have no broken links', () => {
    cy.visit('/collections/addresses/tiles')
    cy.get('a').each(link => {
      const href = link.prop('href')
      if (href && !href.includes('example.com')) {
        cy.request(href)
      }
    })
  })

  it('geodata tiles metadata page (collection-level) should have no a11y violations', () => {
    cy.visit('/collections/addresses/tiles/NetherlandsRDNewQuad')
    cy.injectAxe()
    cy.checkA11y()
  })

  it("geodata tiles metadata page should have valid HTML", () => {
    cy.visit('/collections/addresses/tiles/NetherlandsRDNewQuad');
    cy.htmlvalidate();
  })

  it('geodata tiles metadata page should have no broken links', () => {
    cy.visit('/collections/addresses/tiles/NetherlandsRDNewQuad')
    cy.get('a').each(link => {
      const href = link.prop('href')
      if (href && !href.includes('example.com')) {
        cy.request(href)
      }
    })
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

  it('tileMatrixSets page should have no broken links', () => {
    cy.visit('/tileMatrixSets')
    cy.get('a').each(link => {
      const href = link.prop('href')
      if (href && !href.includes('example.com')) {
        cy.request(href)
      }
    })
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

  it('specific tileMatrixSet (NetherlandsRDNewQuad)  page should have no broken links', () => {
    cy.visit('/tileMatrixSets/NetherlandsRDNewQuad')
    cy.get('a').each(link => {
      const href = link.prop('href')
      if (href && !href.includes('example.com')) {
        cy.request(href)
      }
    })
  })
})