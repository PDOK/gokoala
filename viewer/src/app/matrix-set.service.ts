import { Injectable } from '@angular/core'
import { HttpClient } from '@angular/common/http'
import { Observable } from 'rxjs'
import { Link } from './link'

export interface MatrixSet {
  links: Link[]
  id: string
  title: string
  crs: string
  wellKnownScaleSet: string
  tileMatrices: TileMatrix[]
  orderedAxes: string[]
}

export interface TileMatrix {
  id: number
  tileWidth: number
  tileHeight: number
  matrixWidth: number
  matrixHeight: number
  scaleDenominator: number
  cellSize: number
  pointOfOrigin: number[]
}

export interface Matrix {
  title: string
  links: Link[]
  crs: string
  dataType: string
  tileMatrixSetId: string
  tileMatrixSetLimits: TileMatrixSetLimit[]
}

export interface TileMatrixSetLimit {
  tileMatrix: string
  minTileRow: number
  maxTileRow: number
  minTileCol: number
  maxTileCol: number
}

@Injectable({
  providedIn: 'root',
})
export class MatrixSetService {
  constructor(private http: HttpClient) {}
  getMatrixSet(url: string): Observable<MatrixSet> {
    return this.http.get<MatrixSet>(url)
  }

  getMatrix(url: string): Observable<Matrix> {
    return this.http.get<Matrix>(url)
  }
}
