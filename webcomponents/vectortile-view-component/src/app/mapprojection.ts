import { Projection } from 'ol/proj';
import { register as proj4register } from 'ol/proj/proj4';
import proj4 from 'proj4';

export class MapProjection {
  private _tileUrl: string;

  constructor(tileUrl: string) {
    this.initProj4()
    this._tileUrl=tileUrl
  }

  private initProj4() {
    proj4.defs("EPSG:28992", "+proj=sterea +lat_0=52.15616055555555 +lon_0=5.38763888888889 +k=0.9999079 +x_0=155000 +y_0=463000 +ellps=bessel +towgs84=565.417,50.3319,465.552,-0.398957,0.343988,-1.8774,4.0725 +units=m +no_defs");
    proj4.defs("EPSG:4258", "+proj=longlat +ellps=GRS80 +no_defs +type=crs");
    proj4register(proj4);
  }

  public get Projection(): Projection {

    const rDprojection = new Projection({
      "code": "EPSG:28992",
      "units": "m",
      "extent": [-285401.92, 22598.08, 595401.92, 903401.92],
      "axisOrientation": "enu",
      "global": false,
    })
    
    const mercator = new Projection({
      code: "EPSG:3857",
      units: "m",
      extent: [-20037508.342789244, -20037508.342789244, 20037508.342789244, 20037508.342789244],
      worldExtent: [-180, -85, 180, 85],
      axisOrientation: "enu",
      global: true
    })

    const ETRS89projection = new Projection({
      code: "EPSG:4258",
      units: "m",
      extent: [-16.1, 32.88, 39.65, 84.17],
      axisOrientation: "enu",
      global: false
    })


    if (this._tileUrl.includes('NetherlandsRDNewQuad')) {
      return rDprojection;
    } else {
      if (this._tileUrl.includes('EuropeanETRS89_GRS80')) {
        return ETRS89projection

      } else
        return mercator;

    }
  }

} 