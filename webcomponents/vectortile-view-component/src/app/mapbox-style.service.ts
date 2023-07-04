import { HttpClient } from '@angular/common/http';
import { Injectable } from '@angular/core';
import { Observable } from 'rxjs';



export interface MapboxStyle {
  version: number;
  name: string;
  id: string;
  sprite: string;
  glyphs: string;
  layers: Layer[];
  sources: {};
}

export interface Layer {
  id: string;
  type: Type;
  paint: Paint;
  source: string;
  "source-layer": string;
}

export interface Paint {
  "fill-color"?: FillPattern | string;
  "fill-opacity"?: number;
  "line-color"?: string;
  "line-width"?: number;
  "fill-outline-color"?: string;
  "fill-pattern"?: FillPattern;
  "circle-radius"?: number;
  "circle-color"?: FillPattern | string;
}

export interface FillPattern {
  property: string;
  type: string;
  stops: Array<string[]>;
}



export enum Type {
  Circle = "circle",
  Fill = "fill",
  Line = "line",
}

@Injectable({
  providedIn: 'root'
})

export class MapboxStyleService {

  constructor(private http: HttpClient) { }


  getMapboxStyle(url: string): Observable<MapboxStyle> {
    return (
      this.http.get<MapboxStyle>(url)
    )
  }
}
