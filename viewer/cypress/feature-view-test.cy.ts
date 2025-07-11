import { idle, injectAxe, intercept, logAccessibility, mountFeatureComponent, screenshot } from './shared'

type ProjectionTest = { code: string; testName: string; projection: string; geofix: string }

const tests: ProjectionTest[] = [
  {
    code: 'CRS84',
    testName: 'CRS84',
    projection: 'https://www.opengis.net/def/crs/OGC/1.3/CRS84',
    geofix: 'amsterdam-wgs84.json',
  },
  {
    code: 'EPSG:4258',
    testName: 'etrs europe',
    projection: 'http://www.opengis.net/def/crs/EPSG/0/4258',
    geofix: 'amsterdam-epsg4258.json',
  },
  {
    code: 'EPSG:28992',
    testName: 'RD dutch projection',
    projection: 'http://www.opengis.net/def/crs/EPSG/0/28992',
    geofix: 'amsterdam-epgs28992.json',
  },
  {
    code: 'EPSG:3035',
    testName: 'ETRS89-extended / LAEA Europe',
    projection: 'http://www.opengis.net/def/crs/EPSG/0/3035',
    geofix: 'amsterdam-epgs3035.json',
  },
]

tests.forEach(i => {
  describe(i.geofix + '-feature view-' + i.testName, () => {
    it('It shows Point from url on OSM ', () => {
      injectAxe()
      intercept(i.geofix, false)
      mountFeatureComponent(i.projection, 'OSM')
      idle()
      screenshot('OSM-' + i.code)
      logAccessibility('body')
    })
    it('It can draw and emit boundingbox in ' + i.geofix + 'on BRT', () => {
      intercept(i.geofix, false)
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
