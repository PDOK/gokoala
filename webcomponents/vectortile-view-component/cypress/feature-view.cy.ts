import { HttpClientModule } from '@angular/common/http'
import { createOutputSpy } from 'cypress/angular'
import { FeatureViewComponent } from 'src/app/feature-view/feature-view.component'

beforeEach(() => {
  cy.intercept('GET', 'https://test*', { fixture: 'amsterdam.json' }).as('geo')
  cy.intercept('https://tile.openstreetmap.org/19/269273/172300.png', { fixture: '172300.png' }).as('background')
})

describe('feature-view.cy.ts', () => {
  it('It shows Point from url', () => {
    cy.mount(FeatureViewComponent, {
      imports: [HttpClientModule],
      autoSpyOutputs: true,
      componentProperties: {
        itemsUrl: 'https://test',
        box: createOutputSpy('boxSpy'),
      },
    }).then(comp1 => {
      console.log(comp1)
    })
  })

  it('It can draw and emit boundingbox', () => {
    cy.mount(FeatureViewComponent, {
      imports: [HttpClientModule],
      autoSpyOutputs: true,
      componentProperties: {
        itemsUrl: 'https://test',
      },
    }).then(comp => {
      console.log(comp)
      cy.wait('@geo')
      cy.wait('@background')
      const viewport = comp.component.map.getViewport()
      const position = viewport.getBoundingClientRect()
      cy.log(`left: ${position.left}, top: ${position.top}, width: ${position.width}, height: ${position.height}`)
      //     cy.get('canvas').click()
      cy.get('.ol-viewport').trigger('pointerdown', {
        eventConstructor: 'MouseEvent',
        x: 100,
        y: 100,
        force: true,
        isPrimary: true,
        ctrlKey: true,
      })
      cy.wait(1000)
      cy.get('.ol-viewport').trigger('pointermove', { x: 100, y: 100, ctrlKey: true })
      //   cy.wait(1000)
      cy.get('.ol-viewport').trigger('pointermove', { x: 200, y: 200, ctrlKey: true })
      //  cy.wait(1000)
      cy.get('.ol-viewport').trigger('pointerup', { eventConstructor: 'MouseEvent', force: true, ctrlKey: true })

      cy.get('@boxSpy').should('have.been.called')
    })
  })
})
