import { Component, Input, OnChanges, OnInit, ViewEncapsulation, inject } from '@angular/core'
import { recordStyleLayer } from 'ol-mapbox-style'
import { NgChanges } from '../vectortile-view/vectortile-view.component'
import { LegendItemComponent } from './legend-item/legend-item.component'
import { LegendItem, MapboxStyle, MapboxStyleService } from '../mapbox-style.service'
import { NGXLogger } from 'ngx-logger'

@Component({
  selector: 'app-legend-view',
  templateUrl: './legend-view.component.html',
  styleUrls: ['./legend-view.component.css'],
  imports: [LegendItemComponent],
  standalone: true,
  encapsulation: ViewEncapsulation.Emulated,
})
export class LegendViewComponent implements OnInit, OnChanges {
  private logger = inject(NGXLogger)
  private mapboxStyleService = inject(MapboxStyleService)

  mapboxStyle!: MapboxStyle
  // URL to a Mapbox style JSON endpoint
  @Input() styleUrl!: string
  /*
  This input is used to specify the attributes for legend items. By default, layers are used for legend items.
  Attributes can be specified to create distinct items. For example, for the Dutch BGT, you can use
  titleItems = "type,plus_type,functie,fysiek_voorkomen,openbareruimtetype".
  Refer to: https://github.com/PDOK/vectortile-demo-viewer/blob/a8b49378bcdeef7196aabd1d34402d409421121f/projects/vectortile-demo/src/app/mapstyler/mapstyler.component.ts#L159C15-L159C75
  If titleItems = "id" is used, the "id" (i.e., the layer name from the style JSON) is used.
  */
  @Input() titleItems!: string

  LegendItems: LegendItem[] = []

  constructor() {
    recordStyleLayer(true)
  }

  ngOnChanges(changes: NgChanges<LegendViewComponent>) {
    if (changes.styleUrl?.previousValue !== changes.styleUrl?.currentValue) {
      if (!changes.styleUrl.isFirstChange()) {
        this.generateLegend()
      }
    }
    if (this.titleItems) {
      if (changes.titleItems.previousValue !== changes.titleItems.currentValue) {
        if (!changes.titleItems.isFirstChange()) {
          this.generateLegend()
        }
      }
    }
  }

  ngOnInit(): void {
    this.generateLegend()
  }

  private generateLegend() {
    if (this.styleUrl) {
      this.mapboxStyleService.getMapboxStyle(this.styleUrl).subscribe(style => {
        this.mapboxStyle = this.mapboxStyleService.removeRasterLayers(style)
        if (this.mapboxStyle.metadata?.['gokoala:title-items']) {
          this.titleItems = this.mapboxStyle.metadata?.['gokoala:title-items']
        }
        const titlepart = this.titleItems ? this.titleItems.split(',') : []
        if (this.titleItems) {
          if (this.titleItems.toLocaleLowerCase() === 'id') {
            this.LegendItems = this.mapboxStyleService.getItems(this.mapboxStyle, this.mapboxStyleService.idTitle, titlepart, true)
          } else if (this.titleItems.toLocaleLowerCase().includes('no-id')) {
            this.LegendItems = this.mapboxStyleService.getItems(this.mapboxStyle, this.mapboxStyleService.customTitle, titlepart, false)
          } else {
            this.LegendItems = this.mapboxStyleService.getItems(this.mapboxStyle, this.mapboxStyleService.customTitle, titlepart, true)
          }
        } else {
          this.LegendItems = this.mapboxStyleService.getItems(this.mapboxStyle, this.mapboxStyleService.capitalizeFirstLetter, [], true)
        }
      })
    } else {
      this.logger.error('no style url supplied')
    }
  }
}
