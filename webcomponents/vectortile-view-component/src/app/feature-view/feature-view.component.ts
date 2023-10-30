import { Component, ElementRef, Input, NgIterable, OnChanges, OnInit } from '@angular/core'
import { NgChanges } from '../app.component'

import { Feature, Map as OLMap, Tile, VectorTile, View } from 'ol'
import TileLayer from 'ol/layer/Tile'
import { OSM, Vector as VectorSource } from 'ol/source'

import { FeaturesService } from '../openapi/api/features.service'
import { FeatureCollectionGeoJSON } from '../openapi/model/featureCollectionGeoJSON'
import { Vector as VectorLayer } from 'ol/layer'
import GeoJSON from 'ol/format/GeoJSON'
import { Circle, Fill, Stroke, Style } from 'ol/style'
import { Extent, boundingExtent } from 'ol/extent'
import { Geometry, Point, Polygon, SimpleGeometry } from 'ol/geom'
import { FeatureServiceService, DataUrl } from '../feature-service.service'
import { ProjectionLike } from 'ol/proj'

@Component({
  selector: 'app-feature-view',
  templateUrl: './feature-view.component.html',
  styleUrls: ['./feature-view.component.css'],
})
export class FeatureViewComponent implements OnInit, OnChanges {
  @Input() itemsUrl!: string
  @Input() projection: ProjectionLike = 'EPSG:3857'
  mapHeight = 400
  mapWidth = 600
  map: OLMap = this.getMap()
  featureCollectionGeoJSON!: FeatureCollectionGeoJSON
  features: Feature<Geometry>[] = []

  constructor(
    private el: ElementRef,
    private featureService: FeatureServiceService,
    private openApifeaturesService: FeaturesService
  ) {
    //this.OpenApifeaturesService.wegdelenGetFeatures().subscribe(data => {
    //  this.featureCollectionGeoJSON = data
    //  this.loadfeatures()
    // })
  }

  private getMap(): OLMap {
    return new OLMap({
      view: new View({
        projection: this.projection,
      }),
    })
  }

  private init() {
    console.log('height---' + this.el.nativeElement.offsetHeight)
    console.log('width---' + this.el.nativeElement.offsetWidth)
    this.mapWidth = this.el.nativeElement.offsetWidth * 0.99
    this.mapHeight = this.mapWidth * 0.75 // height = 0.75 * width creates 4:3 aspect ratio
    console.log('mapheight---' + this.mapHeight)
    console.log('mapwidth---' + this.mapWidth)
    const mapdiv: HTMLElement = this.el.nativeElement.querySelector("[id='map']")
    // mapdiv.style.height= "100%" //this.mapHeight.toString()
    // mapdiv.style.width= "100%" //this.mapWidth.toString()
    console.log(mapdiv)
    this.map.setTarget(mapdiv)

    console.log('url: ' + this.itemsUrl)
    console.log('projection: ' + this.projection)

    const aurl: DataUrl = { url: this.itemsUrl, projection: this.projection }
    this.featureService.getFeatures(aurl).subscribe(data => {

      this.loadbackground()
      this.features = data
      this.loadfeatures(this.features)
    })
  }


  ngOnChanges(changes: NgChanges<FeatureViewComponent>) {
    if (
      changes.itemsUrl?.previousValue !== changes.itemsUrl?.currentValue ||
      changes.projection.previousValue !== changes.projection.currentValue
    ) {
      console.log('url: ' + changes.itemsUrl?.currentValue)
      this.init()

    }
  }

  ngOnInit() {
    this.init()

  }

  loadbackground() {
    const backgroundLayer = new TileLayer({
      source: new OSM(),
    })
    this.map.addLayer(backgroundLayer)
  }

  loadfeatures(features: Feature<Geometry>[]) {
    const vsource = new VectorSource({
      //features: new GeoJSON().readFeatures(this.featureCollectionGeoJSON, { featureProjection: this.projection  }),
      features: features,
    })

    this.map.addLayer(
      new VectorLayer({
        source: vsource,
        style: this.getstyle(),
        zIndex: 10,
      })
    )

    this.setViewExtent(vsource.getExtent())
  }

  getstyle() {
    const fill = new Fill({
      color: 'rgba(0,0,255)',
    })
    const stroke = new Stroke({
      color: '#3399CC',
      width: 1.25,
    })
    const styles = [
      new Style({
        image: new Circle({
          fill: fill,
          stroke: stroke,
          radius: 5,
        }),
        fill: fill,
        stroke: stroke,
      }),
    ]
    return styles
  }

  setViewExtent(extent: Extent) {
    const view = new View({ extent: extent })
    view.fit(extent)
    this.map.setView(view)
  }
}

export function createFeatureViewComponent(el: ElementRef, featureService: FeatureServiceService, openApifeaturesService: FeaturesService) {
  return new FeatureViewComponent(el, featureService, openApifeaturesService)
}
