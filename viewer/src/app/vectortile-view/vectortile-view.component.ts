import { coerceBooleanProperty } from '@angular/cdk/coercion'
import {
  ChangeDetectionStrategy,
  ChangeDetectorRef,
  Component,
  CUSTOM_ELEMENTS_SCHEMA,
  ElementRef,
  EventEmitter,
  Input,
  OnChanges,
  Output,
} from '@angular/core'
import Map from 'ol/Map'
import TileDebug from 'ol/source/TileDebug.js'
import VectorTileSource from 'ol/source/VectorTile.js'
import View from 'ol/View'
import { Subject } from 'rxjs'
import { EuropeanETRS89_LAEAQuad, MapProjection, NetherlandsRDNewQuadDefault } from '../mapprojection'
import { ObjectInfoComponent } from '../object-info/object-info.component'

import { CommonModule } from '@angular/common'
import { NGXLogger } from 'ngx-logger'
import { MapBrowserEvent, VectorTile } from 'ol'
import { applyStyle } from 'ol-mapbox-style'
import Collection from 'ol/Collection'
import { defaults as defaultControls, FullScreen } from 'ol/control.js'
import { Coordinate } from 'ol/coordinate'
import { getTopLeft, getWidth } from 'ol/extent'
import { FeatureLike } from 'ol/Feature'
import { MVT } from 'ol/format'
import BaseLayer from 'ol/layer/Base'
import LayerGroup from 'ol/layer/Group'
import TileLayer from 'ol/layer/Tile'
import VectorTileLayer from 'ol/layer/VectorTile'
import { useGeographic } from 'ol/proj'
import Projection from 'ol/proj/Projection'
import { Fill, Stroke, Style } from 'ol/style'
import TileGrid from 'ol/tilegrid/TileGrid'
import { Link, Matrix, MatrixsetService } from '../matrixset.service'

export type NgChanges<Component extends object, Props = ExcludeFunctions<Component>> = {
  [Key in keyof Props]: {
    previousValue: Props[Key]
    currentValue: Props[Key]
    firstChange: boolean
    isFirstChange(): boolean
  }
}

type MarkFunctionPropertyNames<Component> = {
  // eslint-disable-next-line @typescript-eslint/ban-types
  [Key in keyof Component]: Component[Key] extends Function | Subject<never> ? never : Key
}

type ExcludeFunctionPropertyNames<T extends object> = MarkFunctionPropertyNames<T>[keyof T]
type ExcludeFunctions<T extends object> = Pick<T, ExcludeFunctionPropertyNames<T>>

@Component({
  selector: 'app-vectortile-view',
  templateUrl: './vectortile-view.component.html',
  styleUrls: ['./vectortile-view.component.css'],
  //encapsulation: ViewEncapsulation.ShadowDom,
  standalone: true,
  changeDetection: ChangeDetectionStrategy.OnPush,

  imports: [CommonModule, ObjectInfoComponent],
  schemas: [
    CUSTOM_ELEMENTS_SCHEMA, // Tells Angular we will have custom tags in our templates
  ],
})
export class VectortileViewComponent implements OnChanges {
  title = 'view-component'
  map = new Map({})
  xyzSelector = '/{z}/{y}/{x}?f=mvt'
  private _showGrid = false
  private _showObjectInfo = false
  vectorTileLayer: VectorTileLayer | undefined
  curFeature!: FeatureLike
  tileGrid: TileGrid | undefined
  minZoom?: number
  maxZoom?: number
  private _zoom = -1
  private projection!: Projection

  @Input() set showGrid(showGrid: boolean) {
    this._showGrid = coerceBooleanProperty(showGrid)
  }
  get showGrid() {
    return this._showGrid
  }

  @Input() set showObjectInfo(showObjectInfo: boolean) {
    this._showObjectInfo = coerceBooleanProperty(showObjectInfo)
  }
  get showObjectInfo() {
    return this._showObjectInfo
  }

  @Input() tileUrl: string = NetherlandsRDNewQuadDefault
  @Input() styleUrl!: string
  @Input() id!: string | undefined

  @Input()
  set zoom(value: number) {
    this._zoom = value
    if (value != -1) {
      this.map.getView().setZoom(value)
      this.currentZoomLevel.next(value)
    }
  }
  get zoom(): number {
    return this._zoom
  }

  @Output() currentZoomLevel = new EventEmitter<number>()
  @Output() activeFeature = new EventEmitter<FeatureLike>()
  @Output() activeTileUrl = new EventEmitter<string>()
  @Input() centerX!: number
  @Input() centerY!: number
  mapHeight = 600
  mapWidth = 800

  constructor(
    private logger: NGXLogger,
    private elementRef: ElementRef,
    private matrixsetService: MatrixsetService,
    private cdf: ChangeDetectorRef
  ) {}

  ngOnChanges(changes: NgChanges<VectortileViewComponent>) {
    if (changes.styleUrl?.previousValue !== changes.styleUrl?.currentValue) {
      // this.logger.log(this.id + ' style changed')
      if (!changes.styleUrl.isFirstChange()) {
        if (this.vectorTileLayer) {
          this.setStyle(this.vectorTileLayer)
        }
      }
    }
    if (changes.tileUrl?.previousValue !== changes.tileUrl?.currentValue) {
      // this.logger.log(this.id + ' projection changed')
      this.maxZoom = undefined
      this.minZoom = undefined
      this.zoom = -1
      this.checkParams()
      this.initialize()
    }
  }

  private initialize() {
    this.vectorTileLayer = undefined
    let matrixUrl = this.tileUrl.replace('tiles', 'tileMatrixSets') + '?f=json'
    //  this.logger.log('url: ' + this.tileUrl)
    this.matrixsetService.getMatrix(this.tileUrl).subscribe({
      next: tile => {
        const linkUrl = this.FindMatrixUrl(tile.links)
        if (linkUrl) {
          matrixUrl = linkUrl
        } else {
          this.logger.log('tileurl :' + this.tileUrl + 'not found')
        }
        this.drawFromMatrixUrl(tile, matrixUrl)
        this.SetZoomLevel(tile)
        this.cdf.detectChanges()
      },
      error: msg => {
        this.logger.log(this.id + 'error: ' + JSON.stringify(msg))
      },
    })
  }

  private SetZoomLevel(tile: Matrix) {
    tile.tileMatrixSetLimits.forEach(limit => {
      if (!this.minZoom) {
        this.minZoom = parseFloat(limit.tileMatrix) + 0.01
        this.zoom = this.minZoom
      }
      this.maxZoom = parseFloat(limit.tileMatrix) + 1
    })
  }

  private FindMatrixUrl(links: Link[]) {
    let matrixUrl = undefined
    links.forEach(link => {
      if (link.rel == 'http://www.opengis.net/def/rel/ogc/1.0/tiling-scheme') {
        const turl = new URL(this.tileUrl)
        if (this.isFullURL(link.href)) {
          matrixUrl = link.href
        } else {
          const mUrl = new URL(turl.origin + link.href)
          matrixUrl = mUrl.href
        }
      }
    })
    return matrixUrl
  }

  private drawFromMatrixUrl(matrix: Matrix, matrixUrl: string) {
    this.matrixsetService.getMatrixSet(matrixUrl).subscribe({
      next: matrixSet => {
        const resolutions: number[] = []
        const origins: number[][] = []
        const sizes: number[][] = []
        matrixSet.tileMatrices.forEach(x => {
          resolutions[x.id] = x.cellSize
          if (this.tileUrl.includes(EuropeanETRS89_LAEAQuad)) {
            origins[x.id] = [x.pointOfOrigin[1], x.pointOfOrigin[0]] //  x,y swap Workaround?
          } else {
            origins[x.id] = x.pointOfOrigin
          }
          sizes[x.id] = [x.tileWidth, x.tileHeight]
        })

        this.tileGrid = new TileGrid({
          resolutions: resolutions,
          tileSizes: sizes,
          origins: origins,
        })
        this.drawMap(matrix)
      },
      error: error => {
        this.logger.log(this.id + 'tilematrixset not found: ' + matrixUrl, error)
        this.projection = new MapProjection(this.tileUrl).Projection
        this.tileGrid = new TileGrid({
          extent: this.projection.getExtent(),
          resolutions: this.calcResolutions(this.projection),
          tileSize: [256, 256],
          origin: getTopLeft(this.projection.getExtent()),
        })
        this.drawMap(matrix)
      },
    })
  }

  private drawMap(tile: Matrix) {
    this.map.setTarget(undefined)
    this.map = new Map({})
    const map = this.getMap()
    // eslint-disable-next-line @typescript-eslint/no-explicit-any
    map.on('pointermove', (evt: MapBrowserEvent<any>) => {
      map.forEachFeatureAtPixel(
        evt.pixel,
        (feature: FeatureLike) => {
          if (feature) {
            if (this._showObjectInfo) {
              this.curFeature = feature
              this.cdf.detectChanges()
            }
            this.activeFeature.emit(feature)
          }
        },
        { hitTolerance: 3 }
      )
    })

    map.getView().on('change:resolution', () => {
      const zoom = this.map.getView().getZoom()
      this.logger.log('zoom' + zoom)
      if (zoom) {
        this._zoom = zoom
        this.currentZoomLevel.next(zoom)
      }
    })

    this.SetZoomLevel(tile)
    const mapElm: HTMLElement = this.elementRef.nativeElement.querySelector("[id='map']")
    this.mapWidth = this.elementRef.nativeElement.offsetWidth
    this.mapHeight = this.elementRef.nativeElement.offsetWidth * 0.75 // height = 0.75 * width creates 4:3 aspect ratio
    map.setTarget(mapElm)
    this.cdf.detectChanges()
  }

  private checkParams(): void {
    this.logger.log(this.id)
    if (!this.tileUrl) {
      this.logger.error('No TilteUrl was provided for the app-vectortile-view')
    }
    if (!this.styleUrl) {
      this.logger.log('No StyleUrl was provided for the app-vectortile-view')
    }

    if (!this.centerX) {
      this.logger.error('No zoom center-x was provided for the app-vectortile-view')
    } else this.logger.log('center-x=' + this.centerX)
    if (!this.centerY) {
      this.logger.error('No center-y was provided for the app-vectortile-view')
    } else this.logger.log('center-y=' + this.centerY)
  }

  getMap() {
    useGeographic()
    const l = this.generateLayers()
    const layers = l.layers
    const mapCenter: Coordinate = [this.centerX, this.centerY]
    this.vectorTileLayer = l.vectorTileLayer

    const contr = defaultControls({
      // eslint-disable-next-line @typescript-eslint/no-non-null-assertion
      zoom: this.maxZoom! - this.minZoom! > 1,
    }).extend([new FullScreen()])
    this.map = new Map({
      controls: contr,
      layers: layers,
      view: new View({
        center: mapCenter,
        zoom: this.zoom,
        maxZoom: this.maxZoom,
        minZoom: this.minZoom,
        enableRotation: false,
        projection: this.projection,
      }),
    })
    return this.map
  }

  private generateLayers() {
    this.projection = new MapProjection(this.tileUrl).Projection
    const vectorTileLayer = this.getVectorTileLayer(this.projection)
    this.setStyle(vectorTileLayer)
    let layers = [vectorTileLayer] as BaseLayer[] | Collection<BaseLayer> | LayerGroup | undefined

    if (this.showGrid) {
      const source = vectorTileLayer.getSource()
      if (source) {
        const grid = source.getTileGrid()
        if (grid) {
          const debugLayer = new TileLayer({
            source: new TileDebug({
              template: 'z:{z} y:{y} x:{x}',
              projection: this.projection,
              tileGrid: grid,
              wrapX: source.getWrapX(),
              zDirection: 1,
            }),
          })
          layers = [vectorTileLayer, debugLayer]
        }
      }
    }
    return { vectorTileLayer: vectorTileLayer, layers: layers }
  }

  private setStyle(vectorTileLayer: VectorTileLayer) {
    if (this.styleUrl) {
      const projection = vectorTileLayer.getSource()?.getProjection()
      applyStyle(vectorTileLayer, this.styleUrl)
        .then(() => {
          //overrule source url and zoom from style
          if (this.tileUrl !== NetherlandsRDNewQuadDefault) {
            // eslint-disable-next-line @typescript-eslint/no-non-null-assertion
            vectorTileLayer.setSource(this.getVectorTileSource(projection!, this.tileUrl))
          }
        })
        .catch(err => this.logger.error('error loading: ' + this.id + ' ' + this.styleUrl + ' ' + err))
    } else {
      const defaultStyle = new Style({
        fill: new Fill({
          color: 'rgba(255,255,255,0.4)',
        }),
        stroke: new Stroke({
          color: '#3399CC',
          width: 1.25,
        }),
      })
      vectorTileLayer.setStyle(defaultStyle)
    }
  }

  getVectorTileLayer(projection: Projection): VectorTileLayer {
    return new VectorTileLayer({
      source: this.getVectorTileSource(projection, this.tileUrl),
      renderMode: 'hybrid',
      declutter: true,
      useInterimTilesOnError: false,
    })
  }

  private calcResolutions(projection: Projection) {
    const tileSizePixels = 256
    const tileSizeMtrs = getWidth(projection.getExtent()) / tileSizePixels
    const resolutions: Array<number> = []
    for (let i = 0; i <= 21; i++) {
      resolutions[i] = tileSizeMtrs / Math.pow(2, i)
    }
    return resolutions
  }

  private getVectorTileSource(projection: Projection, url: string) {
    const source = new VectorTileSource({
      format: new MVT(),
      projection: projection,
      tileGrid: this.tileGrid,
      url: url + this.xyzSelector,
      cacheSize: 0,
    })
    source.on(['tileloadend'], e => {
      const evt: { type: 'tileloadend'; target: VectorTile; tile: VectorTile } = e as never
      this.activeTileUrl.next(evt.tile.key)
    })
    return source
  }

  isFullURL(url: string): boolean {
    return url.toLowerCase().startsWith('http://') || url.toLowerCase().startsWith('https://')
  }
}
