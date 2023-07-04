import { AfterViewInit, Component, ElementRef, Input, OnInit, ViewChild, ViewEncapsulation } from '@angular/core';
import { getStyleForLayer, applyStyle, recordStyleLayer } from 'ol-mapbox-style';

import { Style, Fill, Stroke } from 'ol/style';
import { NetherlandsRDNewQuadDefault } from '../mapprojection';
import VectorTileLayer from 'ol/layer/VectorTile';
import VectorLayer from 'ol/layer/Vector';
import { Feature } from 'ol';
import { Geometry, LineString, Point, Polygon } from 'ol/geom';
import { MapboxStyle, MapboxStyleService } from '../mapbox-style.service';
import * as OL from 'ol/Map';
import { Vector } from 'ol/source';
import VectorSource from 'ol/source/Vector';
import { Projection } from 'ol/proj';
import { toContext } from 'ol/render';
import CircleStyle from 'ol/style/Circle';
import VectorContext from 'ol/render/VectorContext';
import CanvasImmediateRenderer from 'ol/render/canvas/Immediate';
import { StyleLike } from 'ol/style/Style';





type LegendItem = {
  name: string,
  labelX: number,
  labelY: number
  itemStyle: Style[],
  feature: Feature

}


@Component({
  selector: 'app-legend-view',
  templateUrl: './legend-view.component.html',
  styleUrls: ['./legend-view.component.css'],
  standalone: true,
  encapsulation: ViewEncapsulation.ShadowDom,
})
export class LegendViewComponent implements OnInit, AfterViewInit {
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
  totalHeight: number = 600
  itemHeight: number = 0
  itemWidth: number = 0

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

  ngOnInit(): void {
    if (this.styleUrl) {
      this.mapboxStyleService.getMapboxStyle(this.styleUrl).subscribe((jsonStyle) => {

        jsonStyle.layers.forEach(layer=> 
          {
            layer.id = layer['source-layer']

          }

        )
        const allsourcelayers: string[] = [...new Set(jsonStyle.layers.map(
          layer => layer.id
        ))]
        const sourcelayers = allsourcelayers.sort()
        console.log(sourcelayers)
        this.itemHeight = this.totalHeight / sourcelayers.length
        this.itemWidth = this.itemHeight
        const itemHalf = this.itemHeight / 2
        const iconHeight = this.itemHeight * 0.8
        const iconWidth = this.itemWidth * 0.8
        const iconOfset = this.itemHeight * 0.1

        const resolution = 19.109257071294063
        applyStyle(this.Layer, this.styleUrl)
          .then((x) => {

            let ctx = this.canvas?.nativeElement.getContext('2d');
            if (ctx) {
              const vectorContext = toContext(ctx, { size: [this.totalWidth, this.totalHeight] });
              sourcelayers.forEach((el, index) => {
                const itemName = el
                console.log("layer:" + itemName + index)
                const y = this.itemHeight * (index + 1)

                let Polyfeature = new Feature({
                  geometry: new Polygon([
                    [[iconOfset, iconOfset + y], [iconWidth, iconOfset + y], [iconWidth, iconHeight + y], [iconOfset, iconHeight + y], [iconOfset, iconOfset + y]]]

                  )
                }

                )
                Polyfeature.setProperties({ 'layer': itemName })


                //  this.Layer.getSource().
                const Polystyle = getStyleForLayer(Polyfeature, resolution, this.Layer, itemName)
                if (Polystyle) {
                  // console.log(Polystyle[0].getFill().getColor())
                  drawItem({ name: itemName, labelX: this.itemWidth * 1.1, labelY: y + itemHalf, itemStyle: [...Polystyle], feature: Polyfeature }, vectorContext, ctx!);
                  // this.LegendItems.push({ name: itemName, itemStyle: [...Polystyle], feature: Polyfeature })
                }
                else {
                  let PointFeature = new Feature({
                    geometry: new Point([iconOfset, iconOfset + y]),
                  })
                  PointFeature.setProperties({ 'layer': itemName })

                  const Pointstyle = getStyleForLayer(PointFeature, resolution, this.Layer, itemName)
                  if (Pointstyle) {
                    drawItem({ name: itemName, labelX: this.itemWidth * 1.1, labelY: y + itemHalf, itemStyle: [...Pointstyle], feature: PointFeature }, vectorContext, ctx!);

                    //   this.LegendItems.push({ name: itemName, itemStyle: [...Pointstyle], feature: PointFeature })
                  }
                  else {
                    let lineFeature = new Feature({
                      geometry: new LineString([[iconOfset, iconOfset + y], [iconWidth, iconOfset + y]],),
                    })
                    lineFeature.setProperties({ 'layer': itemName })

                    const linestyle = getStyleForLayer(lineFeature, resolution, this.Layer, itemName)
                    if (linestyle) {
                      // this.LegendItems.push({ name: itemName, itemStyle: [...linestyle], feature: lineFeature })
                      drawItem({ name: itemName, labelX: this.itemWidth * 1.1, labelY: y + itemHalf, itemStyle: [...linestyle], feature: lineFeature }, vectorContext, ctx!);
                    }

                    else {
                      console.log("no style for " + itemName)

                    }
                  }



                }

              });


            }







          })
        //   .catch((err) => console.error('error loading: ' + this.styleUrl + ' ' + err));
      })
    }
  }

  ngAfterViewInit(): void {


  }

  private drawLegend(itemHeight: number = 100) {


    let ctx = this.canvas?.nativeElement.getContext('2d');
    if (ctx) {

      const vectorContext = toContext(ctx, { size: [this.totalWidth, this.totalHeight] });

      this.LegendItems.forEach((item, index) => {
        console.log(item.name)
        if (item.itemStyle) {
          /* item.style.forEach((style : StyleLike, index) => {
             if (style) {
               drawItem(item.geometry as Geometry, style, vectorContext);
             }
             else {
               console.log("nog style array for " + item.name)
             }
           })
           */

        }








        /*
        
              
                */



      })
    }

    else {
      console.log("no canvas")
    }




  }


}



function drawItem(item: LegendItem, vectorContext: CanvasImmediateRenderer, ctx: CanvasRenderingContext2D) {
  vectorContext.drawFeature(item.feature, item.itemStyle[item.itemStyle.length - 1]);
  ctx!.font = 'italic 18px Arial';
  ctx!.textAlign = 'left';
  ctx!.textBaseline = 'middle';
  ctx!.fillStyle = 'black';
  ctx!.fillText(item.name, item.labelX, item.labelY);

}

