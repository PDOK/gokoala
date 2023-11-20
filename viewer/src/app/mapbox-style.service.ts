import { HttpClient } from '@angular/common/http'
import { Injectable } from '@angular/core'
import { Circle, Fill, Stroke, Style } from 'ol/style'
import { Observable } from 'rxjs'
import { Feature } from 'ol'
import { StyleLike } from 'ol/style/Style'

export interface IProperties {
  [key: string]: string
}

export type LegendItem = {
  sourceLayer: unknown
  name: string
  title: string
  geoType: LayerType
  style: StyleLike
  feature: Feature | undefined
  properties: IProperties
}

export interface MapboxStyle {
  version: number
  name: string
  id: string
  sprite: string
  glyphs: string
  layers: Layer[]
  sources: NonNullable<unknown>
}

export interface Layer {
  filterCopy: Filter
  id: string
  type: LayerType
  paint: Paint
  source: string
  layout?: Layout
  'source-layer': string
  filter: Filter
}

export type Filter = filterval[]
type filterval = string | bigint | filterval[]

export interface Paint {
  'fill-color'?: FillPattern | string
  'fill-opacity'?: number
  'line-color'?: string
  'line-width'?: number
  'fill-outline-color'?: string
  'fill-pattern'?: FillPattern
  'circle-radius'?: number
  'circle-color'?: FillPattern | string
}

export enum Line {
  Round = 'round',
}

export interface Layout {
  visibility?: string
  'line-join'?: Line
  'line-cap'?: Line
  'text-field'?: string
  'text-size'?: number
  'text-font'?: string[]
  'symbol-placement'?: LayerType
  'icon-image'?: string
  'icon-size'?: number
  'text-offset'?: number[]
}

export interface FillPattern {
  property: string
  type: string
  stops: Array<string[]>
}

export interface SpriteData {
  height: number
  pixelRatio: number
  width: number
  x: number
  y: number
}

export enum LayerType {
  Circle = 'circle',
  Fill = 'fill',
  Line = 'line',
  Raster = 'raster',
  Symbol = 'symbol',
}

export function exhaustiveGuard(_value: never): never {
  throw new Error(`ERROR! Reached forbidden guard function with unexpected value: ${JSON.stringify(_value)}`)
}

@Injectable({
  providedIn: 'root',
})
export class MapboxStyleService {
  constructor(private http: HttpClient) {}

  getMapboxStyle(url: string): Observable<MapboxStyle> {
    return this.http.get<MapboxStyle>(url)
  }

  getMapboxSpriteData(url: string): Observable<SpriteData> {
    return this.http.get<SpriteData>(url)
  }

  getLayersids(style: MapboxStyle): string[] {
    const ids: string[] = []
    style.layers.forEach((layer: Layer) => {
      ids.push(layer.id)
    })
    return ids
  }

  removefilters(style: MapboxStyle): MapboxStyle {
    style.layers.forEach((layer: Layer) => {
      layer.filterCopy = layer.filter
      layer.filter = []
    })
    return style
  }

  removeRasterLayers(style: MapboxStyle): MapboxStyle {
    style.layers = style.layers.filter(layer => layer.type !== LayerType.Raster)
    return style
  }

  isFillPatternWithStops(paint: string | FillPattern | undefined): paint is FillPattern {
    return (paint as FillPattern).stops !== undefined
  }

  getItems(
    style: MapboxStyle,
    // eslint-disable-next-line @typescript-eslint/ban-types
    titleFunction: Function,
    customTitlePart: string[]
  ): LegendItem[] {
    const names: LegendItem[] = []
    style.layers.forEach((layer: Layer) => {
      const p: IProperties = extractPropertiesFromFilter({}, layer.filter)

      if (layer.layout?.['text-field']) {
        const label = layer.layout?.['text-field'].replace('{', '').replace('}', '')
        p['' + label + ''] = label.substring(0, 6)
      }
      let title = titleFunction(layer['source-layer'], p, customTitlePart)
      this.pushItem(title, layer, names, p)

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
            const prop: IProperties = {}
            prop['' + paint.property + ''] = stop[0]
            title = stop[0]
            this.pushItem(title, layer, names, prop)
          })
        }
      }
    })
    return names.sort((a, b) => a.title.localeCompare(b.title))
  }

  capitalizeFirstLetter(str: string): string {
    return [...str][0].toUpperCase() + str.slice(1)
  }

  customTitle(layername: string, props: IProperties, customTitlePart: string[]): string {
    function gettext(intitle: string, index: string): string {
      if (props[index]) {
        return intitle + ' ' + props[index]
      } else {
        return intitle
      }
    }
    let title = ''
    customTitlePart.forEach(element => {
      title = gettext(title, element)
    })
    if (title === '') {
      title = layername + ' '
    }
    title = title.trimStart()
    title = title.replace('_', ' ')
    return [...title][0].toUpperCase() + title.slice(1)
  }

  private pushItem(title: string, layer: Layer, names: LegendItem[], properties: IProperties = {}) {
    if (!names.find(e => e.title === title)) {
      const i: LegendItem = {
        name: layer.id,
        title: title,
        geoType: layer.type,
        style: [],
        sourceLayer: layer['source-layer'],
        feature: undefined,
        properties: properties,
      }
      names.push(i)
    }
  }

  defaultStyle() {
    const fill = new Fill({
      color: 'rgba(255,255,255,0.4)',
    })
    const stroke = new Stroke({
      color: '#3399CC',
      width: 1.25,
    })
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
    ]
    return styles
  }
}

function extractPropertiesFromFilter(prop: IProperties, filter: Filter) {
  function traverseFilter(filter: filterval) {
    if (Array.isArray(filter)) {
      const operator = filter[0]
      const conditions = filter.slice(1)
      if (operator === 'all' || operator === 'any') {
        conditions.forEach(i => traverseFilter(i))
      } else {
        if (typeof filter[1] === 'string' && typeof filter[2] === 'string') {
          const key: string = filter[1]

          prop[key] = filter[2]
        }
        if (typeof filter[1] === 'string' && typeof filter[2] === 'number') {
          const key: string = filter[1]

          prop[key] = filter[2]
        }
      }
    }
  }
  traverseFilter(filter)
  return prop
}
