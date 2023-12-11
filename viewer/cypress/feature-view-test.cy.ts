import { idle, intercept, mountFeatureComponent, screenshot } from './shared'

type ProjectionTest = { code: string; projection: string; geofix: string }

const tests: ProjectionTest[] = [
  { code: 'CRS84', projection: 'https://www.opengis.net/def/crs/OGC/1.3/CRS84', geofix: 'amsterdam-wgs84.json' },
  { code: 'epgs28992', projection: 'http://www.opengis.net/def/crs/EPSG/0/28992', geofix: 'amsterdam-epgs28992.json' },
  { code: 'epgs3035', projection: 'http://www.opengis.net/def/crs/EPSG/0/3035', geofix: 'amsterdam-epgs3035.json' },
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
      //cy.get('@boxSpy').should('have.been.calledOnce')
      // .should('have.been.calledWith', '4.89516718294036,52.37021597417751,4.895167706985226,52.37021629414647')
      // cy.get('@MapLoaded').should('have.been.calledOnce')
      // cy.get('.ol-zoom-out')
    })
  })
})
