import { inject, Injectable } from '@angular/core'
import { HttpClient } from '@angular/common/http'
import { map, Observable } from 'rxjs'
import { Link } from '../../link'

export interface Collection {
  id: string
  title: string
  version: number
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

  getCollections(locationApiUrl: string | URL | undefined): Observable<Collection[]> {
    const url = new URL('collections', locationApiUrl)
    url.searchParams.append('f', 'json')
    return this._http.get<CollectionResponse>(url.toString()).pipe(map(res => res.collections))
  }
}
