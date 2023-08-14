import { Component, Input, OnInit, ElementRef } from '@angular/core'
import { Feature, Map as OLMap, Tile, VectorTile, View } from 'ol'
import { Projection } from 'ol/proj'
import { MVT } from 'ol/format'
import VectorTileSource from 'ol/source/VectorTile.js'
import VectorTileLayer from 'ol/layer/VectorTile'
import { getCenter } from 'ol/extent'
import { Geometry, LineString, Point, Polygon } from 'ol/geom'
import { LayerType, LegendItem, MapboxStyle, MapboxStyleService, exhaustiveGuard } from '../mapbox-style.service'
import { applyStyle } from 'ol-mapbox-style'

type LegendCfg = {
  iconOfset: number
  iconWidth: number
  iconHeight: number

}

@Component({
  selector: 'app-legend-item',
  templateUrl: './legend-item.component.html',
  styleUrls: ['./legend-item.component.css'],
  standalone: true,
})
export class LegendItemComponent implements OnInit {

  constructor(private mapboxStyleService: MapboxStyleService, private elementRef: ElementRef) {

  }

  @Input() item!: LegendItem
  @Input() mapboxStyle!: MapboxStyle

  itemHeight: number = 30
  itemWidth: number = 60
  itemLeft: number = 10
  itemRight: number = 50
  extent = [0, 0, this.itemWidth, this.itemHeight];

  projection = new Projection({
    code: 'pixel-map',
    units: 'pixels',
    extent: this.extent
  });

  map: OLMap = new OLMap({})
  cvectorSource = new VectorTileSource({
    format: new MVT(),
    projection: this.projection

  });

  cvectorLayer = new VectorTileLayer({
    source: this.cvectorSource,

  });


  ngOnInit() {

    let feature = this.NewFeature(this.item)
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


        features.push(feature)
        vtile.setFeatures(features)

      })
    })


    let resolutions: number[] = []
    resolutions.push(1)
    const sources = this.mapboxStyleService.getLayersids(this.mapboxStyle)

    applyStyle(this.cvectorLayer, this.mapboxStyle, sources, undefined, resolutions)
      .then((mp: OLMap) => {
        console.log(' loading legend style')
      })

      .catch((err: any) => {
        console.error(
          'error loading legend style: ' + ' ' + err
        )
      })
    this.cvectorLayer.getSource()?.refresh()
    const mapdiv: HTMLElement = this.elementRef.nativeElement.querySelector("[id='itemmap']")
    this.map.setTarget(mapdiv)
  }

  NewFeature(item: LegendItem): Feature {
    const cfg: LegendCfg = {
      "iconHeight": this.itemHeight * 0.6,
      "iconWidth": this.itemWidth * 0.8,
      "iconOfset": this.itemWidth * 0.1
    }
    const half = this.itemHeight / 2
    switch (item.geoType) {
      case LayerType.Fill: {
        let f = new Feature({

          geometry: new Polygon([
            [[cfg.iconOfset, cfg.iconOfset], [cfg.iconWidth, cfg.iconOfset], [cfg.iconWidth, cfg.iconHeight], [cfg.iconOfset, cfg.iconHeight], [cfg.iconOfset, cfg.iconOfset]]
          ]),
          layer: item.sourceLayer


        })
        f.setProperties(item.properties)
        return f
      }
      case LayerType.Circle:
      case LayerType.Raster:
      case LayerType.Symbol: {
        let f = new Feature({
          geometry: new Point(getCenter(this.extent)),
          layer: item.sourceLayer
        })
        f.setProperties(item.properties)
        return f
      }
      case LayerType.Line: {
        let f = new Feature({
          geometry: new LineString([[this.itemLeft, half], [this.itemRight, half]]),
          layer: item.sourceLayer
        })
        f.setProperties(item.properties)
        return f
      } default: {
        exhaustiveGuard(item.geoType)
      }
    }
  }
}



