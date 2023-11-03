import { ChangeDetectionStrategy, Component, ElementRef, EventEmitter, Input, OnChanges, OnInit, Output } from '@angular/core'
import { NgChanges } from '../app.component'
import { Feature, MapBrowserEvent, Map as OLMap, Overlay, View } from 'ol'
import TileLayer from 'ol/layer/Tile'
import { OSM, Vector as VectorSource, WMTS as WMTSSource } from 'ol/source'
import WMTSTileGrid from 'ol/tilegrid/WMTS'
import { FeatureCollectionGeoJSON } from '../openapi/model/featureCollectionGeoJSON'
import { Group, Tile, Vector as VectorLayer } from 'ol/layer'
import { Circle, Fill, Stroke, Style } from 'ol/style'
import { Extent, getCenter, getTopLeft } from 'ol/extent'
import { Geometry, Point, Polygon } from 'ol/geom'
import { FeatureServiceService, DataUrl } from '../feature-service.service'
import { Projection, ProjectionLike } from 'ol/proj'
import { take } from 'rxjs/operators'
import { FitOptions } from 'ol/View'
import { DragBox } from 'ol/interaction'
import { platformModifierKeyOnly, shiftKeyOnly } from 'ol/events/condition'
import { DragBoxEvent } from 'ol/interaction/DragBox'
import { WKT } from 'ol/format'
import { fromExtent } from 'ol/geom/Polygon'
import { FeatureLike } from 'ol/Feature'
import { PanIntoViewOptions } from 'ol/Overlay'
export function exhaustiveGuard(_value: never): never {
  throw new Error(`ERROR! Reached forbidden guard function with unexpected value: ${JSON.stringify(_value)}`)
}

@Component({
  selector: 'app-feature-view',
  templateUrl: './feature-view.component.html',
  styleUrls: ['./feature-view.component.css'],
  changeDetection: ChangeDetectionStrategy.OnPush,
})
export class FeatureViewComponent implements OnInit, OnChanges {
  @Input() itemsUrl!: string
  @Input() projection: ProjectionLike = 'EPSG:3857'
  @Input() backgroundMap: 'BRT' | 'OSM' = 'OSM'
  @Output() box = new EventEmitter<string>()
  @Output() activeFeature = new EventEmitter<FeatureLike>()
  mapHeight = 400
  mapWidth = 600
  map: OLMap = this.getMap()
  featureCollectionGeoJSON!: FeatureCollectionGeoJSON
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
    console.log('height---' + this.el.nativeElement.offsetHeight)
    console.log('width---' + this.el.nativeElement.offsetWidth)
    this.mapWidth = this.el.nativeElement.offsetWidth * 0.99
    this.mapHeight = this.mapWidth * 0.75 // height = 0.75 * width creates 4:3 aspect ratio
    console.log('mapheight---' + this.mapHeight)
    console.log('mapwidth---' + this.mapWidth)
    const mapdiv: HTMLElement = this.el.nativeElement.querySelector("[id='map']")
    console.log(mapdiv)
    this.map.setTarget(mapdiv)
    console.log('url: ' + this.itemsUrl)
    console.log('projection: ' + this.projection)
    const aurl: DataUrl = { url: this.itemsUrl, projection: this.projection }
    this.featureService
      .getFeatures(aurl)
      .pipe(take(1))
      .subscribe(data => {
        this.map.setLayerGroup(new Group())

        this.features = data
        const ext = this.loadfeatures(this.features)
        this.loadbackground()
        this.adddragbox()
        this.addFeatureEmit()
      })
  }

  ngOnChanges(changes: NgChanges<FeatureViewComponent>) {
    if (
      changes.itemsUrl?.previousValue !== changes.itemsUrl?.currentValue ||
      changes.projection.previousValue !== changes.projection.currentValue
    ) {
      console.log('url: ' + changes.itemsUrl?.currentValue)
      this.init()
    }
  }

  ngOnInit() {
    this.init()
  }

  loadbackground() {
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

  loadfeatures(features: Feature<Geometry>[]) {
    const vsource = new VectorSource({
      //features: new GeoJSON().readFeatures(this.featureCollectionGeoJSON, { featureProjection: this.projection  }),
      features: features,
    })

    this.map.addLayer(
      new VectorLayer({
        source: vsource,
        style: this.getstyle(),
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

  getstyle() {
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
    const projection = new Projection({ code: 'EPSG:28992', units: 'm', extent: projectionExtent })
    const resolutions = [3440.64, 1720.32, 860.16, 430.08, 215.04, 107.52, 53.76, 26.88, 13.44, 6.72, 3.36, 1.68, 0.84, 0.42, 0.21]
    //const size = ol.extent.getWidth(projectionExtent) / 256

    const matrixIds = []
    for (let i = 0; i < resolutions.length; ++i) {
      matrixIds[i] = 'EPSG:28992:' + i
    }
    return new Tile({
      source: new WMTSSource({
        attributions: 'Kaartgegevens: &copy; <a href="https://www.kadaster.nl">Kadaster</a>',
        url: 'https://service.pdok.nl/brt/achtergrondkaart/wmts/v2_0?',
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

    this.map.on('pointermove', (evt: MapBrowserEvent<any>) => {
      this.map.forEachFeatureAtPixel(
        evt.pixel,
        (feature: FeatureLike) => {
          if (feature) {
            const featureid = feature.getId()
            if (featureid) {
              const currentUrl = new URL(this.itemsUrl)
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
            }
            tooltip.setPosition(getCenter(feature.getGeometry()!.getExtent()))
            tooltipContainer.style.visibility = 'visible'
            this.activeFeature.emit(feature)
          }
        },
        { hitTolerance: 3 }
      )
    })
  }

  adddragbox() {
    const dragBox = new DragBox({
      condition: platformModifierKeyOnly, //shiftKeyOnly,
    })
    this.map.addInteraction(dragBox)
    dragBox.on('boxstart', (evt: DragBoxEvent) => {
      if (this.boxLayer) {
        this.map.removeLayer(this.boxLayer)
        this.box.emit('')
      }
    })

    dragBox.on('boxend', (evt: DragBoxEvent) => {
      const bboxGeometry = dragBox.getGeometry()
      const bbox = new Feature({
        geometry: bboxGeometry,
        name: 'bbox',
      })
      console.log(evt)
      const bboxsource = new VectorSource({})
      bboxsource.addFeature(bbox)
      const format = new WKT()
      this.box.emit(format.writeFeature(bbox))
      const bboxStyle = new Style({
        stroke: new Stroke({
          color: 'blue',
          width: 3,
        }),
        fill: new Fill({
          color: 'rgba(0, 0, 255, 0.06)',
        }),
      })
      this.boxLayer = new VectorLayer({ source: bboxsource, style: bboxStyle })
      this.map.addLayer(this.boxLayer)
    })
  }
}
