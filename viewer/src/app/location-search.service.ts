import { HttpClient } from '@angular/common/http'
import { Injectable } from '@angular/core'
import { NGXLogger } from 'ngx-logger'
import { GeoJSON } from 'ol/format'
import { map as rxjsmap, Observable } from 'rxjs'
import { DataUrl, featureCollectionGeoJSON } from './feature.service'
import { Feature } from 'ol'
import { Geometry } from 'ol/geom'

export interface Search {
  type: string
  timeStamp: Date
  links: Link[]
  features: SearchFeature[]
  numberReturned: number
}

export interface SearchFeature {
  type: string
  properties: SearchProperties
  geometry: SearchGeometry
  id: string
  links: Link[]
}

export interface SearchGeometry {
  type: string
  coordinates: Array<Array<number[]>>
}

export interface Link {
  rel: string
  title: string
  type: string
  href: string
}

export interface SearchProperties {
  collectionGeometryType: string
  collectionId: string
  collectionVersion: string
  displayName: string
  highlight: string
  href: string
  score: number
}

@Injectable({
  providedIn: 'root',
})
export class LocationSearchService {
  constructor(
    private logger: NGXLogger,
    private http: HttpClient
  ) {}

  getResults(url: DataUrl): Observable<Feature<Geometry>[]> {
    this.logger.log(JSON.stringify(url))
    return this.http.get<featureCollectionGeoJSON>(url.url).pipe(
      rxjsmap(data => {
        return new GeoJSON().readFeatures(data, {
          dataProjection: url.dataMapping.dataProjection,
          featureProjection: url.dataMapping.visualProjection,
        })
      })
    )
  }
}
