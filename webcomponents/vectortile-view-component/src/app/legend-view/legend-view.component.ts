import { Component, ElementRef, Input, OnInit, ViewEncapsulation } from '@angular/core'
import { CommonModule } from '@angular/common'
import { LegendItemComponent } from '../legend-item/legend-item.component'
import { recordStyleLayer } from 'ol-mapbox-style'
import { IProperties, LegendItem, MapboxStyle, MapboxStyleService } from '../mapbox-style.service'
import { NgChanges } from '../app.component'

@Component({
  selector: 'app-legend-view',
  templateUrl: './legend-view.component.html',
  styleUrls: ['./legend-view.component.css'],
  imports: [CommonModule, LegendItemComponent],
  standalone: true,
  encapsulation: ViewEncapsulation.Emulated,
})

export class LegendViewComponent implements OnInit {

  mapboxStyle!: MapboxStyle
  @Input() styleUrl!: string
  @Input() titleItems!: string

  LegendItems: LegendItem[] = []


  constructor(private mapboxStyleService: MapboxStyleService, private elementRef: ElementRef) {
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





  generateLegend() {
    if (this.styleUrl) {
      this.mapboxStyleService.getMapboxStyle(this.styleUrl).subscribe((style) => {
        this.mapboxStyle = this.mapboxStyleService.removeRasterLayers(style)
        if (this.titleItems) {
          let titlepart = this.titleItems.split(',')
          this.LegendItems = this.mapboxStyleService.getItems(this.mapboxStyle, this.mapboxStyleService.customTitle, titlepart)
        }
        else {
          this.LegendItems = this.mapboxStyleService.getItems(this.mapboxStyle, this.mapboxStyleService.capitalizeFirstLetter, [])
        }

      })
    }
    else {
      console.error("no style url supplied")
    }
  }

}


