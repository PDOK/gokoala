import { AfterViewInit, ChangeDetectionStrategy, Component, ElementRef, EventEmitter, Input, OnChanges, Output } from '@angular/core'
import { Feature, MapBrowserEvent, Map as OLMap, Overlay, View } from 'ol'
import { FeatureLike } from 'ol/Feature'
import { PanIntoViewOptions } from 'ol/Overlay'
import { FitOptions } from 'ol/View'
import { platformModifierKeyOnly } from 'ol/events/condition'
import { Extent, getCenter, getTopLeft } from 'ol/extent'
import { Geometry } from 'ol/geom'
import { fromExtent } from 'ol/geom/Polygon'
import { DragBox } from 'ol/interaction'
import { Group, Tile, Vector as VectorLayer } from 'ol/layer'
import TileLayer from 'ol/layer/Tile'
import { Projection, ProjectionLike } from 'ol/proj'
import { OSM, Vector as VectorSource, WMTS as WMTSSource } from 'ol/source'
import { Circle, Fill, Stroke, Style } from 'ol/style'
import WMTSTileGrid from 'ol/tilegrid/WMTS'
import { take } from 'rxjs/operators'
import { environment } from 'src/environments/environment'
import { NgChanges } from '../app.component'
import { DataUrl, FeatureServiceService, featureCollectionGeoJSON } from '../feature-service.service'
import { boxControl } from './boxcontrol'

export function exhaustiveGuard(_value: never): never {
  throw new Error(`ERROR! Reached forbidden guard function with unexpected value: ${JSON.stringify(_value)}`)
}

@Component({
  selector: 'app-feature-view',
  templateUrl: './feature-view.component.html',
  styleUrls: ['./feature-view.component.css'],
  changeDetection: ChangeDetectionStrategy.OnPush,
  standalone: true,
})
export class FeatureViewComponent implements OnChanges, AfterViewInit {
  @Input() itemsUrl!: string
  @Input() projection: ProjectionLike = 'EPSG:3857' //Default the map is in Web Mercator(EPSG: 3857), the actual coordinates used are in lat-long (EPSG: 4326)
  @Input() backgroundMap: 'BRT' | 'OSM' = 'OSM'
  @Output() box = new EventEmitter<string>()
  @Output() activeFeature = new EventEmitter<FeatureLike>()
  mapHeight = 400
  mapWidth = 600
  map: OLMap = this.getMap()
  featureCollectionGeoJSON!: featureCollectionGeoJSON
  features: Feature<Geometry>[] = []
  boxLayer!: VectorLayer<VectorSource<Geometry>>

  constructor(
    private el: ElementRef,
    private featureService: FeatureServiceService
  ) {}

  private getMap(): OLMap {
    return new OLMap({
      view: new View({
        projection: this.projection,
      }),
    })
  }

  private init() {
    this.mapWidth = this.el.nativeElement.offsetWidth * 0.99
    this.mapHeight = this.mapWidth * 0.75 // height = 0.75 * width creates 4:3 aspect ratio
    const mapdiv: HTMLElement = this.el.nativeElement.querySelector('#featuremap')
    this.map.setTarget(mapdiv)
    const aurl: DataUrl = { url: this.itemsUrl, projection: this.projection }
    this.featureService
      .getFeatures(aurl)
      .pipe(take(1))
      .subscribe(data => {
        this.map.setLayerGroup(new Group())
        this.features = data
        this.loadFeatures(this.features)
        this.loadBackground()
      })
  }

  ngAfterViewInit() {
    this.map.addControl(new boxControl(this.box, {}))
    this.addFeatureEmit()
  }

  ngOnChanges(changes: NgChanges<FeatureViewComponent>) {
    if (
      changes.itemsUrl?.previousValue !== changes.itemsUrl?.currentValue ||
      changes.projection.previousValue !== changes.projection.currentValue
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
        this.map.addLayer(this.brtLayer())
        return
      }

      default: {
        exhaustiveGuard(this.backgroundMap)
        return
      }
    }
  }

  loadFeatures(features: Feature<Geometry>[]) {
    const vsource = new VectorSource({
      //features: new GeoJSON().readFeatures(this.featureCollectionGeoJSON, { featureProjection: this.projection  }),
      features: features,
    })

    this.map.addLayer(
      new VectorLayer({
        source: vsource,
        style: this.getStyle(),
        zIndex: 10,
      })
    )
    const ext = vsource.getExtent()
    if (features.length < 3) {
      this.setViewExtent(ext, 10)
    } else {
      this.setViewExtent(ext, 1.05)
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

  brtLayer() {
    const projectionExtent = [-285401.92, 22598.08, 595401.9199999999, 903401.9199999999]
    const projection = new Projection({
      code: 'EPSG:28992',
      units: 'm',
      extent: projectionExtent,
    })
    const resolutions = [3440.64, 1720.32, 860.16, 430.08, 215.04, 107.52, 53.76, 26.88, 13.44, 6.72, 3.36, 1.68, 0.84, 0.42, 0.21]
    //const size = ol.extent.getWidth(projectionExtent) / 256

    const matrixIds = []
    for (let i = 0; i < resolutions.length; ++i) {
      matrixIds[i] = 'EPSG:28992:' + i
    }
    return new Tile({
      source: new WMTSSource({
        attributions: 'Kaartgegevens: &copy; <a href="https://www.kadaster.nl">Kadaster</a>',
        url: environment.bgtBackgroundUrl,
        layer: 'grijs',
        matrixSet: 'EPSG:28992',
        format: 'image/png',
        projection: projection,
        tileGrid: new WMTSTileGrid({
          origin: getTopLeft(projectionExtent),
          resolutions: resolutions,
          matrixIds: matrixIds,
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
            const featureid = feature.getId()
            if (featureid) {
              const items = 'items'
              const itemsurl = this.itemsUrl.toLowerCase()
              const currentUrl = new URL(itemsurl.substring(0, itemsurl.indexOf(items) + items.length))
              tooltipContent.innerHTML =
                '<a href="' +
                currentUrl.protocol +
                '//' +
                currentUrl.host +
                currentUrl.pathname +
                '/' +
                featureid +
                '">' +
                featureid +
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
