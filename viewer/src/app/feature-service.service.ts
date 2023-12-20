import { HttpClient } from '@angular/common/http'
import { Injectable } from '@angular/core'
import { map, Observable } from 'rxjs'

import { Feature } from 'ol'
import GeoJSON from 'ol/format/GeoJSON'
import { Geometry } from 'ol/geom'

import { ProjectionLike } from 'ol/proj'

import { NGXLogger } from 'ngx-logger'
import { initProj4 } from './mapprojection'

export type link = {
  /**
   * Supplies the URI to a remote resource (or resource fragment).
   */
  href: string
  /**
   * A hint indicating what the language of the result of dereferencing the link should be.
   */
  hreflang?: string
  length?: number
  /**
   * The type or semantics of the relation.
   */
  rel: string
  /**
   * Use `true` if the `href` property contains a URI template with variables that needs to be substituted by values to get a URI
   */
  templated?: boolean
  /**
   * Used to label the destination of a link such that it can be used as a human-readable identifier.
   */
  title?: string
  /**
   * A hint indicating what the media type of the result of dereferencing the link should be.
   */
  type?: string
  /**
   * Without this parameter you should repeat a link for each media type the resource is offered.
   * Adding this parameter allows listing alternative media types that you can use for this resource. The value in the `type` parameter becomes the recommended media type.
   */
  types?: Array<string>
  /**
   * A base path to retrieve semantic information about the variables used in URL template.
   */
  varBase?: string
}

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
  links?: Array<link>
  // eslint-disable-next-line @typescript-eslint/no-explicit-any
  properties: Record<string, any> | null
}

export type featureCollectionGeoJSON = {
  features: Array<featureGeoJSON>
  links?: Array<link>
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
export class FeatureServiceService {
  constructor(
    private logger: NGXLogger,
    private http: HttpClient
  ) {}

  getFeatures(url: DataUrl): Observable<Feature<Geometry>[]> {
    this.logger.log(JSON.stringify(url))
    return this.http.get<featureCollectionGeoJSON>(url.url).pipe(
      map(data => {
        return new GeoJSON().readFeatures(data, {
          dataProjection: url.dataMapping.dataProjection,
          featureProjection: url.dataMapping.visualProjection,
        })
      })
    )
  }

  getProjectionMapping(value: ProjectionLike = 'http://www.opengis.net/def/crs/OGC/1.3/CRS84'): ProjectionMapping {
    initProj4()

    if (value) {
      if (typeof value === 'string') {
        if (value.substring(value.lastIndexOf('/') + 1).toLocaleUpperCase() === 'CRS84') {
          //'EPSG:3857' Default the map is in Web Mercator(EPSG: 3857), the actual coordinates used are in lat-long (EPSG: 4326)
          return defaultMapping
        }
        if (value.toLowerCase().startsWith('http://www.opengis.net/def/crs/epsg/')) {
          const projection = 'EPSG:' + value.substring(value.lastIndexOf('/') + 1)
          return { dataProjection: projection, visualProjection: 'EPSG:3857' }
        }
        return { dataProjection: value, visualProjection: value }
      } else {
        this.logger.error('wrong value:')
        this.logger.error(value)
        return value as unknown as ProjectionMapping
      }
    }
    return { dataProjection: value, visualProjection: value }
  }
}
