import { ChangeDetectionStrategy, Component, Input, ViewEncapsulation } from '@angular/core'
import { CommonModule } from '@angular/common'
import RenderFeature from 'ol/render/Feature'

type propRow = {
  title: string
  value: string
}

@Component({
  selector: 'app-object-info',
  encapsulation: ViewEncapsulation.ShadowDom,
  imports: [CommonModule],
  changeDetection: ChangeDetectionStrategy.OnPush,
  templateUrl: './object-info.component.html',
  styleUrls: ['./object-info.component.css'],
})
export class ObjectInfoComponent {
  @Input() feature!: RenderFeature

  public getFeatureProperties(): propRow[] {
    const propTable: propRow[] = []
    if (this.feature) {
      const prop = this.feature.getProperties()

      for (const val in prop) {
        if (val !== 'mapbox-layer') {
          const p: propRow = { title: val, value: prop[val] }
          propTable.push(p)
        }
      }
      return propTable
    } else {
      return []
    }
  }
}
