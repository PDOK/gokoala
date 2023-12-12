import { idle, intercept, mountFeatureComponent, screenshot } from './shared'

type ProjectionTest = { code: string; projection: string; geofix: string }

const tests: ProjectionTest[] = [
  { code: 'CRS84', projection: 'https://www.opengis.net/def/crs/OGC/1.3/CRS84', geofix: 'amsterdam-wgs84.json' },
  { code: 'EPSG:28992', projection: 'http://www.opengis.net/def/crs/EPSG/0/28992', geofix: 'amsterdam-epgs28992.json' },
  { code: 'EPSG:3035', projection: 'http://www.opengis.net/def/crs/EPSG/0/3035', geofix: 'amsterdam-epgs3035.json' },
]

tests.forEach(i => {
  describe(i.geofix + '-feature view', () => {
    it('It shows Point from url on OSM ', () => {
      intercept(i.geofix)
      mountFeatureComponent(i.projection, 'OSM')
      idle()
      screenshot('OSM-' + i.code)
    })

    it('It can draw and emit boundingbox in ' + i.geofix + 'on BRT', () => {
      intercept(i.geofix)
      mountFeatureComponent(i.projection, 'BRT')
      cy.get('.innersvg').click()
      cy.get('.ol-viewport').click(100, 100).click(200, 200)
      screenshot('BRT-bbox' + i.code)
    })
  })
})
