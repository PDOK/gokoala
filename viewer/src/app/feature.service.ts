import { HttpClient } from '@angular/common/http'
import { Injectable } from '@angular/core'
import { catchError, map, Observable, of } from 'rxjs'
import GeoJSON from 'ol/format/GeoJSON'
import { ProjectionLike } from 'ol/proj'
import { NGXLogger } from 'ngx-logger'
import { initProj4 } from './map-projection'
import { FeatureLike } from 'ol/Feature'
import { Link } from './link'
import { get as getProj } from 'ol/proj'

export type pointGeoJSON = {
  coordinates: Array<number>
}

export type multipointGeoJSON = {
  coordinates: Array<Array<number>>
}

export type linestringGeoJSON = {
  coordinates: Array<Array<number>>
}

export type multilinestringGeoJSON = {
  coordinates: Array<Array<Array<number>>>
}

export type polygonGeoJSON = {
  coordinates: Array<Array<Array<number>>>
}

export type multipolygonGeoJSON = {
  coordinates: Array<Array<Array<Array<number>>>>
}

export type geometrycollectionGeoJSON = {
  geometries: Array<geometryGeoJSON>
}

export type geometryGeoJSON =
  | pointGeoJSON
  | multipointGeoJSON
  | linestringGeoJSON
  | multilinestringGeoJSON
  | polygonGeoJSON
  | multipolygonGeoJSON
  | geometrycollectionGeoJSON

export type featureGeoJSON = {
  geometry: geometryGeoJSON
  id?: string | number
  links?: Array<Link>
  // eslint-disable-next-line @typescript-eslint/no-explicit-any
  properties: Record<string, any> | null
}

export type featureCollectionGeoJSON = {
  features: Array<featureGeoJSON>
  links?: Array<Link>
  numberReturned?: number
}

export type ProjectionMapping = {
  dataProjection: ProjectionLike //Projection of the data we are reading
  visualProjection: ProjectionLike //Projection of the feature geometries created by this function
  swapAllowed: boolean
}

export type DataUrl = {
  url: string
  dataMapping: ProjectionMapping
}

export const defaultMapping: ProjectionMapping = { dataProjection: 'EPSG:4326', visualProjection: 'EPSG:3857', swapAllowed: false }

@Injectable({
  providedIn: 'root',
})
export class FeatureService {
  constructor(
    private logger: NGXLogger,
    private http: HttpClient
  ) {}

  getFeatures(url: DataUrl): Observable<FeatureLike[]> {
    this.logger.debug('Getfeatures')
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

    return this.http.get<featureCollectionGeoJSON>(url.url).pipe(
      map(data => {
        let processedData = data
        if (url.dataMapping.swapAllowed && dataproj.getAxisOrientation() !== visualproj.getAxisOrientation()) {
          // Swap x/y in all features only if axis orientation differs
          processedData = {
            ...data,
            features: data.features.map(f => ({
              ...f,
              geometry: swapXYCoords(f.geometry),
            })),
          }
        }
        const features = new GeoJSON().readFeatures(processedData, {
          dataProjection: dataproj,
          featureProjection: visualproj,
        })
      }),
      catchError(error => {
        this.logger.error('Error fetching features:', error)
        return of([])
      })
    )
  }

  getProjectionMapping(value: string = 'http://www.opengis.net/def/crs/OGC/1.3/CRS84'): ProjectionMapping {
    initProj4()

    if (value) {
      if (value.substring(value.lastIndexOf('/') + 1).toLocaleUpperCase() === 'CRS84') {
        //'EPSG:3857' Default the map is in Web Mercator(EPSG: 3857), the actual coordinates used are in lat-long (EPSG: 4326)
        return defaultMapping
      }
      if (value.toLowerCase().startsWith('http://www.opengis.net/def/crs/epsg/')) {
        const projection = 'EPSG:' + value.substring(value.lastIndexOf('/') + 1)
        if (projection === 'EPSG:3035' || projection === 'EPSG:4258') {
          return { dataProjection: projection, visualProjection: 'EPSG:3857', swapAllowed: true }
        } else return { dataProjection: projection, visualProjection: projection, swapAllowed: true }
      }
      return { dataProjection: value, visualProjection: value, swapAllowed: true }
    }
    return { dataProjection: value, visualProjection: value, swapAllowed: true }
  }
}
