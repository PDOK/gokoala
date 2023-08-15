import { Component, ElementRef, Input, OnInit, ViewEncapsulation } from '@angular/core'
import { CommonModule } from '@angular/common'
import { LegendItemComponent } from '../legend-item/legend-item.component'
import { recordStyleLayer } from 'ol-mapbox-style'
import { LegendItem, MapboxStyle, MapboxStyleService } from '../mapbox-style.service'

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
  LegendItems: LegendItem[] = []
  mapboxStyle!: MapboxStyle

  constructor(private mapboxStyleService: MapboxStyleService, private elementRef: ElementRef) {
    recordStyleLayer(true)
  }

  ngOnInit() {
    if (this.styleUrl) {
      this.mapboxStyleService.getMapboxStyle(this.styleUrl).subscribe((style) => {
        this.mapboxStyle=this.mapboxStyleService.removefilters(style)
        if (!this.spriteUrl) {
          this.spriteUrl = this.mapboxStyle.sprite + '.json'
        }
        this.mapboxStyleService.getMapboxSpriteData(this.spriteUrl).subscribe((spritedata) => {
          this.LegendItems = this.mapboxStyleService.getItems(this.mapboxStyle)
        })
      })
    }
    else {
      console.error("no style url supplied")
    }
  }

}


