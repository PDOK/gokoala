import { ChangeDetectionStrategy, Component, ElementRef, Input, OnInit } from '@angular/core'
import { NGXLogger } from 'ngx-logger'
import { Feature, Map as OLMap, VectorTile, View } from 'ol'
import { Feature, Map as OLMap, VectorTile, View } from 'ol'
import { applyStyle } from 'ol-mapbox-style'
import { getCenter } from 'ol/extent'
import { MVT } from 'ol/format'
import { Geometry, LineString, Point } from 'ol/geom'
import { fromExtent } from 'ol/geom/Polygon'
import VectorTileLayer from 'ol/layer/VectorTile'
import { Projection } from 'ol/proj'
import VectorTileSource from 'ol/source/VectorTile.js'
import { exhaustiveGuard, LayerType, LegendItem, MapboxStyle, MapboxStyleService } from '../../mapbox-style.service'

@Component({
  selector: 'app-legend-item',
  templateUrl: './legend-item.component.html',
  styleUrls: ['./legend-item.component.css'],
  standalone: true,
  changeDetection: ChangeDetectionStrategy.OnPush,
})
export class LegendItemComponent implements OnInit {
  constructor(
    private logger: NGXLogger,
    private mapboxStyleService: MapboxStyleService,
    private elementRef: ElementRef
  ) {}

  @Input() item!: LegendItem
  @Input() mapboxStyle!: MapboxStyle

  private readonly _itemHeight = 40
  get itemHeight() {
    return this._itemHeight
  }
  private readonly _itemWidth = 60
  get itemWidth() {
    return this._itemWidth
  }
  itemLeft = 10
  itemRight = 50
  readonly extent = [0, 0, this.itemWidth, this.itemHeight]

  projection = new Projection({
    code: 'pixel-map',
    units: 'pixels',
    extent: this.extent,
  })

  map: OLMap = new OLMap({})
  vectorSource = new VectorTileSource({
    format: new MVT(),
    projection: this.projection,
  })

  vectorLayer = new VectorTileLayer({
    source: this.vectorSource,
  })

  ngOnInit() {
    const feature = this.newFeature(this.item)
    this.map = new OLMap({
      controls: [],
      interactions: [],

      layers: [this.vectorLayer],
      view: new View({
        projection: this.projection,
        center: getCenter(this.extent),
        zoom: 2,
        minZoom: 2,
        maxZoom: 2,
      }),
    })

    this.vectorLayer.getSource()?.setTileLoadFunction(tile => {
      const vectorTile = tile as VectorTile<Feature<Geometry>>
      vectorTile.setLoader(() => {
        const features: Feature<Geometry>[] = []
        features.push(feature)
        vectorTile.setFeatures(features)
      })
    })

    const resolutions: number[] = []
    resolutions.push(1)
    const sources = this.mapboxStyleService.getLayersids(this.mapboxStyle)

    applyStyle(this.vectorLayer, this.mapboxStyle, sources, undefined, resolutions)
      .then(() => {
        this.vectorLayer.getSource()?.refresh()
        const mapdiv: HTMLElement = this.elementRef.nativeElement.querySelector("[id='itemmap']")
        this.map.setTarget(mapdiv)
      })
      // eslint-disable-next-line @typescript-eslint/no-explicit-any
      .catch((err: any) => {
        this.logger.error('error loading legend style: ' + ' ' + err)
      })
  }

  private newFeature(item: LegendItem): Feature {
    switch (item.geoType) {
      case LayerType.Fill: {
        const ageom = fromExtent(this.extent)
        ageom.scale(0.05, 0.05)
        const f = new Feature({
          geometry: ageom,
          layer: item.sourceLayer,
        })
        f.setProperties(item.properties)
        return f
      }
      case LayerType.Circle:
      case LayerType.Raster:
      case LayerType.Symbol: {
        const f = new Feature({
          geometry: new Point(getCenter(this.extent)),
          layer: item.sourceLayer,
        })
        f.setProperties(item.properties)
        return f
      }
      case LayerType.Line: {
        const half = this.itemHeight / 2
        const f = new Feature({
          geometry: new LineString([
            [this.itemLeft, half],
            [this.itemRight, half],
          ]),
          layer: item.sourceLayer,
        })
        f.setProperties(item.properties)
        return f
      }
      default: {
        exhaustiveGuard(item.geoType)
      }
    }
  }
}
