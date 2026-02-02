import { Projection } from 'ol/proj'
import { register as proj4register } from 'ol/proj/proj4'
import proj4 from 'proj4'

export const NetherlandsRDNewQuadDefault = 'NetherlandsRDNewQuad'
export const EuropeanETRS89_LAEAQuad = 'EuropeanETRS89_LAEAQuad'

export function initProj4() {
  proj4.defs(
    'EPSG:28992',
    '+proj=sterea +lat_0=52.15616055555555 +lon_0=5.38763888888889 +k=0.9999079 +x_0=155000 +y_0=463000 +ellps=bessel +towgs84=565.417,50.3319,465.552,-0.398957,0.343988,-1.8774,4.0725 +units=m +no_defs'
  )

  proj4.defs(
    'EPSG:3035',
    '+proj=laea +lat_0=52 +lon_0=10 +x_0=4321000 +y_0=3210000 +ellps=GRS80 +towgs84=0,0,0,0,0,0,0 +units=m +no_defs +type=crs'
  )

  proj4.defs(
    'EPSG:4258',
    'GEOGCRS["ETRS89",ENSEMBLE["European Terrestrial Reference System 1989 ensemble", MEMBER["European Terrestrial Reference Frame 1989", ID["EPSG",1178]], MEMBER["European Terrestrial Reference Frame 1990", ID["EPSG",1179]], MEMBER["European Terrestrial Reference Frame 1991", ID["EPSG",1180]], MEMBER["European Terrestrial Reference Frame 1992", ID["EPSG",1181]], MEMBER["European Terrestrial Reference Frame 1993", ID["EPSG",1182]], MEMBER["European Terrestrial Reference Frame 1994", ID["EPSG",1183]], MEMBER["European Terrestrial Reference Frame 1996", ID["EPSG",1184]], MEMBER["European Terrestrial Reference Frame 1997", ID["EPSG",1185]], MEMBER["European Terrestrial Reference Frame 2000", ID["EPSG",1186]], MEMBER["European Terrestrial Reference Frame 2005", ID["EPSG",1204]], MEMBER["European Terrestrial Reference Frame 2014", ID["EPSG",1206]], MEMBER["European Terrestrial Reference Frame 2020", ID["EPSG",1382]], ELLIPSOID["GRS 1980",6378137,298.257222101,LENGTHUNIT["metre",1,ID["EPSG",9001]],ID["EPSG",7019]], ENSEMBLEACCURACY[0.1],ID["EPSG",6258]],CS[ellipsoidal,2,ID["EPSG",6422]],AXIS["Geodetic latitude (Lat)",north],AXIS["Geodetic longitude (Lon)",east],ANGLEUNIT["degree",0.0174532925199433,ID["EPSG",9102]],ID["EPSG",4258]]'
  )

  proj4register(proj4)
}

export function getRijksdriehoek() {
  const projectionExtent = [-285401.92, 22598.08, 595401.9199999999, 903401.9199999999]
  const RDprojection = new Projection({ code: 'EPSG:28992', units: 'm', extent: projectionExtent })
  const resolutions = [3440.64, 1720.32, 860.16, 430.08, 215.04, 107.52, 53.76, 26.88, 13.44, 6.72, 3.36, 1.68, 0.84, 0.42, 0.21]
  //const size = getWidth(projectionExtent) / 256
  const matrixIds = new Array(15)
  for (let z = 0; z < 15; ++z) {
    matrixIds[z] = 'EPSG:28992:' + z
  }
  return {
    projection: RDprojection,
    resolutions: resolutions,
    matrixIds: matrixIds,
  }
}

export class MapProjection {
  private _tileUrl: string

  constructor(tileUrl: string) {
    initProj4()
    this._tileUrl = tileUrl
  }

  public get Projection(): Projection {
    const rdProjection = new Projection({
      code: 'EPSG:28992',
      units: 'm',
      extent: [-285401.92, 22598.08, 595401.92, 903401.92],
      axisOrientation: 'enu',
      global: false,
    })

    const mercator = new Projection({
      code: 'EPSG:3857',
      units: 'm',
      extent: [-20037508.342789244, -20037508.342789244, 20037508.342789244, 20037508.342789244],
      worldExtent: [-180, -85, 180, 85],
      axisOrientation: 'enu',
      global: true,
    })

    const ETRS89projection = new Projection({
      axisOrientation: 'neu',
      code: 'EPSG:3035',
      units: 'm',
      extent: [2000000.0, 2164940.6031185603, 5394791.161618613, 5500000.0],
    })

    if (this._tileUrl.includes(NetherlandsRDNewQuadDefault)) {
      return rdProjection
    }
    if (this._tileUrl.includes(EuropeanETRS89_LAEAQuad)) {
      return ETRS89projection
    }
    return mercator
  }
}
