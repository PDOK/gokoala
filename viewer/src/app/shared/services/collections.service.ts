import { inject, Injectable } from '@angular/core'
import { HttpClient, HttpParams } from '@angular/common/http'
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

  getCollections(): Observable<Collection[]> {
    const params = new HttpParams().set('f', 'json')
    return this._http.get<CollectionResponse>('collections', { params }).pipe(map(res => res.collections))
  }
}
