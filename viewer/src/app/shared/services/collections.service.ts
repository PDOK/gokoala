import { inject, Injectable } from '@angular/core'
import { HttpClient } from '@angular/common/http'
import { map, Observable } from 'rxjs'
import { environment } from '../../../environments/environment'
import { Link } from '../../link'

export interface Collection {
  id: string
  title: string
  version: number
  displayNameTemplate: string
  links: Link[]
}

export interface CollectionResponse {
  links: Link[]
  collections: Collection[]
}

@Injectable({
  providedIn: 'root',
})
export class CollectionsService {
  private _http = inject(HttpClient)

  getCollections(): Observable<Collection[]> {
    const url = new URL('collections', environment.locationApi)
    url.searchParams.append('f', 'json')
    return this._http.get<CollectionResponse>(url.toString()).pipe(map(res => res.collections))
  }
}
