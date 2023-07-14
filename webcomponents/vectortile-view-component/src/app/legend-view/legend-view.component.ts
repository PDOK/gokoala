import { Component, ElementRef, Input, OnInit, ViewChild, ViewEncapsulation } from '@angular/core';
import { CommonModule } from '@angular/common';
import { recordStyleLayer, stylefunction } from 'ol-mapbox-style';

import { Style } from 'ol/style';
import VectorTileLayer from 'ol/layer/VectorTile';
import { LegendCfg, LegendItem, MapboxStyleService } from '../mapbox-style.service';
import { Projection } from 'ol/proj';
import { toContext } from 'ol/render';
import CanvasImmediateRenderer from 'ol/render/canvas/Immediate';
import { StyleFunction } from 'ol/style/Style';




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
  @Input() spriteUrl!: string
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
  itemHeight: number = 30
  itemWidth: number = 100

  totalWidth: number = 800
  projection = new Projection({
    code: 'pixel-map',
    units: 'pixels',
    extent: [0, 0, 100, 400],
  });
  layer: VectorTileLayer = new VectorTileLayer({});
  // map: OL.default= new OL.default({layers:[this.Layer], projection: this.projection})
  // });
  @ViewChild('canvas', { static: true })
  canvas?: ElementRef<HTMLCanvasElement>;



  constructor(private mapboxStyleService: MapboxStyleService) {
    recordStyleLayer(true)
  }

  ngOnInit() {
    if (this.styleUrl) {


      this.mapboxStyleService.getMapboxStyle(this.styleUrl).subscribe((mapboxStyle) => {

        if (!this.spriteUrl) {
          this.spriteUrl = mapboxStyle.sprite + '.json'
        }


        this.mapboxStyleService.getMapboxSpriteData(this.spriteUrl).subscribe((spritedata) => {
          let resolutions: number[] = []
          resolutions.push(1)
          const sources = this.mapboxStyleService.getLayersids(mapboxStyle)
          let stfunction = stylefunction(this.layer, this.mapboxStyleService.removefilters(this.mapboxStyleService.removeRasterLayers(mapboxStyle)), sources, resolutions, spritedata, mapboxStyle.glyphs) as StyleFunction;
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
              //const style = getStyleForLayer(item.feature!, resolution, this.layer, item.name)
             // if (item.title !== 'pattern pand') {
              
            //    return
             // }

              // const styleextra  = getStyleForLayer(item.feature!, resolution, this.layer, item.name) as any
           

              const style = stfunction(item.feature!, 1) as Style | Style[]
              if (style) {
                if ((Array.isArray(style))) {
                  item.style = style
                  this.drawItem(item, i, vectorContext, ctx!);
                }
                else {
                  const s: Style[] = []
                  s.push(style)
                  item.style = s
                  this.drawItem(item, i, vectorContext, ctx!);
                }
              }
              else {
                console.warn("no style "+ item.name + ' ' + item.geoType)
                this.drawItem(item, i, vectorContext, ctx!);

              }

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

    //if (item.title == 'Begroeidterreindeel') {
     // console.log('Openbareruimtelabel : ' + item.name + ' ' + item.style.length)
     //  console.log(JSON.stringify(item))
      //  //vectorContext.drawImage(image, dx, dy)


   // }
    if (item.style) {
      let drawOnce= true
      item.style.forEach((style) => {
        vectorContext.setStyle(style)
        const color = style.getRenderer()
        console.log(JSON.stringify(color))
        if (!style.getText())   {
        vectorContext.drawFeature(item.feature!, style);    
        } else
        { if (drawOnce)
          vectorContext.drawFeature(item.feature!, style);  
          drawOnce= false  


        }

        
      })


    }
    else {
      console.log('null draw: ' + item.name)
    }

    ctx.font = 'italic 18px Arial';
    ctx.textAlign = 'left';
    ctx.textBaseline = 'middle';
    ctx.fillStyle = 'black';
    ctx.fillText(item.title, item.labelX, item.labelY!);

  }
}



