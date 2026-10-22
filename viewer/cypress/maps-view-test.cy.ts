import { idle, injectAxe, logAccessibility, mountMapsComponent, screenshot } from './shared'

type ProjectionTest = { code: string; testName: string; collection: string; uri: string; bbox: string }

const RD_NEW_URI = 'http://www.opengis.net/def/crs/EPSG/0/28992'
const ETRS89_LAEA_URI = 'http://www.opengis.net/def/crs/EPSG/0/3035'
const ETRS89_URI = 'http://www.opengis.net/def/crs/EPSG/0/4258'
const RD_BBOX = '10000,300000,280000,620000'
const MAX_SIZE = 5000

const tests: ProjectionTest[] = [
  { code: 'EPSG:28992', testName: 'RD dutch projection (EPSG code)', collection: 'foo', uri: RD_NEW_URI, bbox: RD_BBOX },
  { code: RD_NEW_URI, testName: 'RD dutch projection (OGC URI)', collection: 'bar', uri: RD_NEW_URI, bbox: RD_BBOX },
  {
    code: 'EPSG:3035',
    testName: 'ETRS89 LAEA projection (EPSG code)',
    collection: 'baz',
    uri: ETRS89_LAEA_URI,
    bbox: '3800000,3000000,4100000,3400000',
  },
  { code: ETRS89_URI, testName: 'ETRS89 geographic projection (OGC URI)', collection: 'qux', uri: ETRS89_URI, bbox: '3,50,8,54' },
]

tests.forEach(i => {
  describe(i.collection + '-maps view-' + i.testName, () => {
    it('It shows the map image with OGC CRS URIs and only supported parameters for ' + i.code, () => {
      injectAxe()
      mountMapsComponent(i.code, `https://test/collections/${i.collection}/map`, i.bbox)
      cy.wait('@map').then(({ request }) => {
        const params = new URL(request.url).searchParams
        expect(params.get('f')).to.equal('png')
        expect(params.get('crs')).to.equal(i.uri)
        expect(params.get('bbox-crs')).to.equal(i.uri)
        expect(params.get('bbox')?.split(',')).to.have.length(4)
        expect(Number(params.get('width'))).to.be.within(1, MAX_SIZE)
        expect(Number(params.get('height'))).to.be.within(1, MAX_SIZE)
      })
      idle()
      screenshot('maps-' + i.collection)
      logAccessibility('body')
    })
  })
})
