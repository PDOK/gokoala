import { HttpClient } from '@angular/common/http'
import { Injectable } from '@angular/core'
import { catchError, map, Observable, of } from 'rxjs'
import GeoJSON from 'ol/format/GeoJSON'
import { ProjectionLike } from 'ol/proj'
import { NGXLogger } from 'ngx-logger'
import { initProj4 } from './map-projection'
import { FeatureLike } from 'ol/Feature'
import { Link } from './link'

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
}

export type DataUrl = {
  url: string
  dataMapping: ProjectionMapping
}

export const defaultMapping: ProjectionMapping = { dataProjection: 'EPSG:4326', visualProjection: 'EPSG:3857' }

@Injectable({
  providedIn: 'root',
})
export class FeatureService {
  constructor(
    private logger: NGXLogger,
    private http: HttpClient
  ) {}

  getFeatures(url: DataUrl): Observable<FeatureLike[]> {
    this.logger.log(JSON.stringify(url))
    return this.http.get<featureCollectionGeoJSON>(url.url).pipe(
      map(data => {
        return new GeoJSON().readFeatures(data, {
          dataProjection: url.dataMapping.dataProjection,
          featureProjection: url.dataMapping.visualProjection,
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
          return { dataProjection: projection, visualProjection: 'EPSG:3857' }
        } else return { dataProjection: projection, visualProjection: projection }
      }
      return { dataProjection: value, visualProjection: value }
    }
    return { dataProjection: value, visualProjection: value }
  }
}
