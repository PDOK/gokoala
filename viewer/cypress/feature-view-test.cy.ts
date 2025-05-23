import { Feature } from 'ol'
import { idle, injectAxe, intercept, logAccessibility, mountFeatureComponent, screenshot } from './shared'
import { Polygon } from 'ol/geom'

type ProjectionTest = { code: string; projection: string; geofix: string }

const tests: ProjectionTest[] = [
  { code: 'CRS84', projection: 'https://www.opengis.net/def/crs/OGC/1.3/CRS84', geofix: 'amsterdam-wgs84.json' },
  { code: 'EPSG:4258', projection: 'http://www.opengis.net/def/crs/EPSG/0/4258', geofix: 'amsterdam-epsg4258.json' },
  { code: 'EPSG:28992', projection: 'http://www.opengis.net/def/crs/EPSG/0/28992', geofix: 'amsterdam-epgs28992.json' },
  { code: 'EPSG:3035', projection: 'http://www.opengis.net/def/crs/EPSG/0/3035', geofix: 'amsterdam-epgs3035.json' },
]

tests.forEach(i => {
  describe(i.geofix + '-feature view', () => {
    it('It shows Point from url on OSM ', () => {
      injectAxe()
      intercept(i.geofix)
      mountFeatureComponent(i.projection, 'OSM')
      idle()
      screenshot('OSM-' + i.code)
      logAccessibility('body')
    })
    it('It can draw and emit boundingbox in ' + i.geofix + 'on BRT', () => {
      intercept(i.geofix)
      mountFeatureComponent(i.projection, 'BRT')
      cy.get('.innersvg').click()
      cy.get('.ol-viewport').click(100, 100)
      cy.get('.ol-viewport').click(200, 200)
      screenshot('BRT-bbox' + i.code)
      cy.get('@boxSpy')
        .should('have.been.calledOnce')
        // eslint-disable-next-line @typescript-eslint/no-explicit-any
        .should((spy: any) => {
          const firstCallArgs = spy.getCall(0).args[0].split(',')
          expect(firstCallArgs[0]).to.match(/^4.8/)
          expect(firstCallArgs[1]).to.match(/^52.37/)
        })
    })
  })
})

describe('searchbox for location API', () => {
  it('It can draw feature on it', () => {
    intercept('amsterdam-epgs28992.json')

    const coordinates = [
      [115000, 500000], // Top-left corner (northwest)
      [125000, 500000], // Top-right corner (northeast)
      [125000, 480000], // Bottom-right corner (southeast)
      [115000, 480000], // Bottom-left corner (southwest)
      [115000, 500000], // Closing the polygon by returning to the first point
    ]

    const drawFeature = new Feature({
      geometry: new Polygon([coordinates]),
    })

    mountFeatureComponent('http://www.opengis.net/def/crs/EPSG/0/28992', 'BRT', 'default', { itemsUrl: 'https://test/items' }, drawFeature)
    screenshot('drawFeature')
  })
})
