import { Injectable } from '@angular/core';
import { HttpClient } from '@angular/common/http';
import { Observable } from 'rxjs';
export interface MatrixSet {
  links: Link[];
  id: string;
  title: string;
  crs: string;
  wellKnownScaleSet: string;
  tileMatrices: TileMatrix[];
  orderedAxes: string[];
}

export interface Link {
  rel: string;
  type: string;
  title: string;
  href: string;
}

export interface TileMatrix {
  id: number;
  tileWidth: number;
  tileHeight: number;
  matrixWidth: number;
  matrixHeight: number;
  scaleDenominator: number;
  cellSize: number;
  pointOfOrigin: number[];
}


@Injectable({
  providedIn: 'root'
})
export class MatrixsetService {
  constructor(private http: HttpClient) { }
  getMatrixSet(url: string): Observable<MatrixSet> {
    return (
      this.http.get<MatrixSet>(url)
    )
  }

  

}
