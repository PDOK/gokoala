import { HttpClient } from '@angular/common/http'
import { Injectable } from '@angular/core'
import { map, Observable } from 'rxjs'

import { Feature } from 'ol'
import GeoJSON from 'ol/format/GeoJSON'
import { Geometry } from 'ol/geom'

import { ProjectionLike } from 'ol/proj'
import { FeatureCollectionGeoJSON } from './openapi/model/featureCollectionGeoJSON'

export type DataUrl = {
  url: string
  projection: ProjectionLike
}

@Injectable({
  providedIn: 'root',
})
export class FeatureServiceService {
  constructor(private http: HttpClient) {}

  getFeatures(url: DataUrl): Observable<Feature<Geometry>[]> {
    console.log(url)
    return this.http.get<FeatureCollectionGeoJSON>(url.url).pipe(
      map(data => {
        return new GeoJSON().readFeatures(data, { featureProjection: url.projection })
      })
    )
  }
}
