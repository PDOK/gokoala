import { HttpClientModule } from '@angular/common/http'
import { createOutputSpy } from 'cypress/angular'
import { Map as OLMap } from 'ol'
import { FeatureViewComponent } from 'src/app/feature-view/feature-view.component'
import 'cypress-network-idle'
import { LoggerModule } from 'ngx-logger'
import { environment } from 'src/environments/environment'

// eslint-disable-next-line @typescript-eslint/no-explicit-any
export const getTestTitle = (test: Mocha.Suite = (Cypress as any).mocha.getRunner().suite.ctx.test): string =>
  test.parent?.title ? `${getTestTitle(test.parent)} -- ${test.title}` : test.title
export function intercept(geofix: string) {
  cy.viewport(550, 750)
  cy.intercept('GET', 'https://test*', { fixture: geofix }).as('geo')
  if (Cypress.env('realmaps')) {
    cy.log('realmaps')
    cy.intercept('GET', '*grijs*').as('BRTbackground')
    cy.intercept('GET', 'https://tile.openstreetmap.org/*/*/*.png').as('OSMbackground')
  } else {
    cy.log('stubs maps')
    cy.intercept('GET', '*grijs*', { fixture: 'backgroundstub.png' }).as('BRTbackground')
    cy.intercept('GET', 'https://tile.openstreetmap.org/*/*/*.png', { fixture: '172300.png' }).as('OSMbackground')
  }
}

export function mountFeatureComponent(aprojection: string, abackground: 'OSM' | 'BRT' | undefined = 'OSM') {
  cy.mount(FeatureViewComponent, {
    imports: [
      HttpClientModule,
      LoggerModule.forRoot({
        level: environment.loglevel,
      }),
    ],
    componentProperties: {
      itemsUrl: 'https://test/',
      box: createOutputSpy('boxSpy'),
      backgroundMap: abackground,
      projection: aprojection,
    },
  }).then(comp1 => {
    cy.log(JSON.stringify(comp1))
    const map = comp1.component.map as OLMap
    map.addEventListener('loadend', cy.stub().as('MapLoaded'))
    const viewport = map.getViewport()
    const position = viewport.getBoundingClientRect()
    cy.log(`left: ${position.left}, top: ${position.top}, width: ${position.width}, height: ${position.height}`)
  })

  cy.wait('@geo')

  cy.get('@MapLoaded').should('have.been.calledOnce')
}

export function idle() {
  if (Cypress.env('networkIdle')) {
    cy.waitForNetworkIdle('*', '*', Cypress.env('networkIdle'))
  }
}
export function screenshot(aname: string = '') {
  cy.screenshot(aname)
}

export function zoomout(aname: string) {
  Cypress._.times(30, i => {
    cy.get('.ol-zoom-out').click()
    idle()
    cy.screenshot(getTestTitle() + aname + '_zoomout_' + i)
  })

  cy.screenshot('zoomout_final/' + aname)
}
