import { AfterViewInit, ChangeDetectionStrategy, Component, ElementRef, EventEmitter, Input, OnChanges, Output } from '@angular/core'
import { Feature, MapBrowserEvent, Map as OLMap, Overlay, View } from 'ol'
import { FeatureLike } from 'ol/Feature'
import { defaults as defaultControls } from 'ol/control'

import { PanIntoViewOptions } from 'ol/Overlay'
import { FitOptions } from 'ol/View'
import { Extent, getCenter, getTopLeft } from 'ol/extent'
import { Geometry } from 'ol/geom'
import { fromExtent } from 'ol/geom/Polygon'
import { Tile, Vector as VectorLayer } from 'ol/layer'
import TileLayer from 'ol/layer/Tile'
import { Projection } from 'ol/proj'
import { OSM, Vector as VectorSource, WMTS as WMTSSource } from 'ol/source'
import { Circle, Fill, Stroke, Style, Text } from 'ol/style'
import WMTSTileGrid from 'ol/tilegrid/WMTS'
import { take } from 'rxjs/operators'
import { environment } from 'src/environments/environment'
import { DataUrl, FeatureServiceService, ProjectionMapping, defaultMapping } from '../feature-service.service'
import { projectionSetRD } from '../mapprojection'
import { NgChanges } from '../vectortile-view/vectortile-view.component'
import { boxControl, emitBox } from './boxcontrol'
import { fullBoxControl } from './fullboxcontrol'
import { Types as BrowserEventType } from 'ol/MapBrowserEventType'
import { Options as TextOptions } from 'ol/style/Text'
import { getPointResolution, get as getProjection, transform } from 'ol/proj'
import { NGXLogger } from 'ngx-logger'

/** Coerces a data-bound value (typically a string) to a boolean. */
// eslint-disable-next-line @typescript-eslint/no-explicit-any
export function coerceBooleanProperty(value: any): boolean {
  return value != null && `${value}` !== 'false'
}

export function exhaustiveGuard(_value: never): never {
  throw new Error(`ERROR! Reached forbidden guard function with unexpected value: ${JSON.stringify(_value)}`)
}

export type BackgroundMap = 'BRT' | 'OSM'

@Component({
  selector: 'app-feature-view',
  templateUrl: './feature-view.component.html',
  styleUrls: ['./feature-view.component.css'],
  changeDetection: ChangeDetectionStrategy.OnPush,
  standalone: true,
})
export class FeatureViewComponent implements OnChanges, AfterViewInit {
  private _showBoundingBoxButton: boolean = true
  initial: boolean = true

  @Input() set showBoundingBoxButton(showBox) {
    this._showBoundingBoxButton = coerceBooleanProperty(showBox)
    this.showButtons()
  }
  get showBoundingBoxButton() {
    return this._showBoundingBoxButton
  }
  private _showFillExtentButton: boolean = false
  @Input() set showFillExtentButton(showBox) {
    this._showFillExtentButton = coerceBooleanProperty(showBox)
    this.showButtons()
  }
  get showFillExtentButton() {
    return this._showFillExtentButton
  }
  @Input() itemsUrl!: string
  private _projection: ProjectionMapping = defaultMapping

  @Input() backgroundMap: BackgroundMap = 'OSM'
  @Input() minFitResolution: number = 0.1
  @Input() maxFitResolution: number | undefined = undefined
  @Input() fillColor: string = 'rgba(0,0,255)'
  @Input() strokeColor: string = '#3399CC'
  @Input() mode: 'default' | 'auto' = 'default'

  @Input() labelField = undefined
  @Input() labelOptions: string | undefined = undefined

  @Input() set projection(value: string) {
    this._projection = this.featureService.getProjectionMapping(value)
  }

  @Output() box = new EventEmitter<string>()
  @Output() activeFeature = new EventEmitter<FeatureLike>()
  mapHeight = 400
  mapWidth = 600

  map: OLMap = this.getMap()
  private _view: View = new View({
    projection: this._projection.visualProjection,
    zoom: 1,
  })
  features: FeatureLike[] = []

  constructor(
    private el: ElementRef,
    private featureService: FeatureServiceService,
    private logger: NGXLogger
  ) {}

  private getMap(): OLMap {
    return new OLMap({
      view: this._view,

      controls: [],
    })
  }

  private init() {
    this.mapWidth = this.el.nativeElement.offsetWidth * 0.99
    this.mapHeight = this.mapWidth * 0.75 // height = 0.75 * width creates 4:3 aspect ratio
    const mapElm: HTMLElement = this.el.nativeElement.querySelector('#featuremap')
    this.map.setTarget(mapElm)
    const featuresUrl: DataUrl = { url: this.itemsUrl, dataMapping: this._projection }
    this.featureService
      .getFeatures(featuresUrl)
      .pipe(take(1))
      .subscribe(data => {
        this.features = data
        //this.map.setLayerGroup(new Group())
        this.changeView()
        this.loadFeatures(this.features)
        this.loadBackground()
        this.logger.log(this.map.getView().getProjection())
        this.logger.log('resolution' + this.map.getView().getResolution())
      })
  }

  ngAfterViewInit() {
    this.changeMode()
  }

  changeMode() {
    this.features = []
    this.showButtons()
    this.addFeatureEmit()
  }

  showButtons() {
    this.map.getControls().forEach(x => {
      this.map.removeControl(x)
    })
    defaultControls({
      attribution: false,
      zoom: true,
    }).forEach(x => this.map.addControl(x))
    if (this.mode === 'default') {
      if (this._showBoundingBoxButton) {
        this.map.addControl(new boxControl(this.box, {}))
      }
      if (this._showFillExtentButton) {
        this.map.addControl(new fullBoxControl(this.box, {}))
      }
    }
  }

  ngOnChanges(changes: NgChanges<FeatureViewComponent>) {
    if (
      changes.itemsUrl?.previousValue !== changes.itemsUrl?.currentValue ||
      changes.projection?.previousValue !== changes.projection?.currentValue
    ) {
      if (changes.itemsUrl?.currentValue) {
        this.init()
      }
    }

    if (changes.mode?.previousValue && changes.mode?.previousValue != changes.mode?.currentValue) {
      if (changes.mode?.currentValue) {
        this.changeMode()
      }
    }
  }

  loadBackground() {
    switch (this.backgroundMap) {
      case 'OSM': {
        const osm = new TileLayer({
          source: new OSM(),
        })
        this.map.addLayer(osm)
        return
      }
      case 'BRT': {
        this.map.addLayer(this.brtLayer(projectionSetRD()))
        return
      }

      default: {
        exhaustiveGuard(this.backgroundMap)
      }
    }
  }

  loadFeatures(features: FeatureLike[]) {
    const vsource = new VectorSource({
      features: features,
    }) as VectorSource<Feature<Geometry>>

    this.map.addLayer(
      new VectorLayer({
        source: vsource,
        style: feature => this.getStyle(feature),
        zIndex: 10,
      })
    )
    const extent = vsource.getExtent()
    if (this.mode === 'default') {
      if (features.length > 0) {
        if (features.length < 3) {
          this.setViewExtent(extent, 10)
        } else {
          this.setViewExtent(extent, 1.05)
        }
      }
    } else {
      if (this.initial) {
        this.setViewExtent(extent, 1)
        this.initial = false
      }
    }
  }

  getStyle(feature: FeatureLike) {
    const fill = new Fill({
      color: this.fillColor,
    })
    const stroke = new Stroke({
      color: this.strokeColor,
      width: 1.25,
    })

    let text = undefined
    if (this.labelField) {
      if (this.labelOptions != undefined) {
        const opt = JSON.parse(this.labelOptions) as TextOptions
        text = new Text(opt)
        text.setText(feature.get(this.labelField))
      } else {
        text = new Text({ text: feature.get(this.labelField) })
      }
    }

    return [
      new Style({
        image: new Circle({
          fill: fill,
          stroke: stroke,
          radius: 5,
        }),
        fill: fill,
        stroke: stroke,
        text: text,
      }),
    ]
  }

  changeView() {
    const currentProjection = this._view.getProjection()
    const newProjection = getProjection(this._projection.visualProjection)
    if (!newProjection) {
      throw new Error('Nieuwe projectie is niet gedefinieerd')
    }
    const throwUndefinedError = (msg: string) => {
      throw new Error(msg)
    }
    const currentResolution = this._view.getResolution() ?? throwUndefinedError('Huidige resolutie is niet gedefinieerd')
    const currentCenter = this._view.getCenter() ?? [0, 0]
    const newCenter = transform(currentCenter, currentProjection, newProjection) ?? [0, 0]
    const currentRotation = this._view.getRotation() ?? 0
    const currentMPU = currentProjection?.getMetersPerUnit() ?? throwUndefinedError('Huidige MPU is niet gedefinieerd')
    const newMPU = newProjection?.getMetersPerUnit() ?? throwUndefinedError('Nieuwe MPU is niet gedefinieerd')
    const currentPointResolution = getPointResolution(currentProjection, 1 / currentMPU, currentCenter, 'm') * currentMPU
    const newPointResolution = getPointResolution(newProjection, 1 / newMPU, newCenter, 'm') * newMPU
    const newResolution = (currentResolution * currentPointResolution) / newPointResolution

    this._view = new View({
      center: newCenter,
      resolution: newResolution,
      rotation: currentRotation,
      projection: newProjection,
    })

    this.map.setView(this._view)
  }

  setViewExtent(extent: Extent, scale: number) {
    const fitOptions: FitOptions = {
      size: this.map.getSize(),
      minResolution: this.minFitResolution,
    }

    const geom = fromExtent(extent)
    geom.scale(scale)
    this._view.fit(geom, fitOptions)
    if (this.maxFitResolution) {
      const res = Math.min(this._view.getResolution()!, this.maxFitResolution)
      this._view.setResolution(res)
    }
    this.map.setView(this._view)
  }

  brtLayer(p: { projection: Projection; resolutions: number[]; matrixIds: string[] }) {
    return new Tile({
      source: new WMTSSource({
        attributions: 'Kaartgegevens: &copy; <a href="https://www.kadaster.nl">Kadaster</a>',
        url: environment.bgtBackgroundUrl,
        layer: 'grijs',
        matrixSet: p.projection.getCode(),
        format: 'image/png',
        projection: p.projection,
        tileGrid: new WMTSTileGrid({
          origin: getTopLeft(p.projection.getExtent()),
          resolutions: p.resolutions,
          matrixIds: p.matrixIds,
        }),
        style: 'default',
        wrapX: false,
      }),
    })
  }

  addFeatureEmit() {
    const tooltipContainer = this.el.nativeElement.querySelector("[id='tooltip']")
    tooltipContainer.style.visibility = 'hidden'
    const tooltipContent = this.el.nativeElement.querySelector("[id='tooltip-content']")
    const tooltip = new Overlay({
      element: tooltipContainer,
      autoPan: {
        duration: 250,
      } as PanIntoViewOptions,
    })

    this.map.addOverlay(tooltip)
    let eventType: BrowserEventType = 'pointermove'
    if (this.labelField) {
      eventType = 'click'
    }
    this.map.on(eventType, (evt: MapBrowserEvent<UIEvent>) => {
      this.map.forEachFeatureAtPixel(
        evt.pixel,
        (feature: FeatureLike) => {
          if (feature) {
            const featureId = feature.getId()
            if (featureId) {
              const items = 'items'
              const itemsUrl = this.itemsUrl.toLowerCase()
              const currentUrl = new URL(itemsUrl.substring(0, itemsUrl.indexOf(items) + items.length))
              const link = currentUrl.protocol + '//' + currentUrl.host + currentUrl.pathname + '/' + featureId
              tooltipContent.innerHTML = '<a href="' + link + '">' + featureId + '</a>'
              const f = feature.getGeometry()
              if (f) {
                tooltip.setPosition(getCenter(f.getExtent()))

                if (!this.labelField) {
                  tooltipContainer.style.visibility = 'visible'
                }
              }
              this.activeFeature.emit(feature)
            }
          }
        },
        { hitTolerance: 3 }
      )
    })

    this.map.on('moveend', () => {
      const size = this.map.getSize()
      const extent = this.map.getView().calculateExtent(size)
      const polygon = fromExtent(extent) as Geometry
      if (this.mode === 'auto') {
        emitBox(this.map, polygon, this.box)
      }
    })
  }
}
