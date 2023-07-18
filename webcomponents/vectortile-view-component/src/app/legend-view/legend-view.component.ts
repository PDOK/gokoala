import { Component, ElementRef, Input, OnInit, ViewChild, ViewEncapsulation } from '@angular/core'
import { CommonModule } from '@angular/common'
import { apply, applyStyle, getStyleForLayer, recordStyleLayer, stylefunction } from 'ol-mapbox-style'

import { Fill, Icon, Stroke, Style, Text as OlText, Circle } from 'ol/style'
import VectorTileLayer from 'ol/layer/VectorTile'
import { LayerType, LegendCfg, LegendItem, MapboxStyle, MapboxStyleService, SpriteData } from '../mapbox-style.service'
import { Projection } from 'ol/proj'
import { toContext } from 'ol/render'
import CanvasImmediateRenderer from 'ol/render/canvas/Immediate'
import { StyleFunction } from 'ol/style/Style'
import { Feature, Map as OLMap, Tile, VectorTile, View } from 'ol'
import { Vector } from 'ol/layer'
import { Vector as VectorSource } from 'ol/source'
import { Control, defaults as defaultControls } from 'ol/control.js'

import VectorLayer from 'ol/layer/Vector'
import { getCenter } from 'ol/extent'
import ImageLayer from 'ol/layer/Image'
import Static from 'ol/source/ImageStatic'
import { MVT } from 'ol/format'
import VectorTileSource from 'ol/source/VectorTile.js'
import { Geometry, Point } from 'ol/geom'
import { EventType } from 'ol/layer/Group'
import BaseEvent from 'ol/events/Event'






@Component({
  selector: 'app-legend-view',
  templateUrl: './legend-view.component.html',
  styleUrls: ['./legend-view.component.css'],
  imports: [CommonModule],
  standalone: true,
  encapsulation: ViewEncapsulation.Emulated,
})



export class LegendViewComponent implements OnInit {
  @Input() styleUrl!: string
  @Input() spriteUrl!: string

  itemHeight: number = 30
  itemWidth: number = 100
  totalHeight: number = 968

  totalWidth: number = 1024

  //extent = [0, 0, this.totalWidth, this.totalHeight]
  extent = [0, 0, 1024, 968];

  projection = new Projection({
    code: 'pixel-map',
    units: 'pixels',
    extent: this.extent
  });


  cvectorSource = new VectorTileSource({
    format: new MVT(),
    projection: this.projection

  });

  cvectorLayer = new VectorTileLayer({
    source: this.cvectorSource,

  });

  LegendItems: LegendItem[] = []





  layer: VectorTileLayer = new VectorTileLayer({});
  // map: OL.default= new OL.default({layers:[this.Layer], projection: this.projection})
  // });
  @ViewChild('canvas', { static: true })
  legendCanvas?: ElementRef<HTMLCanvasElement>
  map: OLMap = new OLMap({})




  constructor(private mapboxStyleService: MapboxStyleService, private elementRef: ElementRef) {
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
          //let stfunction = stylefunction(this.layer, this.mapboxStyleService.removefilters(this.mapboxStyleService.removeRasterLayers(mapboxStyle)), sources, resolutions, spritedata, mapboxStyle.glyphs) as StyleFunction
          const cfg: LegendCfg = {
            "itemHeight": this.itemHeight,
            "itemWidth": this.itemHeight,
            "iconHeight": this.itemHeight * 0.8,
            "iconWidth": this.itemWidth * 0.8,
            "iconOfset": this.itemHeight * 0.1,
            "totalHeight": this.LegendItems.length * this.itemHeight + this.itemHeight * 0.1
          }
          this.LegendItems = this.mapboxStyleService.getItems(mapboxStyle, cfg)


          this.totalHeight = this.LegendItems.length * this.itemHeight + cfg.iconOfset
          const resolution = 1
          let s = this.mapboxStyleService.removefilters(this.mapboxStyleService.removeRasterLayers(mapboxStyle))
          this.drawmap(cfg, this.LegendItems, s, sources, resolutions, spritedata, mapboxStyle.glyphs)
          //  let ctx = this.legendCanvas?.nativeElement.getContext('2d')
          //  if (ctx) {
          // const vectorContext = toContext(ctx, { size: [this.totalWidth, this.totalHeight] })
          /*
         this.LegendItems.forEach((item, i) => {
                          const style  = getStyleForLayer(item.feature!, resolution, this.layer, item.name) as any
          

          
          const style = stfunction(item.feature!, 1) as any
            item.feature?.setStyle(style)

   

            if (style) {
              if ((Array.isArray(style))) {
                item.style = style
                this.drawItem(item, i, vectorContext, ctx!)
              }
              else {
                const s: Style[] = []
                s.push(style)
                item.style = s
                this.drawItem(item, i, vectorContext, ctx!)
              }
            }
            else {
              console.warn("no style " + item.name + ' ' + item.geoType)
              this.drawItem(item, i, vectorContext, ctx!)

            }

          })

        */


          // }






        })

      })

    }
    else {
      console.error("no style url supplied")
    }


  }

  drawmap(cfg: LegendCfg, legendItems: LegendItem[], style: MapboxStyle, sources: string[], resolutions: number[], spritedata: SpriteData, glyphs: string) {


    const mapdiv: HTMLElement = this.elementRef.nativeElement.querySelector("[id='lmap']")
    this.map = new OLMap({
      controls: [],
      interactions: [],

      layers: [
        this.cvectorLayer
      ],
      view: new View({

        projection: this.projection,
        center: getCenter(this.extent),
        zoom: 2,
        minZoom: 2,
        maxZoom: 2
      })
    })

    this.cvectorLayer.getSource()!.setTileLoadFunction((tile: Tile, url) => {
      const vtile = tile as VectorTile
      vtile.setLoader(function (extent, resolution, projection) {
        let features: Feature<Geometry>[] = []

        legendItems.forEach((item, i) => {
          features.push(item.feature!.clone())
      //    features.push(newLegendlabel(item, cfg, i))

        })
        vtile.setFeatures(features)

      })
    })

    this.cvectorLayer.getSource()!.on(["tileloadend"], (evt: any) => {
      let features = evt.tile.getFeatures()
      let ctx = this.legendCanvas?.nativeElement.getContext('2d')
      if (ctx) {
        const vectorContext = toContext(ctx, { size: [this.totalWidth, this.totalHeight] })
        features.forEach((f: Feature, index: number) => {
          //   console.log("feat:" + JSON.stringify(f))
          //    this.LegendItems[index].style = f.clone().getStyle()!
            this.drawItem(this.LegendItems[index], index, vectorContext, ctx!)
        })
      }
    })



    applyStyle(this.cvectorLayer, style, sources, undefined, resolutions)
      .then((mp: OLMap) => {
        console.log(' loading legend style: ' + this.styleUrl)
      })

      .catch((err: any) => {
        console.error(
          'error loading legend style: ' + this.styleUrl + ' ' + err
        )
      })
    //const mapdiv: HTMLElement = this.elementRef.nativeElement.querySelector("[id='lmap']")
    this.map.setTarget(mapdiv)
  }



  drawItem(item: LegendItem, index: number, vectorContext: CanvasImmediateRenderer, ctx: CanvasRenderingContext2D) {

    console.log('y:' + item.labelY)
    //vectorContext.drawGeometry(item.feature?.getGeometry()!)
    ctx.font = 'italic 18px Arial'
    ctx.textAlign = 'left'
    ctx.textBaseline = 'middle'
    ctx.fillStyle = 'black'
    ctx.fillText(item.title, item.labelX, item.labelY!)

    const style = item.style
    if (item.style) {
      if ((Array.isArray(style))) {
        style.forEach((s) => {
          vectorContext.drawFeature(item.feature!, s)

        })
      } else {
        if (typeof style === 'function') {
          console.log('is function')
        }

        else {
          const s: Style[] = []
          s.push(style)
          item.style = s
          vectorContext.drawFeature(item.feature!, style)
        }
      }
    }
    else {
      console.warn("no style " + item.name + ' ' + item.geoType)


    }

  }
}



function newLegendlabel(item: LegendItem, cfg: LegendCfg, y: number) {
  const half = cfg.itemHeight / 2

  const labelStyle: Style = new Style({
    image: new Circle({
      radius: 7,
      fill: new Fill({ color: 'black' }),
      stroke: new Stroke({
        color: [255, 0, 0], width: 2
      })
    })
    ,

    text: new OlText({
      font: '13px Calibri,sans-serif',
      fill: new Fill({
        color: '#000',
      }),
      stroke: new Stroke({
        color: '#fff',
        width: 4,
      }),
    }),
  })

  return new Feature({

    geometry: new Point([cfg.iconWidth + (cfg.iconOfset * 3), cfg.iconOfset + y + half]),
    style: labelStyle

  })
}




