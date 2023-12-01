import { HttpClientModule } from '@angular/common/http'
import { createOutputSpy } from 'cypress/angular'
import { Map as OLMap } from 'ol'
import { FeatureViewComponent } from 'src/app/feature-view/feature-view.component'

// eslint-disable-next-line @typescript-eslint/no-explicit-any
const getTestTitle = (test: Mocha.Suite = (Cypress as any).mocha.getRunner().suite.ctx.test): string =>
  test.parent?.title ? `${getTestTitle(test.parent)} -- ${test.title}` : test.title

beforeEach(() => {
  cy.viewport(550, 750)
  cy.intercept('GET', 'https://api.pdok.nl/items', { fixture: 'pdokwegdelen.json' }).as('geo')
  cy.intercept('GET', '*grijs*', { fixture: 'backgroundstub.png' }).as('background')
  cy.mount(FeatureViewComponent, {
    imports: [HttpClientModule],
    componentProperties: {
      itemsUrl: 'https://api.pdok.nl/items',
      box: createOutputSpy('boxSpy'),
      projection: 'EPSG:28992',
      backgroundMap: 'BRT',
    },
  }).then(comp1 => {
    //  console.log(comp1)
    const map = comp1.component.map as OLMap
    map.addEventListener('loadend', cy.stub().as('MapLoaded'))
    const viewport = map.getViewport()
    const position = viewport.getBoundingClientRect()
    cy.log(`left: ${position.left}, top: ${position.top}, width: ${position.width}, height: ${position.height}`)
  })
  cy.wait('@geo')
  cy.wait('@background')
  cy.get('@MapLoaded').should('have.been.calledOnce')
})

describe('feature-view.cy.ts works for RD', () => {
  it('It can draw and emit boundingbox in RD', () => {
    cy.get('.ol-viewport').trigger('pointerdown', {
      eventConstructor: 'MouseEvent',
      x: 100,
      y: 100,
      force: true,
      isPrimary: true,
      ctrlKey: true,
    })

    cy.get('.ol-viewport').trigger('pointermove', { x: 100, y: 100, ctrlKey: true })
    //   cy.wait(1000)
    cy.get('.ol-viewport').trigger('pointermove', { x: 200, y: 200, ctrlKey: true })
    //  cy.wait(1000)
    cy.screenshot(getTestTitle() + 'amsterdam')
    cy.get('.ol-viewport').trigger('pointerup', { eventConstructor: 'MouseEvent', force: true, ctrlKey: true })
    //to do new interaction    cy.get('@boxSpy').should('have.been.calledOnce')
    //.should('have.been.calledWith', Cypress.sinon.match('/1586*/gm'))
    //.should('have.been.calledWith', Cypress.sinon.match('/1586/d/d./d*,3735/d/d./d*,159/d/d/d./d*,3745/d/d./d*/gm'))
    //  should(expect.stringMatching('have.been.calledWithMatch', '1586*,3735*,159*,3745*'))
    // -[ '158626.42180172657,373585.0164062996,159592.38087845314,374549.73239709437' ]
    ///  '/1586/d/d./d*,      3735/d/d./d*,     159/d/d/d./d*,     3745/d/d./d*/gm'
    // +[ '158630.1373691982,373577.4538369284,159599.81201339638,374547.1580840892' ]

    cy.get('@MapLoaded').should('have.been.calledOnce')
  })
})
