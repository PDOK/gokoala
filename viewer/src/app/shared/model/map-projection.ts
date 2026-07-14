import { Projection } from 'ol/proj'
import proj4 from 'proj4'
import { CrsMap } from './crs-map'
import { register } from 'ol/proj/proj4'
import { NGXLogger } from 'ngx-logger'
import { get as getProj } from 'ol/proj'

export const NetherlandsRDNewQuadDefault = 'NetherlandsRDNewQuad'
export const EuropeanETRS89_LAEAQuad = 'EuropeanETRS89_LAEAQuad'

const CRS_84_SRID = '100000'
const EPSG_PREFIX = 'EPSG:'

function initProj4WithTilesDefaults(logger: NGXLogger) {
  const crsMap: CrsMap = {
    '28992':
      '+proj=sterea +lat_0=52.15616055555555 +lon_0=5.38763888888889 +k=0.9999079 +x_0=155000 +y_0=463000 +ellps=bessel +towgs84=565.417,50.3319,465.552,-0.398957,0.343988,-1.8774,4.0725 +units=m +no_defs',
    '3035': '+proj=laea +lat_0=52 +lon_0=10 +x_0=4321000 +y_0=3210000 +ellps=GRS80 +towgs84=0,0,0,0,0,0,0 +units=m +no_defs +type=crs',
    '4258': `GEOGCRS["ETRS89",
    ENSEMBLE["European Terrestrial Reference System 1989 ensemble",
        MEMBER["European Terrestrial Reference Frame 1989"],
        MEMBER["European Terrestrial Reference Frame 1990"],
        MEMBER["European Terrestrial Reference Frame 1991"],
        MEMBER["European Terrestrial Reference Frame 1992"],
        MEMBER["European Terrestrial Reference Frame 1993"],
        MEMBER["European Terrestrial Reference Frame 1994"],
        MEMBER["European Terrestrial Reference Frame 1996"],
        MEMBER["European Terrestrial Reference Frame 1997"],
        MEMBER["European Terrestrial Reference Frame 2000"],
        MEMBER["European Terrestrial Reference Frame 2005"],
        MEMBER["European Terrestrial Reference Frame 2014"],
        MEMBER["European Terrestrial Reference Frame 2020"],
        ELLIPSOID["GRS 1980",6378137,298.257222101,
            LENGTHUNIT["metre",1]],
        ENSEMBLEACCURACY[0.1]],
    PRIMEM["Greenwich",0,
        ANGLEUNIT["degree",0.0174532925199433]],
    CS[ellipsoidal,2],
        AXIS["geodetic latitude (Lat)",north,
            ORDER[1],
            ANGLEUNIT["degree",0.0174532925199433]],
        AXIS["geodetic longitude (Lon)",east,
            ORDER[2],
            ANGLEUNIT["degree",0.0174532925199433]],
    USAGE[
        SCOPE["Spatial referencing."],
        AREA["Europe - onshore and offshore: Albania; Andorra; Austria; Belgium; Bosnia and Herzegovina; Bulgaria; Croatia; Czechia; Denmark; Estonia; Faroe Islands; Finland; France; Germany; Gibraltar; Greece; Hungary; Ireland; Italy; Kosovo; Latvia; Liechtenstein; Lithuania; Luxembourg; Malta; Moldova; Monaco; Montenegro; Netherlands; North Macedonia; Norway including Svalbard and Jan Mayen; Poland; Portugal - mainland; Romania; San Marino; Serbia; Slovakia; Slovenia; Spain - mainland and Balearic islands; Sweden; Switzerland; United Kingdom (UK) including Channel Islands and Isle of Man; Vatican City State."],
        BBOX[33.26,-16.1,84.73,38.01]],
    ID["EPSG",4258]]`,
  }
  initProj4WithDynamicCrs(crsMap, logger)
}

export function initProj4WithDynamicCrs(crsMap: CrsMap, logger: NGXLogger) {
  try {
    if (Object.keys(crsMap).length === 0) logger.warn('No CRS defined, only the default projection for OL (WGS84) is available')
    Object.keys(crsMap).forEach(key => {
      if (key === CRS_84_SRID) return // skip CRS84 as it ships with proj4js
      const proj4def = crsMap[key]
      const code = EPSG_PREFIX + key
      if (!getProj(code)) proj4.defs(code, proj4def)
    })
  } catch (error) {
    logger.error(`Error registering projections: ${error}`)
  }
  register(proj4)
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

  constructor(tileUrl: string, logger: NGXLogger) {
    initProj4WithTilesDefaults(logger)
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
