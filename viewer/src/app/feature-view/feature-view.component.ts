import { AfterViewInit, ChangeDetectionStrategy, Component, ElementRef, EventEmitter, Input, OnChanges, Output } from '@angular/core'
import { Feature, Map as OLMap, MapBrowserEvent, Overlay, View } from 'ol'
import { FeatureLike } from 'ol/Feature'
import { PanIntoViewOptions } from 'ol/Overlay'
import { FitOptions } from 'ol/View'
import { Extent, getCenter, getTopLeft } from 'ol/extent'
import { Geometry } from 'ol/geom'
import { fromExtent } from 'ol/geom/Polygon'
import { Group, Tile, Vector as VectorLayer } from 'ol/layer'
import TileLayer from 'ol/layer/Tile'
import { Projection, ProjectionLike } from 'ol/proj'
import { OSM, Vector as VectorSource, WMTS as WMTSSource } from 'ol/source'
import { Circle, Fill, Stroke, Style } from 'ol/style'
import WMTSTileGrid from 'ol/tilegrid/WMTS'
import { take } from 'rxjs/operators'
import { environment } from 'src/environments/environment'
import { NgChanges } from '../vectortile-view/vectortile-view.component'
import { DataUrl, defaultMapping, FeatureServiceService, ProjectionMapping } from '../feature-service.service'
import { projectionSetMercator } from '../mapprojection'
import { boxControl } from './boxcontrol'

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
  @Input() itemsUrl!: string
  private _projection: ProjectionMapping = defaultMapping

  @Input() backgroundMap: BackgroundMap = 'OSM'
  @Input() set projection(value: ProjectionLike) {
    this._projection = this.featureService.getProjectionMapping(value)
  }
  private _showBoundingBoxButton: boolean = true
  @Input() set showBoundingBoxButton(showBox) {
    this._showBoundingBoxButton = coerceBooleanProperty(showBox)
  }
  get showBoundingBoxButton() {
    return this._showBoundingBoxButton
  }
  @Output() box = new EventEmitter<string>()
  @Output() activeFeature = new EventEmitter<FeatureLike>()
  mapHeight = 400
  mapWidth = 600

  map: OLMap = this.getMap()
  features: FeatureLike[] = []

  constructor(
    private el: ElementRef,
    private featureService: FeatureServiceService
  ) {}

  private getMap(): OLMap {
    return new OLMap({
      view: new View({
        projection: this._projection.visualProjection,
      }),
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
        this.map.setLayerGroup(new Group())
        this.loadFeatures(this.features)
        this.loadBackground()
      })
  }

  ngAfterViewInit() {
    if (this._showBoundingBoxButton) {
      this.map.addControl(new boxControl(this.box, {}))
    }
    this.addFeatureEmit()
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
        this.map.addLayer(this.brtLayer(projectionSetMercator()))
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
        style: this.getStyle(),
        zIndex: 10,
      })
    )
    const ext = vsource.getExtent()
    if (features.length > 0) {
      if (features.length < 3) {
        this.setViewExtent(ext, 10)
      } else {
        this.setViewExtent(ext, 1.05)
      }
    }

    return ext
  }

  getStyle() {
    const fill = new Fill({
      color: 'rgba(0,0,255)',
    })
    const stroke = new Stroke({
      color: '#3399CC',
      width: 1.25,
    })
    return [
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
  }

  setViewExtent(extent: Extent, scale: number) {
    const view = new View({})
    const fitOptions: FitOptions = {
      size: this.map.getSize(),
    }
    const geom = fromExtent(extent)
    geom.scale(scale)
    view.fit(geom, fitOptions)
    this.map.setView(view)
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

    this.map.on('pointermove', (evt: MapBrowserEvent<UIEvent>) => {
      this.map.forEachFeatureAtPixel(
        evt.pixel,
        (feature: FeatureLike) => {
          if (feature) {
            const featureId = feature.getId()
            if (featureId) {
              const items = 'items'
              const itemsUrl = this.itemsUrl.toLowerCase()
              const currentUrl = new URL(itemsUrl.substring(0, itemsUrl.indexOf(items) + items.length))
              tooltipContent.innerHTML =
                '<a href="' +
                currentUrl.protocol +
                '//' +
                currentUrl.host +
                currentUrl.pathname +
                '/' +
                featureId +
                '">' +
                featureId +
                '</a>'
              const f = feature.getGeometry()
              if (f) {
                tooltip.setPosition(getCenter(f.getExtent()))
                tooltipContainer.style.visibility = 'visible'
              }
              this.activeFeature.emit(feature)
            }
          }
        },
        { hitTolerance: 3 }
      )
    })
  }
}
