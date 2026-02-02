import { HttpClientModule } from '@angular/common/http'
import { createOutputSpy } from 'cypress/angular'
import { Map as OLMap } from 'ol'
import { FeatureViewComponent } from 'src/app/feature-view/feature-view.component'
import 'cypress-network-idle'
import { LoggerModule, NgxLoggerLevel } from 'ngx-logger'

export type ProjectionTest = { code: string; projection: string; geofix: string }

export const tests: ProjectionTest[] = [
  { code: 'CRS84', projection: 'https://www.opengis.net/def/crs/OGC/1.3/CRS84', geofix: 'amsterdam-wgs84.json' },
  { code: 'EPSG:4258', projection: 'http://www.opengis.net/def/crs/EPSG/0/4258', geofix: 'amsterdam-epsg4258.json' },
  { code: 'EPSG:28992', projection: 'http://www.opengis.net/def/crs/EPSG/0/28992', geofix: 'amsterdam-epgs28992.json' },
  { code: 'EPSG:3035', projection: 'http://www.opengis.net/def/crs/EPSG/0/3035', geofix: 'amsterdam-epgs3035.json' },
]

// eslint-disable-next-line @typescript-eslint/no-explicit-any
export const getTestTitle = (test: Mocha.Suite = (Cypress as any).mocha.getRunner().suite.ctx.test): string =>
  test.parent?.title ? `${getTestTitle(test.parent)} -- ${test.title}` : test.title
export function intercept(geofix: string, realmaps: boolean = Cypress.env('realmaps'), url: string = 'https://test/items*') {
  cy.viewport(550, 750)
  cy.intercept('GET', url, { fixture: geofix }).as('geo')

  if (realmaps) {
    cy.log('realmaps')
    cy.intercept('GET', '*grijs*').as('BRTbackground')
    cy.intercept('GET', 'https://tile.openstreetmap.org/*/*/*.png').as('OSMbackground')
  } else {
    cy.log('stubs maps')
    cy.intercept('GET', '*grijs*', { fixture: 'backgroundstub.png' }).as('BRTbackground')
    cy.intercept('GET', 'https://tile.openstreetmap.org/*/*/*.png', { fixture: '172300.png' }).as('OSMbackground')
  }
}

interface Prop {
  [key: string]: string | number | string[]
}

export function mountFeatureComponent(
  aprojection: string,
  abackground: 'OSM' | 'BRT' | undefined = 'OSM',
  amode: 'auto' | 'default' | undefined = 'default',
  aprop: Prop = { itemUrls: ['https://test/items'] }
) {
  const prop: Prop = {
    box: createOutputSpy('boxSpy'),
    backgroundMap: abackground,
    projection: aprojection,
    mode: amode,
  }

  const allprop = { ...prop, ...aprop }
  cy.log(JSON.stringify(allprop))

  cy.mount(FeatureViewComponent, {
    imports: [
      HttpClientModule,
      LoggerModule.forRoot({
        level: NgxLoggerLevel.DEBUG,
      }),
    ],
    componentProperties: allprop,
  }).then(comp1 => {
    const map = comp1.component.map as unknown as OLMap
    map.addEventListener('loadend', cy.stub().as('MapLoaded'))

    const viewport = map.getViewport()
    const position = viewport.getBoundingClientRect()
    cy.log(`left: ${position.left}, top: ${position.top}, width: ${position.width}, height: ${position.height}`)
    cy.log(JSON.stringify(comp1.component.itemUrls))
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
  Cypress.Screenshot.defaults({
    overwrite: true,
  })
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

export function checkAccessibility(selector: string) {
  cy.get(selector).should('be.visible')
  cy.checkA11y(selector)
}

export function logAccessibility(selector: string) {
  cy.log('Todo: fix or change to checkAccessibility()')
  cy.get(selector)
    .should('be.visible')
    .then($el => {
      const el = $el.get(0) //native DOM element
      cy.log(el.innerHTML)
    })
  cy.checkA11y(selector, undefined, undefined, true)
}
