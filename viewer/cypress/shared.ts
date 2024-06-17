import { HttpClientModule } from '@angular/common/http'
import { createOutputSpy } from 'cypress/angular'
import { Map as OLMap } from 'ol'
import { FeatureViewComponent } from 'src/app/feature-view/feature-view.component'
import 'cypress-network-idle'
import { LoggerModule } from 'ngx-logger'
import { environment } from 'src/environments/environment'

import { register as proj4register } from 'ol/proj/proj4'
import proj4 from 'proj4'
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
export function intercept(geofix: string, realmaps = Cypress.env('realmaps'), url: string = 'https://test/items*') {
  cy.viewport(550, 750)
  cy.intercept('GET', url, { fixture: geofix }).as('geo')

  if (realmaps) {
    cy.log('realmaps')
    cy.intercept('GET', '*grijs*').as('BRTbackground')
    cy.intercept('GET', 'https://tile.openstreetmap.org/*/*/*.png').as('OSMbackground')

    cy.log('stubs maps')
    cy.intercept('GET', '*grijs*', { fixture: 'backgroundstub.png' }).as('BRTbackground')
    cy.intercept('GET', 'https://tile.openstreetmap.org/*/*/*.png', { fixture: '172300.png' }).as('OSMbackground')
  }
}

interface Prop {
  [key: string]: string
}

export function mountFeatureComponent(
  aprojection: string,
  abackground: 'OSM' | 'BRT' | undefined = 'OSM',
  amode: 'auto' | 'default' | undefined = 'default',
  aprop: Prop = { itemsUrl: 'https://test/items' }
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
        level: environment.loglevel,
      }),
    ],
    componentProperties: allprop,
  }).then(comp1 => {
    const map = comp1.component.map as OLMap
    map.addEventListener('loadend', cy.stub().as('MapLoaded'))

    const viewport = map.getViewport()
    const position = viewport.getBoundingClientRect()
    cy.log(`left: ${position.left}, top: ${position.top}, width: ${position.width}, height: ${position.height}`)
    cy.log(JSON.stringify(comp1.component.itemsUrl))
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

export function generateRDSquareGrid(xStart: number, yStart: number, gridSize: number, count: number) {
  proj4.defs(
    'EPSG:28992',
    '+proj=sterea +lat_0=52.15616055555555 +lon_0=5.38763888888889 +k=0.9999079 +x_0=155000 +y_0=463000 +ellps=bessel +towgs84=565.417,50.3319,465.552,-0.398957,0.343988,-1.8774,4.0725 +units=m +no_defs'
  )

  proj4.defs(
    'EPSG:3035',
    '+proj=laea +lat_0=52 +lon_0=10 +x_0=4321000 +y_0=3210000 +ellps=GRS80 +towgs84=0,0,0,0,0,0,0 +units=m +no_defs +type=crs'
  )

  proj4.defs('EPSG:4258', '+proj=longlat +ellps=GRS80 +no_defs +type=crs')

  proj4register(proj4)

  const features = []

  for (let x = 0; x < count; x++) {
    for (let y = 0; y < count; y++) {
      const x0 = xStart + x * gridSize
      const y0 = yStart + y * gridSize
      const x1 = x0 + gridSize
      const y1 = y0 + gridSize

      const bottomLeft = proj4('EPSG:28992').inverse([x0, y0])
      const topLeft = proj4('EPSG:28992').inverse([x0, y1])
      const topRight = proj4('EPSG:28992').inverse([x1, y1])
      const bottomRight = proj4('EPSG:28992').inverse([x1, y0])

      const feature = {
        id: 'id' + x + ' ' + y,
        type: 'Polygon',
        coordinates: [bottomLeft, topLeft, topRight, bottomRight, bottomLeft],
        properties: {
          label: 'label: ' + x + ' ' + y,
        }
      }

      features.push(feature)
    }
  }

  return {
    type: 'FeatureCollection',
    features: features,
  }
}
