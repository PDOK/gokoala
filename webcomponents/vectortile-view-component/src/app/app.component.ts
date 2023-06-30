import {
  Component,
  OnInit,
  Input,
  ElementRef,
  Output,
  EventEmitter,
  CUSTOM_ELEMENTS_SCHEMA
} from '@angular/core';
import { coerceBooleanProperty } from '@angular/cdk/coercion';
import { Subject } from 'rxjs';
import { ObjectInfoComponent } from './object-info/object-info.component';
import VectorTileSource from 'ol/source/VectorTile.js';
import TileDebug from 'ol/source/TileDebug.js';
import Map from 'ol/Map';
import View from 'ol/View';
import { EuropeanETRS89_GRS80, MapProjection, NetherlandsRDNewQuadDefault } from '../app/mapprojection'
import { applyStyle } from 'ol-mapbox-style';
import Projection from 'ol/proj/Projection';
import { Fill, Stroke, Style } from "ol/style";
import { MVT } from "ol/format";
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
import { MatrixsetService } from './matrixset.service';


export type NgChanges<Component extends object, Props = ExcludeFunctions<Component>> = {
  [Key in keyof Props]: {
    previousValue: Props[Key];
    currentValue: Props[Key];
    firstChange: boolean;
    isFirstChange(): boolean;
  }
}

type MarkFunctionPropertyNames<Component> = {
  [Key in keyof Component]: Component[Key] extends Function | Subject<any> ? never : Key;
}

type ExcludeFunctionPropertyNames<T extends object> = MarkFunctionPropertyNames<T>[keyof T];
type ExcludeFunctions<T extends object> = Pick<T, ExcludeFunctionPropertyNames<T>>;

@Component({
  selector: 'app-vectortile-view',
  templateUrl: './app.component.html',
  styleUrls: ['./app.component.css'],
  //encapsulation: ViewEncapsulation.ShadowDom, 
  standalone: true,
  imports: [CommonModule, ObjectInfoComponent],
  schemas: [
    CUSTOM_ELEMENTS_SCHEMA // Tells Angular we will have custom tags in our templates
  ]

})
export class
  AppComponent implements OnInit {
  title = 'vectortile-view-component';
  map = new Map({});
  selector = '/{z}/{y}/{x}?f=mvt'
  private _showGrid: boolean = false;
  private _showObjectInfo: boolean = false;
  vectorTileLayer!: VectorTileLayer;
  curFeature!: FeatureLike;
  tileGrid: TileGrid | undefined;
  minZoom?: number;
  maxZoom?: number;

  @Input() set showGrid(showGrid: any) {
    this._showGrid = coerceBooleanProperty(showGrid);
  }
  get showGrid() {
    return this._showGrid
  }

  @Input() set showObjectInfo(showObjectInfo: any) {
    this._showObjectInfo = coerceBooleanProperty(showObjectInfo);
  }
  get showObjectInfo() {
    return this._showObjectInfo
  }

  @Input() tileUrl: string = NetherlandsRDNewQuadDefault
  @Input() styleUrl!: string
  @Input() id!: string | undefined
  @Input() zoom!: number
  @Input() centerX!: number;
  @Input() centerY!: number;
  totalHeight: number = 600
  totalWidth: number = 800
  @Output() activeFeature = new EventEmitter<FeatureLike>();

  constructor(private elementRef: ElementRef, private matrixsetService: MatrixsetService) {
  }

  ngOnChanges(changes: NgChanges<AppComponent>) {
    if (changes.styleUrl?.previousValue !== changes.styleUrl?.currentValue) {
      //console.log(this.id + ' style changed')
      if (this.vectorTileLayer) {
        this.setStyle(this.vectorTileLayer);
      }
    }
    if (changes.tileUrl?.previousValue !== changes.tileUrl?.currentValue) {
      //console.log(this.id + ' projection changed')
      if (this.vectorTileLayer) {
        this.setNewProjection();
      }
    }
  }

  ngOnInit() {
    this.checkParams();
    let matrixurl = this.tileUrl.replace("tiles", "tileMatrixSets") + '?f=json';
    console.log("url: " + this.tileUrl)
    this.matrixsetService.getMatrix(this.tileUrl).subscribe({
      next: tile => {

        tile.links.forEach(link => {

          if (link.rel == 'http://www.opengis.net/def/rel/ogc/1.0/tiling-scheme') {
            console.log(this.id + " url for matrix: " + link.href)
            let turl = new URL(this.tileUrl)

            if (isFullURL(link.href)) {
              matrixurl = link.href
            }
            else {
              let mUrl = new URL(turl.origin + link.href)
              matrixurl = mUrl.href
            }
          }
        })

        tile.tileMatrixSetLimits.forEach(limit => {


          if (!this.zoom) {
            this.zoom = parseFloat(limit.tileMatrix) + 1
          }
          // Only show available tiles
          if (!this.minZoom) {

            this.minZoom = parseFloat(limit.tileMatrix) + 1
          }
          this.maxZoom = parseFloat(limit.tileMatrix) + 1
        })
        this.matrixsetService.getMatrixSet(matrixurl).subscribe({
          next: matrixset => {
            let resolutions: number[] = [];
            let origins: number[][] = []
            let sizes: number[][] = []
            matrixset.tileMatrices.forEach(x => {
              resolutions[x.id] = x.cellSize

              if (this.tileUrl.includes(EuropeanETRS89_GRS80)) {
                origins[x.id] = [x.pointOfOrigin[1], x.pointOfOrigin[0]] //  x,y swap Workaround? 
              }
              else {
                origins[x.id] = x.pointOfOrigin
              }
              sizes[x.id] = [x.tileWidth, x.tileHeight]
            })

            this.tileGrid = new TileGrid({
              resolutions: resolutions,
              tileSizes: sizes,
              origins: origins
            })
            this.drawMap();
          },
          error: error => {
            console.log(this.id + 'tilematrixset not found: ' + matrixurl)
            const proj = new MapProjection(this.tileUrl).Projection
            this.tileGrid = new TileGrid({
              extent: proj.getExtent(),
              resolutions: this.calcResolutions(proj),
              tileSize: [256, 256],
              origin: getTopLeft(proj.getExtent())
            })
            this.drawMap();
          }
        })








      },
      error: msg => {

        console.log(this.id + "error: " + JSON.stringify(msg))
      }

    })


    function isFullURL(url: string): boolean {
      return url.toLowerCase().startsWith('http://') || url.toLowerCase().startsWith('https://');
    }


    //console.log("surl:" + JSON.stringify(this.styleUrl))
  }




  private drawMap() {
    this.map = this.getMap();
    this.map.on('pointermove', (evt: { pixel: any; }) => {
      this.map.forEachFeatureAtPixel(evt.pixel, (feature: FeatureLike) => {
        if (feature) {
          if (this._showObjectInfo) {
            this.curFeature = feature;
            //this.setSelectStyle(this.curFeature)
          }
          this.activeFeature.emit(feature);

        }
      });
    });

    const mapdiv: HTMLElement = this.elementRef.nativeElement.querySelector("[id='map']");
    //console.log('height' + this.elementRef.nativeElement.offsetHeight)  //<<<===here
    //console.log('width' + this.elementRef.nativeElement.offsetWidth)
    this.totalWidth = this.elementRef.nativeElement.offsetWidth;
    this.totalWidth = this.elementRef.nativeElement.offsetHeigh;


    this.map.setTarget(mapdiv);
  }

  private checkParams(): void {
    console.log(this.id)
    if (!this.tileUrl) {
      console.error("No TilteUrl was provided for the app-vectortile-view");
    }
    if (!this.styleUrl) {
      console.log("No StyleUrl was provided for the app-vectortile-view");
    }
    if (!this.zoom) {
      console.error("No zoom was provided for the app-vectortile-view");
    }
    else
      console.log("zoom=" + this.zoom);
    if (!this.centerX) {
      console.error("No zoom center-x was provided for the app-vectortile-view");
    }
    else
      console.log("center-x=" + this.centerX);
    if (!this.centerY) {
      console.error("No center-y was provided for the app-vectortile-view");
    }
    else
      console.log("center-y=" + this.centerY);

  }

  getMap() {
    useGeographic();

    let layers = this.generateLayers();

    let acenter: Coordinate = [this.centerX, this.centerY]
    //console.log("project " + JSON.stringify(this.vectorTileLayer.getSource()?.getProjection()))
    //console.log("axis: " + this.vectorTileLayer.getSource()?.getProjection()?.getAxisOrientation())
    //console.log("acenter=" + acenter)
    return new Map({

      layers: layers,
      view: new View({
        center: acenter,
        zoom: this.zoom,
        maxZoom: this.maxZoom,
        minZoom: this.minZoom,
        enableRotation: false,
        projection: this.vectorTileLayer.getSource()?.getProjection() as ProjectionLike,
      }),
    });
  }

  private generateLayers() {
    this.vectorTileLayer = this.getVectortileLayer(new MapProjection(this.tileUrl).Projection);
    this.setStyle(this.vectorTileLayer);


    let layers = [this.vectorTileLayer] as BaseLayer[] | Collection<BaseLayer> | LayerGroup | undefined;

    if (this.showGrid) {
      const debugLayer = new TileLayer({
        source: new TileDebug({
          template: 'z:{z} y:{y} x:{x}',
          projection: this.vectorTileLayer.getSource()!.getProjection() as ProjectionLike,
          tileGrid: this.vectorTileLayer.getSource()!.getTileGrid() as TileGrid,
          wrapX: this.vectorTileLayer.getSource()!.getWrapX(),
          zDirection: this.vectorTileLayer.getSource()!.zDirection
        }),
      });
      layers = [this.vectorTileLayer, debugLayer];
    }
    return layers;
  }

  private setStyle(vectorTileLayer: VectorTileLayer) {
    if (this.styleUrl) {
      applyStyle(vectorTileLayer, this.styleUrl)
        .then(() => {
          //console.log('style loaded ' + this.styleUrl);

          //overrule source url from style
          if (this.tileUrl !== NetherlandsRDNewQuadDefault) {
            vectorTileLayer.getSource()?.setUrl(this.tileUrl + this.selector);
          }
        })
        .catch((err) => console.error('error loading: ' + this.id + ' ' + this.styleUrl + ' ' + err));
    }
    else {
      const defaultStyle = new Style({
        fill: new Fill({
          color: 'rgba(255,255,255,0.4)',
        }),
        stroke: new Stroke({
          color: '#3399CC',
          width: 1.25,
        })
      })
      vectorTileLayer.setStyle(defaultStyle)


    }
  }

  private setNewProjection() {
    const newLayers = this.generateLayers() as BaseLayer[] | Collection<BaseLayer>;
    const newView = new View({
      center: this.map.getView().getCenter(),
      zoom: this.zoom,
      enableRotation: false,
      projection: this.vectorTileLayer.getSource()?.getProjection() as ProjectionLike,
    });
    this.map.setView(newView);
    this.map.setLayers(newLayers);
    //console.log('project ' + JSON.stringify(this.vectorTileLayer.getSource()?.getProjection()))
  }

  getVectortileLayer(projection: Projection): VectorTileLayer {
    const vectorTileLayer = new VectorTileLayer(
      {
        source: this.getVectorTileSource(projection, this.tileUrl),
        renderMode: 'hybrid',
        declutter: true,
        useInterimTilesOnError: false

      });

    return (vectorTileLayer)
  }

  private calcResolutions(projection: Projection) {
    const tileSizePixels = 256;
    const tileSizeMtrs = getWidth(projection.getExtent()) / tileSizePixels;
    let resolutions: Array<number> = [];
    for (let i = 0; i <= 21; i++) {
      resolutions[i] = tileSizeMtrs / Math.pow(2, i);
    }
    return (resolutions)
  }

  private getVectorTileSource(projection: Projection, url: string) {

    return new VectorTileSource({
      format: new MVT(),
      projection: projection,
      tileGrid: this.tileGrid,
      url: url + this.selector,
      cacheSize: 0
    })
  }

  /*private setSelectStyle(feature: Feature) {
    const selected = new Style({
      fill: new Fill({
        color: '#eeeeee',
      }),
      stroke: new Stroke({
        color: 'rgba(255, 255, 255, 0.7)',
        width: 2,
      }),
    });
    const color = feature.get('COLOR') || '#eeeeee';
    
    feature.setStyle(selected.getFill().setColor(color)!)
  }
  */

  getMapStyle() {
    return ` z-index: 1;
    position: relative;  
    display: flex;"
    width: 300px;
    height: 400px;
    `

  }
}


