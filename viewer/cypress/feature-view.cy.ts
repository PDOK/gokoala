import { HttpClientModule } from '@angular/common/http'
import { createOutputSpy } from 'cypress/angular'
import { Map as OLMap } from 'ol'
import { FeatureViewComponent } from 'src/app/feature-view/feature-view.component'

// eslint-disable-next-line @typescript-eslint/no-explicit-any
const getTestTitle = (test: Mocha.Suite = (Cypress as any).mocha.getRunner().suite.ctx.test): string =>
  test.parent?.title ? `${getTestTitle(test.parent)} -- ${test.title}` : test.title

beforeEach(() => {
  cy.viewport(550, 750)
  cy.intercept('GET', 'https://test*', { fixture: 'amsterdam.json' }).as('geo')
  cy.intercept('GET', 'https://tile.openstreetmap.org/19/269273/172300.png', { fixture: '172300.png' }).as('background')
  cy.mount(FeatureViewComponent, {
    imports: [HttpClientModule],
    componentProperties: {
      itemsUrl: 'https://test',
      box: createOutputSpy('boxSpy'),
      backgroundMap: 'OSM',
    },
  }).then(comp1 => {
    console.log(comp1)
    const map = comp1.component.map as OLMap
    map.addEventListener('loadend', cy.stub().as('MapLoaded'))
    const viewport = map.getViewport()
    const position = viewport.getBoundingClientRect()
    cy.log(`left: ${position.left}, top: ${position.top}, width: ${position.width}, height: ${position.height}`)
  })
  cy.get('@MapLoaded').should('have.been.calledOnce')
  cy.wait('@geo')
})

describe('feature-view.cy.ts', () => {
  it('It shows Point from url', () => {
    cy.wait('@background')
    cy.screenshot(getTestTitle() + 'amsterdam')
    cy.get('.ol-zoom-out').click()
  })

  it('It can draw and emit boundingbox', () => {
    cy.get('.innersvg').click()

    cy.get('.ol-viewport').click(100, 100).click(200, 200)

    cy.screenshot(getTestTitle() + 'amsterdam')

    cy.get('@boxSpy').should('have.been.calledOnce')
    // .should('have.been.calledWith', '4.89516718294036,52.37021597417751,4.895167706985226,52.37021629414647')

    cy.get('@MapLoaded').should('have.been.calledOnce')
  })
})
