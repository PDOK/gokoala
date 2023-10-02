import { ChangeDetectionStrategy, Component, ElementRef, Input, OnInit } from '@angular/core';
import { Feature, Map as OLMap, Tile, VectorTile, View } from 'ol';
import { Projection } from 'ol/proj';
import { MVT } from 'ol/format';
import VectorTileSource from 'ol/source/VectorTile.js';
import VectorTileLayer from 'ol/layer/VectorTile';
import { getCenter } from 'ol/extent';
import { Geometry, LineString, Point } from 'ol/geom';
import { exhaustiveGuard, LayerType, LegendItem, MapboxStyle, MapboxStyleService } from '../mapbox-style.service';
import { applyStyle } from 'ol-mapbox-style';
import { fromExtent } from 'ol/geom/Polygon';

@Component({
  selector: 'app-legend-item',
  templateUrl: './legend-item.component.html',
  styleUrls: ['./legend-item.component.css'],
  standalone: true,
  changeDetection: ChangeDetectionStrategy.OnPush,
})
export class LegendItemComponent implements OnInit {
  constructor(
    private mapboxStyleService: MapboxStyleService,
    private elementRef: ElementRef
  ) {}

  @Input() item!: LegendItem;
  @Input() mapboxStyle!: MapboxStyle;

  itemHeight = 40;
  itemWidth = 60;
  itemLeft = 10;
  itemRight = 50;
  extent = [0, 0, this.itemWidth, this.itemHeight];

  projection = new Projection({
    code: 'pixel-map',
    units: 'pixels',
    extent: this.extent,
  });

  map: OLMap = new OLMap({});
  cvectorSource = new VectorTileSource({
    format: new MVT(),
    projection: this.projection,
  });

  cvectorLayer = new VectorTileLayer({
    source: this.cvectorSource,
  });

  ngOnInit() {
    const feature = this.NewFeature(this.item);
    this.map = new OLMap({
      controls: [],
      interactions: [],

      layers: [this.cvectorLayer],
      view: new View({
        projection: this.projection,
        center: getCenter(this.extent),
        zoom: 2,
        minZoom: 2,
        maxZoom: 2,
      }),
    });

    this.cvectorLayer.getSource()?.setTileLoadFunction((tile: Tile) => {
      const vtile = tile as VectorTile;
      vtile.setLoader(() => {
        const features: Feature<Geometry>[] = [];

        features.push(feature);
        vtile.setFeatures(features);
      });
    });

    const resolutions: number[] = [];
    resolutions.push(1);
    const sources = this.mapboxStyleService.getLayersids(this.mapboxStyle);

    applyStyle(this.cvectorLayer, this.mapboxStyle, sources, undefined, resolutions)
      .then(() => {
        console.log(' loading legend style');
      })
      // eslint-disable-next-line @typescript-eslint/no-explicit-any
      .catch((err: any) => {
        console.error('error loading legend style: ' + ' ' + err);
      });
    this.cvectorLayer.getSource()?.refresh();
    const mapdiv: HTMLElement = this.elementRef.nativeElement.querySelector("[id='itemmap']");
    this.map.setTarget(mapdiv);
  }

  NewFeature(item: LegendItem): Feature {
    const half = this.itemHeight / 2;
    switch (item.geoType) {
      case LayerType.Fill: {
        const ageom = fromExtent(this.extent);
        ageom.scale(0.05, 0.05);
        const f = new Feature({
          geometry: ageom,
          layer: item.sourceLayer,
        });
        f.setProperties(item.properties);
        return f;
      }
      case LayerType.Circle:
      case LayerType.Raster:
      case LayerType.Symbol: {
        const f = new Feature({
          geometry: new Point(getCenter(this.extent)),
          layer: item.sourceLayer,
        });
        f.setProperties(item.properties);
        return f;
      }
      case LayerType.Line: {
        const f = new Feature({
          geometry: new LineString([
            [this.itemLeft, half],
            [this.itemRight, half],
          ]),
          layer: item.sourceLayer,
        });
        f.setProperties(item.properties);
        return f;
      }
      default: {
        exhaustiveGuard(item.geoType);
      }
    }
  }
}
