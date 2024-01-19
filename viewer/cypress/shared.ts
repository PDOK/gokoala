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

export function downloadPng(selector: string, filename: string) {
  expect(filename).to.be.a('string')
  expect(selector).to.be.a('string')
  const path = Cypress.config('screenshotsFolder') + '/' + filename
  return cy.get(selector).then(canvas => {
    const url = (canvas[0] as HTMLCanvasElement).toDataURL()
    const data = url.replace(/^data:image\/png;base64,/, '')
    cy.writeFile(path, data, 'base64')
    cy.wrap(path)
  })
}

export function injectAxe() {
  //cy.injectAxe();
  // cy.injectAxe is currently broken. https://github.com/component-driven/cypress-axe/issues/82

  // Creating our own injection logic
  cy.readFile('node_modules/axe-core/axe.min.js').then(source => {
    return cy.window({ log: false }).then(window => {
      window.eval(source)
    })
  })
}

export function checkAccessibility() {
  cy.get('body').should('be.visible')
  cy.checkA11y('body')
}

export function logAccessibility() {
  cy.log('Todo: fix or change to checkAccessibility()')
  cy.get('body')
    .should('be.visible')
    .then($el => {
      const el = $el.get(0) //native DOM element
      cy.log(el.innerHTML)
    })
  cy.checkA11y('body', null, null, true)
}
