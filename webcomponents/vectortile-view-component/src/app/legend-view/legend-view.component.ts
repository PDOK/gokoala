import { AfterViewInit, Component, Directive, ElementRef, Input, OnInit, QueryList, ViewChild, ViewChildren, ViewEncapsulation } from '@angular/core';
import { CommonModule } from '@angular/common';
import apply, { getStyleForLayer, applyStyle, recordStyleLayer } from 'ol-mapbox-style';

import { Style, Fill, Stroke } from 'ol/style';
import { NetherlandsRDNewQuadDefault } from '../mapprojection';
import VectorTileLayer from 'ol/layer/VectorTile';
import VectorLayer from 'ol/layer/Vector';
import { Feature } from 'ol';
import { Geometry, LineString, Point, Polygon } from 'ol/geom';
import { LegendCfg, LegendItem, MapboxStyle, MapboxStyleService } from '../mapbox-style.service';
import * as OL from 'ol/Map';
import { Vector } from 'ol/source';
import VectorSource from 'ol/source/Vector';
import { Projection } from 'ol/proj';
import { toContext } from 'ol/render';
import CircleStyle from 'ol/style/Circle';
import VectorContext from 'ol/render/VectorContext';
import CanvasImmediateRenderer from 'ol/render/canvas/Immediate';
import { StyleLike } from 'ol/style/Style';







@Component({
  selector: 'app-legend-view',
  templateUrl: './legend-view.component.html',
  styleUrls: ['./legend-view.component.css'],
  imports: [CommonModule],
  standalone: true,
  encapsulation: ViewEncapsulation.ShadowDom,
})



export class LegendViewComponent implements OnInit {
  @Input() styleUrl!: string
  vectorsource = {
    'geojson': {
      type: 'geojson',
      data: {
        type: 'FeatureCollection',
        features: []
      }
    }
  }




  LegendItems: LegendItem[] = []
  totalHeight: number = 11600
  itemHeight: number = 100
  itemWidth: number = 100

  totalWidth: number = 800
  projection = new Projection({
    code: 'pixel-map',
    units: 'pixels',
    extent: [0, 0, 100, 400],
  });
  Layer: VectorTileLayer = new VectorTileLayer({});
  // map: OL.default= new OL.default({layers:[this.Layer], projection: this.projection})
  // });
  @ViewChild('canvas', { static: true })
  canvas?: ElementRef<HTMLCanvasElement>;



  constructor(private mapboxStyleService: MapboxStyleService) {
    recordStyleLayer(true)
  }

  ngOnInit() {
    if (this.styleUrl) {
      applyStyle(this.Layer, this.styleUrl)
        .then((x) => {
          this.mapboxStyleService.getMapboxStyle(this.styleUrl).subscribe((mapboxStyle) => {

            const cfg: LegendCfg = {
              "itemHeight": this.itemHeight,
              "itemWidth": this.itemHeight,
              "iconHeight": this.itemHeight * 0.8,
              "iconWidth": this.itemWidth * 0.8,
              "iconOfset": this.itemHeight * 0.1,

            }
            this.LegendItems = this.mapboxStyleService.getItems(mapboxStyle, cfg)
            this.totalHeight = this.LegendItems.length * this.itemHeight + cfg.iconOfset
            const resolution = 1
            let ctx = this.canvas?.nativeElement.getContext('2d');
            if (ctx) {
              const vectorContext = toContext(ctx, { size: [this.totalWidth, this.totalHeight] });
              this.LegendItems.forEach((item, i) => {
                console.log(item.name)
              //  item.style = getStyleForLayer(item.feature, resolution, this.Layer, item.name)
                this.drawItem(item, i, vectorContext, ctx!);

              })
            }
          })
        })
    }
    else {
      console.error("no style url supplied")
    }
  }


  drawItem(item: LegendItem, index: number, vectorContext: CanvasImmediateRenderer, ctx: CanvasRenderingContext2D) {
    console.log('draw: ' + item.name)
    if (item.style) {
      item.style.forEach((style) => {
        vectorContext.drawFeature(item.feature, style);

      })

    }

    ctx!.font = 'italic 18px Arial';
    ctx!.textAlign = 'left';
    ctx!.textBaseline = 'middle';
    ctx!.fillStyle = 'black';
    ctx!.fillText(item.name, item.labelX, item.labelY);

  }
}



