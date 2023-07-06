import { HttpClient } from '@angular/common/http';
import { Injectable } from '@angular/core';
import { Circle, Fill, Stroke, Style } from 'ol/style';
import { Observable } from 'rxjs';
import { Feature } from 'ol';
import { LineString, Point, Polygon } from 'ol/geom';

export type LegendItem = {
  name: string,
  geoType: Type
  labelX: number,
  labelY: number
  style: Style[],
  feature: Feature

}

export type LegendCfg = {
  itemHeight: number;
  itemWidth: number,
  iconOfset: number;
  iconWidth: number;
  iconHeight: number;
}

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
  Raster = "raster",
  Symbol = "symbol"
}

export function exhaustiveGuard(_value: never): never {
  throw new Error(`ERROR! Reached forbidden guard function with unexpected value: ${JSON.stringify(_value)}`);
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

  xxgetLayersids(url: string): string[] {
    let ids: string[] = []
    this.getMapboxStyle(url).forEach((style: MapboxStyle) => {
      style.layers.forEach((layer: Layer) => {
        ids.push(layer.id)
      })
    })
    return ids
  }

  getItems(style: MapboxStyle, cfg: LegendCfg): LegendItem[] {
    let names: LegendItem[] = []
    style.layers.forEach((layer: Layer, index) => {
      const y = cfg.itemHeight * index ;
      const feature = this.newFeature(layer.type, cfg, y);
      feature.setProperties({ 'layer': layer['source-layer'] })

      const i: LegendItem = {
        name: layer.id + "/" + layer['source-layer'] + ' ' + layer.type,
        geoType: layer.type,
        feature: feature,
        labelX: cfg.itemWidth * 1.1,
        labelY: y + cfg.itemHeight / 2,
        style: this.defaultStyle()
      }
      names.push(i)
    })
    return names
  }



  newFeature(geoType: Type, cfg: LegendCfg, y: number) {
    {
      const half = cfg.itemHeight/2 
      switch (geoType) {
        case Type.Fill: {
          return new Feature({

            geometry: new Polygon([
              [[cfg.iconOfset, cfg.iconOfset + y], [cfg.iconWidth, cfg.iconOfset + y], [cfg.iconWidth, cfg.iconHeight + y], [cfg.iconOfset, cfg.iconHeight + y], [cfg.iconOfset, cfg.iconOfset + y]]
            ])
          })

        }

        case Type.Circle: {
          return new Feature({
            geometry: new Point([cfg.iconOfset, cfg.iconOfset + y + half]),
          })

        }

        case Type.Raster: {
          return new Feature({
            geometry: new Point([cfg.iconOfset, cfg.iconOfset + y+ half]),
          })

        }

        case Type.Symbol: {
          return new Feature({
            geometry: new Point([cfg.iconOfset, cfg.iconOfset + y + half]),
          })

        }

        case Type.Line: {
          return new Feature({
            geometry: new LineString([[cfg.iconOfset, cfg.iconOfset + y + half], [cfg.iconWidth, cfg.iconOfset + y+half]],),
          })


        } default: {
          exhaustiveGuard(geoType);

        }

      }


    }





  }

  defaultStyle() {
    const fill = new Fill({
      color: 'rgba(255,255,255,0.4)',

    });
    const stroke = new Stroke({
      color: '#3399CC',
      width: 1.25,
    });
    const styles = [
      new Style({
        image: new Circle({
          fill: fill,
          stroke: stroke,
          radius: 5,
        }),
        fill: fill,
        stroke: stroke,
      }),
    ];
    return styles
  }

}