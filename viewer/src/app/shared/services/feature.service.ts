import { HttpClient, HttpParams } from '@angular/common/http'
import { Injectable } from '@angular/core'
import { map, Observable, of } from 'rxjs'
import GeoJSON from 'ol/format/GeoJSON'
import { get as getProj, ProjectionLike } from 'ol/proj'
import { NGXLogger } from 'ngx-logger'
import { FeatureLike } from 'ol/Feature'
import { Link } from '../model/link'
import { environment } from '../../../environments/environment'

export type PointGeoJSON = {
  coordinates: Array<number>
}

export type MultipointGeoJSON = {
  coordinates: Array<Array<number>>
}

export type LinestringGeoJSON = {
  coordinates: Array<Array<number>>
}

export type MultilinestringGeoJSON = {
  coordinates: Array<Array<Array<number>>>
}

export type PolygonGeoJSON = {
  coordinates: Array<Array<Array<number>>>
}

export type MultipolygonGeoJSON = {
  coordinates: Array<Array<Array<Array<number>>>>
}

export type geometrycollectionGeoJSON = {
  geometries: Array<GeometryGeoJSON>
}

export type GeometryGeoJSON =
  | PointGeoJSON
  | MultipointGeoJSON
  | LinestringGeoJSON
  | MultilinestringGeoJSON
  | PolygonGeoJSON
  | MultipolygonGeoJSON
  | geometrycollectionGeoJSON

export type FeatureGeoJSON = {
  type: string
  geometry: GeometryGeoJSON
  id?: string | number
  links?: Array<Link>
  // eslint-disable-next-line @typescript-eslint/no-explicit-any
  properties: Record<string, any> | null
}

export type FeatureCollectionGeoJSON = {
  type: string
  features: Array<FeatureGeoJSON>
  links?: Array<Link>
  numberReturned?: number
}

export type ProjectionMapping = {
  dataProjection: ProjectionLike //Projection of the data we are reading
  visualProjection: ProjectionLike //Projection of the feature geometries created by this function
}

export type DataUrl = {
  url: string
  dataMapping: ProjectionMapping
}

export const defaultMapping: ProjectionMapping = { dataProjection: 'CRS:84', visualProjection: 'EPSG:3857' }

@Injectable({
  providedIn: 'root',
})
export class FeatureService {
  constructor(
    private logger: NGXLogger,
    private http: HttpClient
  ) {}

  queryFeatures(q: string, searchParams: { [key: string]: number }, crs?: string, bbox?: string): Observable<FeatureGeoJSON[]> {
    let params = new HttpParams().set('q', q)
    if (crs) {
      params = params.set('crs', crs)
    }
    if (bbox) {
      params = params.set('bbox', bbox)
    }
    for (const key in searchParams) {
      params = params.append(`${key}[relevance]`, searchParams[key].toString()).append(`${key}[version]`, '1')
    }
    params = params.append('limit', '10')
    return this.http.get<FeatureCollectionGeoJSON>('search', { params }).pipe(map(res => res.features))
  }

  getFeatures(url: DataUrl): Observable<FeatureLike[]> {
    this.logger.debug(JSON.stringify(url))
    const dataproj = getProj(url.dataMapping.dataProjection)!
    this.logger.debug(dataproj.getAxisOrientation()) // Ensure the projection is initialized

    const visualproj = getProj(url.dataMapping.visualProjection)!
    this.logger.debug(visualproj.getAxisOrientation()) // Ensure the visual projection is initialized

    // Helper to swap x/y in coordinates recursively
    // eslint-disable-next-line @typescript-eslint/no-explicit-any
    function swapXYCoords(geom: any): any {
      if (Array.isArray(geom)) {
        if (typeof geom[0] === 'number' && typeof geom[1] === 'number') {
          // Swap [x, y] => [y, x]
          return [geom[1], geom[0], ...geom.slice(2)]
        }
        // eslint-disable-next-line @typescript-eslint/no-explicit-any
        return (geom as any[]).map(swapXYCoords)
      } else if (geom && typeof geom === 'object') {
        if ('coordinates' in geom) {
          return { ...geom, coordinates: swapXYCoords(geom.coordinates) }
        }
        if ('geometries' in geom) {
          return { ...geom, geometries: swapXYCoords(geom.geometries) }
        }
      }
      return geom
    }
    if (url.url == '') return of([])
    return this.http.get<FeatureCollectionGeoJSON | FeatureGeoJSON>(url.url).pipe(
      map(data => {
        let processedData = data
        // make sure swapping works for ANY CRS
        if (dataproj.getAxisOrientation().startsWith('n') || dataproj.getAxisOrientation().startsWith('s')) {
          if (data.type === 'FeatureCollection') {
            // Swap x/y in all features only if axis orientation differs
            const collection = data as FeatureCollectionGeoJSON
            processedData = {
              ...collection,
              features: collection.features?.map(f => ({
                ...f,
                geometry: swapXYCoords(f.geometry),
              })),
            } as FeatureCollectionGeoJSON
          } else {
            const feature = data as FeatureGeoJSON
            processedData = {
              ...feature,
              geometry: swapXYCoords(feature.geometry),
            } as FeatureGeoJSON
          }
        }
        const features = new GeoJSON().readFeatures(processedData, {
          dataProjection: dataproj,
          featureProjection: visualproj,
        })

        return features as FeatureLike[]
      })
    )
  }
  getProjectionMapping(value: string): ProjectionMapping {
    // If no value is passed to the component use CRS84 for data and EPSG:3857 (wgs 84) for rendering
    if (!value) return defaultMapping
    const projection = this.getProjectionCodeFromUrl(value)
    // if bgt background supports the projection, return it. Else default to wgs 84
    if (environment.bgt.projections.includes(projection)) return { dataProjection: projection, visualProjection: projection }
    else return { dataProjection: projection, visualProjection: 'EPSG:3857' }
  }

  private getProjectionCodeFromUrl(value: string): string {
    const EPSG_PREFIX = 'EPSG'
    // get the code from the url
    const code = value.substring(value.lastIndexOf('/') + 1).toLocaleUpperCase()
    // check if the url contains epsg. If so, prefix, otherwise it's CRS or some other XXX00 code. Split letters from numbers and insert ":".
    const captureAlphaNumericGroups = /^([A-Z]+)(\d+)$/
    return value.toLowerCase().includes(EPSG_PREFIX.toLowerCase())
      ? `${EPSG_PREFIX}:${code}`
      : code.replace(captureAlphaNumericGroups, '$1:$2')
  }
}
