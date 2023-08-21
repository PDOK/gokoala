import { Component, ElementRef, Input, OnInit, ViewEncapsulation } from '@angular/core'
import { CommonModule } from '@angular/common'
import { LegendItemComponent } from '../legend-item/legend-item.component'
import { recordStyleLayer } from 'ol-mapbox-style'
import { IProperties, LegendItem, MapboxStyle, MapboxStyleService } from '../mapbox-style.service'

@Component({
  selector: 'app-legend-view',
  templateUrl: './legend-view.component.html',
  styleUrls: ['./legend-view.component.css'],
  imports: [CommonModule, LegendItemComponent],
  standalone: true,
  encapsulation: ViewEncapsulation.Emulated,
})

export class LegendViewComponent implements OnInit {

  @Input() styleUrl!: string
  @Input() spriteUrl!: string
  @Input() titleItems!: string

  LegendItems: LegendItem[] = []
  mapboxStyle!: MapboxStyle

  constructor(private mapboxStyleService: MapboxStyleService, private elementRef: ElementRef) {
    recordStyleLayer(true)
  }

  ngOnInit() {
    if (this.styleUrl) {
      this.mapboxStyleService.getMapboxStyle(this.styleUrl).subscribe((style) => {
        this.mapboxStyle = this.mapboxStyleService.removeRasterLayers(style)
        if (!this.spriteUrl) {
          this.spriteUrl = this.mapboxStyle.sprite + '.json'
        }
        this.mapboxStyleService.getMapboxSpriteData(this.spriteUrl).subscribe((spritedata) => {
          if (this.titleItems) {
            let titlepart = this.titleItems.split(',')
            this.LegendItems = this.mapboxStyleService.getItems(this.mapboxStyle, this.mapboxStyleService.customTitle, titlepart)
          }
          else {
            this.LegendItems = this.mapboxStyleService.getItems(this.mapboxStyle, this.mapboxStyleService.capitalizeFirstLetter, [])
          }
        })
      })
    }
    else {
      console.error("no style url supplied")
    }
  }
}


