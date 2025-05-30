import { HttpClient } from '@angular/common/http'
import { Injectable } from '@angular/core'
import { Observable } from 'rxjs'
import { Feature } from 'ol'
import { StyleLike } from 'ol/style/Style'
import { NGXLogger } from 'ngx-logger'

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
  metadata?: Metadata
  name: string
  id: string
  sprite: string
  glyphs: string
  layers: Layer[]
  sources: NonNullable<unknown>
}

export interface Metadata {
  'ol:webfonts'?: string
  'gokoala:title-items'?: string
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
  'fill-color'?: StopsPattern | string
  'fill-opacity'?: number
  'line-color'?: StopsPattern | string
  'line-width'?: number
  'fill-outline-color'?: string
  'fill-pattern'?: StopsPattern
  'circle-radius'?: number
  'circle-color'?: StopsPattern | string
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

export interface StopsPattern {
  property: string
  type: string
  stops: Array<string[]>
}

export type MatchPattern = Array<string | string>

export type Pattern = MatchPattern | StopsPattern

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
  constructor(
    private http: HttpClient,
    private logger: NGXLogger
  ) {}

  getMapboxStyle(url: string): Observable<MapboxStyle> {
    return this.http.get<MapboxStyle>(url)
  }

  getLayersids(style: MapboxStyle): string[] {
    const ids: string[] = []
    style.layers.forEach((layer: Layer) => {
      ids.push(layer.id)
    })
    return ids
  }

  removeRasterLayers(style: MapboxStyle): MapboxStyle {
    style.layers = style.layers.filter(layer => layer.type !== LayerType.Raster)
    return style
  }

  isPatternWithStops(paint: string | Pattern | undefined): paint is StopsPattern {
    return (paint as StopsPattern).stops !== undefined
  }

  isPatternWithMatch(paint: string | Pattern | undefined): paint is MatchPattern {
    if (!paint || typeof paint === 'string') {
      return false
    }
    if ((paint as MatchPattern) && (paint as MatchPattern).length > 0 && (paint as MatchPattern)[0] === 'match') {
      return true
    }
    return false
  }

  // eslint-disable-next-line @typescript-eslint/no-unsafe-function-type
  getItems(style: MapboxStyle, titleFunction: Function, customTitlePart: string[], addLayerName: boolean): LegendItem[] {
    const names: LegendItem[] = []
    style.layers.forEach((layer: Layer) => {
      const p: IProperties = extractPropertiesFromFilter({}, layer.filter, this.logger)

      if (layer.layout?.['text-field'] && typeof layer.layout['text-field'] === 'string') {
        const label = layer.layout['text-field'].replace('{', '').replace('}', '')
        p['' + label + ''] = label.substring(0, 6)
        const labelTitle = titleFunction(layer['source-layer'], p, customTitlePart, layer['id'], addLayerName)
        const showLabel = label[0].toUpperCase() + label.substring(1)
        this.pushItem(labelTitle + ' ' + showLabel, layer, names, p)
      } else {
        let title = titleFunction(layer['source-layer'], p, customTitlePart, layer['id'], addLayerName)
        if (addLayerName) {
          this.pushItem(title, layer, names, p)
        }
        let paint: Pattern | undefined = undefined
        paint = layer.paint['circle-color'] as Pattern

        if (layer.type == LayerType.Line) {
          paint = layer.paint['line-color'] as Pattern
        }
        if (layer.type == LayerType.Fill) {
          paint = layer.paint['fill-color'] as Pattern
          if (!paint) {
            paint = layer.paint['fill-pattern'] as Pattern
          }
        }
        if (paint !== undefined) {
          if (paint && this.isPatternWithStops(paint)) {
            paint.stops.forEach(stop => {
              const prop: IProperties = {}
              prop['' + paint.property + ''] = stop[0]
              title = stop[0]
              this.pushItem(title, layer, names, prop)
            })
          } else if (this.isPatternWithMatch(paint)) {
            if (this.isPatternWithMatch(paint)) {
              this.domatch(paint, layer, names)
            } else {
              this.logger.warn('Invalid paint pattern detected:', paint)
            }
          }
        }
      }
    })
    return names.sort((a, b) => a.title.localeCompare(b.title))
  }

  domatch(paint: MatchPattern, layer: Layer, names: LegendItem[]): void {
    this.logger.log('match', paint)
    if (paint.length < 3) {
      this.logger.warn('Invalid match pattern:', paint)
      return
    }

    const matchField = paint[1].slice(1)
    this.logger.log('matchField', matchField)

    let title = ''
    paint.forEach((m: string, i) => {
      if (i < 2) {
        this.logger.log('skip', m)
      } else {
        if (m.startsWith('#')) {
          const prop: IProperties = {}
          if (title !== '') {
            prop[matchField] = title
            this.pushItem(title, layer, names, prop)
            title = ''
          }
        } else {
          title = m + ''
        }
      }
    })
  }

  capitalizeFirstLetter(str: string): string {
    return [...str][0].toUpperCase() + str.slice(1)
  }

  idTitle(layername: string, props: IProperties, customTitlePart: string[], id: string): string {
    return id
  }

  customTitle(layername: string, props: IProperties, customTitlePart: string[], addLayerName: boolean): string {
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
    if (addLayerName) {
      if (title === '') {
        title = layername + ' '
      }
    }
    title = title.trimStart()
    title = title.replace('_', ' ')
    return [...title][0].toUpperCase() + title.slice(1)
  }

  private pushItem(title: string, layer: Layer, names: LegendItem[], properties: IProperties = {}) {
    properties['size'] = '1'
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
}

function extractPropertiesFromFilter(prop: IProperties, filter: Filter, logger: NGXLogger) {
  function traverseFilter(filter: filterval): void {
    if (Array.isArray(filter)) {
      const operator = filter[0]
      const conditions = filter.slice(1)
      if (operator === 'all' || operator === 'any') {
        conditions.forEach(traverseFilter)
      } else {
        if (typeof filter[1] === 'string') {
          const key: string = filter[1]
          if (typeof filter[2] === 'string' || typeof filter[2] === 'number') {
            prop[key] = filter[2]
          } else {
            logger.warn(`Unexpected filter value type for key "${key}":`, filter[2])
          }
        } else {
          logger.warn('Unexpected filter key type:', filter[1])
        }
      }
    }
  }
  traverseFilter(filter)
  return prop
}
