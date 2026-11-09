import { idle, injectAxe, logAccessibility, mountMapsComponent, screenshot } from './shared'
import ImageLayer from 'ol/layer/Image'
import TileLayer from 'ol/layer/Tile'
import { OSM, WMTS } from 'ol/source'
import { transform } from 'ol/proj'

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
        expect(params.get('transparent')).to.equal('true')
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

describe('map transparency', () => {
  it('forwards explicit false to map image requests', () => {
    mountMapsComponent(RD_NEW_URI, 'https://test/collections/foo/map', RD_BBOX, false)
    cy.wait('@map').then(({ request }) => {
      expect(new URL(request.url).searchParams.get('transparent')).to.equal('false')
    })
  })
})

describe('map backgrounds', () => {
  for (const background of ['OSM', 'BRT'] as const) {
    it(`shows ${background} below the map without changing the RD output CRS`, () => {
      mountMapsComponent(RD_NEW_URI, 'https://test/collections/foo/map', RD_BBOX, true, background).then(({ component }) => {
        const layers = component.map.getLayers().getArray()
        expect(layers).to.have.length(2)
        const backgroundLayer = layers[0]
        if (!(backgroundLayer instanceof TileLayer)) {
          throw new Error('Expected background tile layer')
        }
        expect(backgroundLayer.getSource()).to.be.instanceOf(background === 'OSM' ? OSM : WMTS)
        expect(layers[1]).to.be.instanceOf(ImageLayer)
        expect(component.map.getView().getProjection().getCode()).to.equal(RD_NEW_URI)
        const osmCenter = transform([155000, 463000], RD_NEW_URI, 'EPSG:3857')
        expect(osmCenter[0]).to.be.within(590000, 610000)
        expect(osmCenter[1]).to.be.within(6800000, 6900000)
        if (background === 'BRT') {
          expect((backgroundLayer.getSource() as WMTS).getMatrixSet()).to.equal('EPSG:28992')
        }
      })
      cy.get('@MapLoaded').should('have.been.called')
      cy.get('[role="status"]').should('not.exist')
    })
  }

  it('keeps the map when BRT does not support the output CRS', () => {
    mountMapsComponent(ETRS89_LAEA_URI, 'https://test/collections/foo/map', '3800000,3000000,4100000,3400000', true, 'BRT').then(
      ({ component }) => {
        expect(component.map.getLayers().getLength()).to.equal(1)
        expect(component.map.getView().getProjection().getCode()).to.equal(ETRS89_LAEA_URI)
      }
    )
    cy.wait('@map')
    cy.get('[role="status"]').should('contain.text', 'BRT background map is unavailable')
  })
})
