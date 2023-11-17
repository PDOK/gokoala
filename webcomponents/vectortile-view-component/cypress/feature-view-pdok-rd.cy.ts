import { HttpClientModule } from '@angular/common/http'
import { createOutputSpy } from 'cypress/angular'
import { Map as OLMap } from 'ol'
import { FeatureViewComponent } from 'src/app/feature-view/feature-view.component'

const getTestTitle = (test: Mocha.Suite = (Cypress as any).mocha.getRunner().suite.ctx.test): string =>
  test.parent?.title ? `${getTestTitle(test.parent)} -- ${test.title}` : test.title

beforeEach(() => {
  cy.intercept('GET', 'https://api.pdok.nl/*', { fixture: 'pdokwegdelen.json' }).as('geo')
  cy.intercept('GET', '*grijs*', { fixture: 'backgroundstub.png' }).as('background')
  cy.mount(FeatureViewComponent, {
    imports: [HttpClientModule],
    componentProperties: {
      itemsUrl: 'https://api.pdok.nl/lv/bgt/ogc/v1_0-preprod/collections/wegdelen/items',
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

    cy.get('@boxSpy')
      .should('have.been.calledOnce')
      .should('have.been.calledWith', '158733.95734529808,373369.3879571492,159807.45196559615,374444.4679655064')

    cy.get('@MapLoaded').should('have.been.calledOnce')
  })
})