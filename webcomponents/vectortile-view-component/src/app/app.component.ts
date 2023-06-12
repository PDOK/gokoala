import {
  Component,
  OnInit,
  Input,
  ElementRef,
  SimpleChanges,
  Output,
  EventEmitter,
  CUSTOM_ELEMENTS_SCHEMA 
} from '@angular/core';
import { coerceBooleanProperty } from '@angular/cdk/coercion';
import { Subject } from 'rxjs';

import {getUid} from 'ol/util';

import Select from 'ol/interaction/Select.js';
import { altKeyOnly, click, pointerMove } from 'ol/events/condition.js';
import VectorTileSource from 'ol/source/VectorTile.js';
import TileDebug from 'ol/source/TileDebug.js';
import Map from 'ol/Map';
import View from 'ol/View';
import { MapProjection, NetherlandsRDNewQuadDefault } from '../app/mapprojection'

import { applyStyle, apply } from 'ol-mapbox-style';


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
import { Feature } from 'ol';
import { StyleFunction } from 'ol/style/Style';
import { FeatureLike } from 'ol/Feature';
import RenderFeature from 'ol/render/Feature';
import { ObjectInfoComponent } from './object-info/object-info.component';





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
  // encapsulation: ViewEncapsulation.ShadowDom
  standalone: true,
   imports: [ObjectInfoComponent],
   schemas: [
    CUSTOM_ELEMENTS_SCHEMA // Tells Angular we will have custom tags in our templates
  ]
 
})
export class // encapsulation: ViewEncapsulation.ShadowDom
  AppComponent implements OnInit {

  title = 'vectortile-view-component';
  map = new Map({});
  selector = '/{z}/{y}/{x}?f=mvt'


  @Input() tileUrl: string = NetherlandsRDNewQuadDefault
  @Input() styleUrl!: string | undefined
  @Input() zoom!: number
  @Input() centerX!: number;
  @Input() centerY!: number;
  private _showGrid: boolean = false;
  vectorTileLayer!: VectorTileLayer;
  curFeature:Feature| undefined= undefined;
  @Input()
  set showGrid(showGrid: any) {
    this._showGrid = coerceBooleanProperty(showGrid);
  }
  get showGrid() {
    return this._showGrid
  }

  @Output() activeFeature = new EventEmitter<RenderFeature>();

 


  constructor( private elementRef:ElementRef) {

   
   
  }

  ngOnChanges(changes: NgChanges<AppComponent>) {
    if (changes.styleUrl.previousValue !== changes.styleUrl.currentValue) {
      console.log('changed')
      this.setStyle(this.vectorTileLayer);
    }
  }

  ngOnInit() {
    this.checkParams();

    this.map = this.getMap()
    const mapdiv= this.elementRef.nativeElement.querySelector("[id='map']")
    this.map.setTarget(mapdiv);


    this.map.on('pointermove', (evt: { pixel: any; }) => {
      this.map.forEachFeatureAtPixel(evt.pixel,  (feature: any, layer: any) => {
      
        this.curFeature = feature
        this.activeFeature.emit(feature)
     
        //console.log(layer);
      });

    })
  }



  private checkParams(): void {
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
    const style = new Style({
      fill: new Fill({
        color: 'rgba(255,255,255,0.4)',
      }),
      stroke: new Stroke({
        color: '#3399CC',
        width: 1.25,
      })
    })

    this.vectorTileLayer = this.getVectortileLayer(new MapProjection(this.tileUrl).Projection, style)
    this.setStyle(this.vectorTileLayer);


    let layers = [this.vectorTileLayer] as BaseLayer[] | Collection<BaseLayer> | LayerGroup | undefined

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
      layers = [this.vectorTileLayer, debugLayer]
    }

    let acenter: Coordinate = [this.centerX, this.centerY]
    console.log("project " + JSON.stringify(this.vectorTileLayer.getSource()?.getProjection()))
    console.log("axis: " + this.vectorTileLayer.getSource()?.getProjection()?.getAxisOrientation())
    console.log("acenter=" + acenter)
    return new Map({
    
      layers: layers,
      view: new View({
        center: acenter,
        zoom: this.zoom,
        enableRotation: false,
        projection: this.vectorTileLayer.getSource()?.getProjection() as ProjectionLike,
      }),
    });
  }

  private setStyle(vectorTileLayer: VectorTileLayer) {
   // if (this.styleUrl) {
      applyStyle(vectorTileLayer, this.styleUrl)
        .then(() => {
          console.log('style loaded ' + this.styleUrl);

          //overrule source url from style
          if (this.tileUrl !== NetherlandsRDNewQuadDefault) {
            vectorTileLayer.getSource()?.setUrl(this.tileUrl + this.selector);
          }
        })
        .catch(() => console.log('error loading: ' + this.styleUrl));
   // }
  }

  getVectortileLayer(projection: Projection, style: Style) {
    const vectorTileLayer = new VectorTileLayer(
      {
        source: this.getVectorTileSource(projection, this.tileUrl),
        renderMode: 'hybrid',
        declutter: true,
        useInterimTilesOnError: false,
        style: style
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
      tileGrid: new TileGrid({
        extent: projection.getExtent(),
        resolutions: this.calcResolutions(projection),
        tileSize: [256, 256],
        origin: getTopLeft(projection.getExtent())
      }),
      url: url + this.selector,
      cacheSize: 0
    })
  }

  private selectStyle(feature: FeatureLike, resolution: number): Style {
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
    selected.getFill().setColor(color);
    return selected;
  }
}
