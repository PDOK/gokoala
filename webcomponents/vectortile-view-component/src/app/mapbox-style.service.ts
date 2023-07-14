import { HttpClient } from '@angular/common/http';
import { Injectable } from '@angular/core';
import { Circle, Fill, Stroke, Style } from 'ol/style';
import { Observable } from 'rxjs';
import { Feature } from 'ol';
import { LineString, Point, Polygon } from 'ol/geom';

export interface IProperties {
  [key: string]: string
}

export type LegendItem = {
  sourceLayer: any;
  name: string,
  title: string,
  geoType: LayerType
  labelX: number,
  labelY: number | undefined
  style: Style[],
  feature: Feature | undefined
  properties: IProperties


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
  type: LayerType;
  paint: Paint;
  source: string;
  "source-layer": string;
  filter: Filter;


}


export type Filter = filterval[];
type filterval = string | bigint | filterval[];

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

export interface SpriteData {
  height: number;
  pixelRatio: number;
  width: number;
  x: number;
  y: number;
}




export enum LayerType {
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

  getMapboxSpriteData(url: string): Observable<SpriteData> {
    return (
      this.http.get<SpriteData>(url)
    )
  }

  getLayersids(style: MapboxStyle): string[] {
    let ids: string[] = []
    style.layers.forEach((layer: Layer) => {
      ids.push(layer.id)
    })
    return ids
  }

  removefilters(style: MapboxStyle): MapboxStyle {
    style.layers.forEach((layer: Layer) => {
      layer.filter = []


    })
    return style
  }

  removeRasterLayers(style: MapboxStyle): MapboxStyle {
    style.layers = style.layers.filter(layer => layer.type !== LayerType.Raster)
    return style
  }


  isFillPatternWithStops(paint: string | FillPattern | undefined): paint is FillPattern {
    return (paint as FillPattern).stops !== undefined;
  }


  getItems(style: MapboxStyle, cfg: LegendCfg): LegendItem[] {
    let names: LegendItem[] = []
    style.layers.forEach((layer: Layer) => {
      const title = this.capitalizeFirstLetter(layer['source-layer']);
      //   const title = layer['id'];
      this.PushItem(title, layer, names, cfg, {});
      let paint = layer.paint['circle-color'] as FillPattern
      if (layer.type == LayerType.Fill) {
        paint = layer.paint['fill-color'] as FillPattern
        if (!paint) {
          paint = layer.paint['fill-pattern'] as FillPattern
        }
      }
      if (paint) {
        if (this.isFillPatternWithStops(paint)) {
          paint.stops.forEach(stop => {
            let prop: IProperties = {}
            prop['' + paint.property + ''] = stop[0]
            this.PushItem(stop[0], layer, names, cfg, prop);
          })
        }
      }
    })
    let sorted = names.sort((a, b) => a.title.localeCompare(b.title))
    let modified = sorted.map((x, i) => {
      x.labelY = cfg.itemHeight * i + cfg.itemHeight / 2 - cfg.iconOfset / 2
      x.feature = this.NewFeature(x, cfg, cfg.itemHeight * i)
      x.feature.set('layer', x.sourceLayer)
      x.feature.setProperties(x.properties)
      return x
    })
    return modified
  }

  capitalizeFirstLetter(str: string): string {
    return [...str][0].toUpperCase() + str.slice(1);
  }



  private PushItem(title: string, layer: Layer, names: LegendItem[], cfg: LegendCfg, properties: IProperties = {}) {
    if (!names.find(e => e.title === title)) {
      const i: LegendItem = {
        name: layer.id,
        title: title,
        geoType: layer.type,
        labelX: cfg.itemWidth * 3,
        labelY: undefined,
        style: [],  
        sourceLayer: layer['source-layer'],
        feature: undefined,
        properties: properties
      };
      names.push(i);
    }
  }

  NewFeature(item: LegendItem, cfg: LegendCfg, y: number) {
    {
      const half = cfg.itemHeight / 2
      switch (item.geoType) {
        case LayerType.Fill: {
          return new Feature({

            geometry: new Polygon([
              [[cfg.iconOfset, cfg.iconOfset + y], [cfg.iconWidth, cfg.iconOfset + y], [cfg.iconWidth, cfg.iconHeight + y], [cfg.iconOfset, cfg.iconHeight + y], [cfg.iconOfset, cfg.iconOfset + y]]
            ])
          })
        }

        case LayerType.Circle:
        case LayerType.Raster:
        case LayerType.Symbol: {
          return new Feature({
            geometry: new Point([cfg.iconWidth / 2, cfg.iconOfset + y + half]),
          })

        }

        case LayerType.Line: {
          return new Feature({
            geometry: new LineString([[cfg.iconOfset, cfg.iconOfset + y + half], [cfg.iconWidth, cfg.iconOfset + y + half]],),
          })

        } default: {
          exhaustiveGuard(item.geoType);

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