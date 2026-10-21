import { idle, injectAxe, logAccessibility, mountMapsComponent, screenshot } from './shared'

type ProjectionTest = { code: string; testName: string; collection: string }

const RD_NEW_URI = 'http://www.opengis.net/def/crs/EPSG/0/28992'
const MAX_SIZE = 5000

const tests: ProjectionTest[] = [
  { code: 'EPSG:28992', testName: 'RD dutch projection (EPSG code)', collection: 'foo' },
  { code: RD_NEW_URI, testName: 'RD dutch projection (OGC URI)', collection: 'bar' },
]

tests.forEach(i => {
  describe(i.collection + '-maps view-' + i.testName, () => {
    it('It shows the map image for ' + i.code, () => {
      injectAxe()
      mountMapsComponent(i.code, `https://test/collections/${i.collection}/map`, '10000,300000,280000,620000')
      cy.wait('@map')
      idle()
      screenshot('maps-' + i.collection)
      logAccessibility('body')
    })

    it('It requests the map with OGC CRS URIs and only supported parameters for ' + i.code, () => {
      mountMapsComponent(i.code, `https://test/collections/${i.collection}/map`, '10000,300000,280000,620000')
      cy.wait('@map').then(({ request }) => {
        const params = new URL(request.url).searchParams
        expect(params.get('f')).to.equal('png')
        expect(params.get('crs')).to.equal(RD_NEW_URI)
        expect(params.get('bbox-crs')).to.equal(RD_NEW_URI)
        expect(params.get('bbox')?.split(',')).to.have.length(4)
        expect(Number(params.get('width'))).to.be.within(1, MAX_SIZE)
        expect(Number(params.get('height'))).to.be.within(1, MAX_SIZE)
      })
    })
  })
})
