import {
  Component,
  OnInit,
  Input,
  ElementRef,
} from '@angular/core';


import OGCVectorTile from 'ol/source/OGCVectorTile.js';
import VectorTileSource from 'ol/source/VectorTile.js';
import Map from 'ol/Map';
import View from 'ol/View';
import { MapProjection } from '../app/mapprojection'


import Projection from 'ol/proj/Projection';
import { Fill, Stroke, Style } from "ol/style";
import { MVT } from "ol/format";
import VectorTileLayer from 'ol/layer/VectorTile';
import { getTopLeft, getWidth } from 'ol/extent';
import TileGrid from 'ol/tilegrid/TileGrid';
import { ProjectionLike, useGeographic } from 'ol/proj';
import { Coordinate } from 'ol/coordinate';


@Component({
  selector: 'app-vectortile-view',
  templateUrl: './app.component.html',
  styleUrls: ['./app.component.css'],
  // encapsulation: ViewEncapsulation.ShadowDom
})
export class // encapsulation: ViewEncapsulation.ShadowDom
  AppComponent implements OnInit {

  title = 'vectortile-view-component';
  map = new Map({});

  @Input() tileUrl!: string
  @Input() styleUrl!: string
  @Input() zoom!: number
  @Input() centerX!: number;
  @Input() centerY!: number;

  constructor(private elementRef: ElementRef
  ) {
  }

  ngOnInit() { 
    this.checkParams();
    this.map = this.getMap()
    this.map.setTarget(this.elementRef.nativeElement);
  }

  private checkParams(): void {
    if (!this.tileUrl) {
      console.error("No TilteUrl was provided for the app-vectortile-view");
    }
    if (!this.styleUrl) {
      console.error("No StyleUrl was provided for the app-vectortile-view");
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

    const vectorTileLayer = this.getVectortileLayer(new MapProjection(this.tileUrl).Projection, style)
    console.log(JSON.stringify(new OGCVectorTile({
      url: this.tileUrl,
      format: new MVT(),
    }) as any))

    let acenter: Coordinate = [this.centerX, this.centerY]
    console.log("project " + JSON.stringify(vectorTileLayer.getSource()?.getProjection()))
    console.log("axis: " + vectorTileLayer.getSource()?.getProjection()?.getAxisOrientation())
    console.log("acenter=" + acenter)
    return new Map({
      target: 'app-vectortile-view',
      layers: [vectorTileLayer],
      view: new View({
        center: acenter,
        zoom: this.zoom,
        enableRotation: false,
        projection: vectorTileLayer.getSource()?.getProjection() as ProjectionLike,
      }),
    });
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
    let selector = '/{z}/{y}/{x}?f=mvt'
    if (url.includes('pdok')) {
      selector = '/{z}/{x}/{y}?f=mvt'
    }
    return new VectorTileSource({
      format: new MVT(),
      projection: projection,
      tileGrid: new TileGrid({
        extent: projection.getExtent(),
        resolutions: this.calcResolutions(projection),
        tileSize: [256, 256],
        origin: getTopLeft(projection.getExtent())
      }),
      url: url + selector,
      cacheSize: 0
    })
  }
}
