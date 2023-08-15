import {
  Component,
  Input,
  ElementRef,
  Output,
  EventEmitter,
  CUSTOM_ELEMENTS_SCHEMA,
} from '@angular/core';
import { coerceBooleanProperty } from '@angular/cdk/coercion';
import { Subject } from 'rxjs';
import { ObjectInfoComponent } from './object-info/object-info.component';
import VectorTileSource from 'ol/source/VectorTile.js';
import TileDebug from 'ol/source/TileDebug.js';
import Map from 'ol/Map';
import View from 'ol/View';
import {
  EuropeanETRS89_LAEAQuad,
  MapProjection,
  NetherlandsRDNewQuadDefault,
} from '../app/mapprojection';

import { FullScreen, defaults as defaultControls } from 'ol/control.js';
import { applyStyle } from 'ol-mapbox-style';
import Projection from 'ol/proj/Projection';
import { Fill, Stroke, Style } from 'ol/style';
import { MVT } from 'ol/format';
import VectorTileLayer from 'ol/layer/VectorTile';
import { getTopLeft, getWidth } from 'ol/extent';
import TileGrid from 'ol/tilegrid/TileGrid';
import { ProjectionLike, useGeographic } from 'ol/proj';
import { Coordinate } from 'ol/coordinate';
import TileLayer from 'ol/layer/Tile';
import BaseLayer from 'ol/layer/Base';
import Collection from 'ol/Collection';
import LayerGroup from 'ol/layer/Group';
import { FeatureLike } from 'ol/Feature';
import { CommonModule } from '@angular/common';
import { Link, Matrix, MatrixsetService } from './matrixset.service';

export type NgChanges<
  Component extends object,
  Props = ExcludeFunctions<Component>
> = {
  [Key in keyof Props]: {
    previousValue: Props[Key];
    currentValue: Props[Key];
    firstChange: boolean;
    isFirstChange(): boolean;
  };
};

type MarkFunctionPropertyNames<Component> = {
  [Key in keyof Component]: Component[Key] extends Function | Subject<any>
    ? never
    : Key;
};

type ExcludeFunctionPropertyNames<T extends object> =
  MarkFunctionPropertyNames<T>[keyof T];
type ExcludeFunctions<T extends object> = Pick<
  T,
  ExcludeFunctionPropertyNames<T>
>;

@Component({
  selector: 'app-vectortile-view',
  templateUrl: './app.component.html',
  styleUrls: ['./app.component.css'],
  //encapsulation: ViewEncapsulation.ShadowDom,
  standalone: true,
  imports: [CommonModule, ObjectInfoComponent],
  schemas: [
    CUSTOM_ELEMENTS_SCHEMA, // Tells Angular we will have custom tags in our templates
  ],
})
export class AppComponent {
  title = 'vectortile-view-component';
  map = new Map({});
  selector = '/{z}/{y}/{x}?f=mvt';
  private _showGrid: boolean = false;
  private _showObjectInfo: boolean = false;
  vectorTileLayer: VectorTileLayer | undefined;
  curFeature!: FeatureLike;
  tileGrid: TileGrid | undefined;
  minZoom?: number;
  maxZoom?: number;

  @Input() set showGrid(showGrid: any) {
    this._showGrid = coerceBooleanProperty(showGrid);
  }
  get showGrid() {
    return this._showGrid;
  }

  @Input() set showObjectInfo(showObjectInfo: any) {
    this._showObjectInfo = coerceBooleanProperty(showObjectInfo);
  }
  get showObjectInfo() {
    return this._showObjectInfo;
  }

  @Input() tileUrl: string = NetherlandsRDNewQuadDefault;
  @Input() styleUrl!: string;
  @Input() id!: string | undefined;

  @Input() get zoom(): number {
    const z = this.map.getView().getZoom();
    if (z) {
      return z;
    } else {
      return -1;
    }
  }

  set zoom(value: number) {
    this.map.getView().setZoom(value);
    this.currentZoomLevel.next(value);
  }

  @Output() currentZoomLevel = new EventEmitter<number>();
  @Input() centerX!: number;
  @Input() centerY!: number;
  mapHeight: number = 600;
  mapWidth: number = 800;

  @Output() activeFeature = new EventEmitter<FeatureLike>();

  constructor(
    private elementRef: ElementRef,
    private matrixsetService: MatrixsetService
  ) {}

  ngOnChanges(changes: NgChanges<AppComponent>) {
    if (changes.styleUrl?.previousValue !== changes.styleUrl?.currentValue) {
      //console.log(this.id + ' style changed')
      if (!changes.styleUrl.isFirstChange()) {
        if (this.vectorTileLayer) {
          this.setStyle(this.vectorTileLayer);
        }
      }
    }
    if (changes.tileUrl?.previousValue !== changes.tileUrl?.currentValue) {
      //console.log(this.id + ' projection changed')
      if (changes.tileUrl.isFirstChange()) {
        this.checkParams();
      } else {
        this.maxZoom = undefined;
        this.minZoom = undefined;
      }
      this.initialize();
    }
  }

  private initialize() {
    this.vectorTileLayer = undefined;

    let matrixurl = this.tileUrl.replace('tiles', 'tileMatrixSets') + '?f=json';
    console.log('url: ' + this.tileUrl);
    this.matrixsetService.getMatrix(this.tileUrl).subscribe({
      next: (tile) => {
        const linkurl = this.FindMatrixUrl(tile.links);
        if (linkurl) {
          matrixurl = linkurl;
        }
        this.SetZoom(tile);
        this.drawFromMatrixUrl(tile, matrixurl);
      },
      error: (msg) => {
        console.log(this.id + 'error: ' + JSON.stringify(msg));
      },
    });
  }

  private SetZoom(tile: Matrix) {
    tile.tileMatrixSetLimits.forEach((limit) => {
      let zoomHack = 0;

      if (this.tileUrl.includes('WebMercatorQuad')) {
        //the matrix is not correct on server?  size of vector 512px,  256px in tile grid correct?


        zoomHack = 1;
      }

      if (this.tileUrl.includes('EuropeanETRS89_LAEAQuad')){
        //why is this needed??
        zoomHack = -1;
      }


        this.zoom = parseFloat(limit.tileMatrix) + 1 + zoomHack;
            // Only show available tiles

        this.minZoom = parseFloat(limit.tileMatrix) + 1 + zoomHack;

      this.maxZoom = parseFloat(limit.tileMatrix) + 1 + zoomHack  ;
    });
  }

  private FindMatrixUrl(links: Link[]) {
    let matrixurl = undefined;
    links.forEach((link) => {
      if (link.rel == 'http://www.opengis.net/def/rel/ogc/1.0/tiling-scheme') {
        console.log(this.id + ' url for matrix: ' + link.href);
        let turl = new URL(this.tileUrl);
        if (this.isFullURL(link.href)) {
          matrixurl = link.href;
        } else {
          let mUrl = new URL(turl.origin + link.href);
          matrixurl = mUrl.href;
        }
      }
    });
    return matrixurl;
  }

  private drawFromMatrixUrl(matrix: Matrix, matrixurl: string) {
    this.matrixsetService.getMatrixSet(matrixurl).subscribe({
      next: (matrixset) => {
        let resolutions: number[] = [];
        let origins: number[][] = [];
        let sizes: number[][] = [];
        matrixset.tileMatrices.forEach((x) => {
          resolutions[x.id] = x.cellSize;

          if (this.tileUrl.includes(EuropeanETRS89_LAEAQuad)) {
            origins[x.id] = [x.pointOfOrigin[1], x.pointOfOrigin[0]]; //  x,y swap Workaround?
          } else {
            origins[x.id] = x.pointOfOrigin;
          }
          sizes[x.id] = [x.tileWidth, x.tileHeight];
        });

        this.tileGrid = new TileGrid({
          resolutions: resolutions,
          tileSizes: sizes,
          origins: origins,
        });

        this.drawMap(matrix);
      },
      error: (error) => {
        console.log(this.id + 'tilematrixset not found: ' + matrixurl);
        const proj = new MapProjection(this.tileUrl).Projection;
        this.tileGrid = new TileGrid({
          extent: proj.getExtent(),
          resolutions: this.calcResolutions(proj),
          tileSize: [256, 256],
          origin: getTopLeft(proj.getExtent()),
        });
        this.drawMap(matrix);
      },
    });
  }

  private drawMap(tile: Matrix) {
    this.map.setTarget(undefined);
    this.map =new Map({});
    this.zoom = -1
    let map = this.getMap();

    map.on('pointermove', (evt: { pixel: any }) => {
      map.forEachFeatureAtPixel(
        evt.pixel,
        (feature: FeatureLike) => {
          if (feature) {
            if (this._showObjectInfo) {
              this.curFeature = feature;
            }
            this.activeFeature.emit(feature);
          }
        },
        { hitTolerance: 3 }
      );
    });

    map.getView().on('change:resolution', (event) => {
      console.log('zoom changed');
      this.currentZoomLevel.next(this.map.getView().getZoom()!);
    });

    this.SetZoom(tile);

    const mapdiv: HTMLElement =
      this.elementRef.nativeElement.querySelector("[id='map']");


    this.mapWidth = this.elementRef.nativeElement.offsetWidth;
    this.mapHeight = this.elementRef.nativeElement.offsetWidth * 0.75 // height = 0.75 * width creates 4:3 aspect ratio
    map.setTarget(mapdiv);

  }

  private checkParams(): void {
    console.log(this.id);
    if (!this.tileUrl) {
      console.error('No TilteUrl was provided for the app-vectortile-view');
    }
    if (!this.styleUrl) {
      console.log('No StyleUrl was provided for the app-vectortile-view');
    }

    if (!this.centerX) {
      console.error(
        'No zoom center-x was provided for the app-vectortile-view'
      );
    } else console.log('center-x=' + this.centerX);
    if (!this.centerY) {
      console.error('No center-y was provided for the app-vectortile-view');
    } else console.log('center-y=' + this.centerY);
  }

  getMap() {
    useGeographic();

    const l = this.generateLayers();
    const layers = l.layers;

    let acenter: Coordinate = [this.centerX, this.centerY];
    this.vectorTileLayer = l.vectorTileLayer;
    this.map = new Map({
      controls: defaultControls().extend([new FullScreen()]),

      layers: layers,
      view: new View({
        center: acenter,
        zoom: this.zoom,
        maxZoom: this.maxZoom,
        minZoom: this.minZoom,
        enableRotation: false,
        projection: l.vectorTileLayer
          .getSource()
          ?.getProjection() as ProjectionLike,
      }),
    });
    return this.map;
  }

  private generateLayers() {


    let vectorTileLayer = this.getVectortileLayer(new MapProjection(this.tileUrl).Projection);
    this.setStyle(vectorTileLayer);

    let layers = [vectorTileLayer] as
      | BaseLayer[]
      | Collection<BaseLayer>
      | LayerGroup
      | undefined;

    if (this.showGrid) {
      const debugLayer = new TileLayer({
        source: new TileDebug({
          template: 'z:{z} y:{y} x:{x}',
          projection: vectorTileLayer
            .getSource()!
            .getProjection() as ProjectionLike,
          tileGrid: vectorTileLayer.getSource()!.getTileGrid() as TileGrid,
          wrapX: vectorTileLayer.getSource()!.getWrapX(),
          zDirection: vectorTileLayer.getSource()!.zDirection,
        }),
      });
      layers = [vectorTileLayer, debugLayer];
    }
    return { vectorTileLayer: vectorTileLayer, layers: layers };
  }

  private setStyle(vectorTileLayer: VectorTileLayer) {
    if (this.styleUrl) {
      applyStyle(vectorTileLayer, this.styleUrl)
        .then(() => {
          //overrule source url from style
          if (this.tileUrl !== NetherlandsRDNewQuadDefault) {
            vectorTileLayer.getSource()?.setUrl(this.tileUrl + this.selector);
          }
        })
        .catch((err) =>
          console.error(
            'error loading: ' + this.id + ' ' + this.styleUrl + ' ' + err
          )
        );
    } else {
      const defaultStyle = new Style({
        fill: new Fill({
          color: 'rgba(255,255,255,0.4)',
        }),
        stroke: new Stroke({
          color: '#3399CC',
          width: 1.25,
        }),
      });
      vectorTileLayer.setStyle(defaultStyle);
    }
  }

  getVectortileLayer(projection: Projection): VectorTileLayer {
    const vectorTileLayer = new VectorTileLayer({
      source: this.getVectorTileSource(projection, this.tileUrl),
      renderMode: 'hybrid',
      declutter: true,
      useInterimTilesOnError: false,
    });
    return vectorTileLayer;
  }

  private calcResolutions(projection: Projection) {
    const tileSizePixels = 256;
    const tileSizeMtrs = getWidth(projection.getExtent()) / tileSizePixels;
    let resolutions: Array<number> = [];
    for (let i = 0; i <= 21; i++) {
      resolutions[i] = tileSizeMtrs / Math.pow(2, i);
    }
    return resolutions;
  }

  private getVectorTileSource(projection: Projection, url: string) {
    return new VectorTileSource({
      format: new MVT(),
      projection: projection,
      tileGrid: this.tileGrid,
      url: url + this.selector,
      cacheSize: 0,
    });
  }

  isFullURL(url: string): boolean {
    return (
      url.toLowerCase().startsWith('http://') ||
      url.toLowerCase().startsWith('https://')
    );
  }
}
