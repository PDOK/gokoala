import { ChangeDetectionStrategy, Component, Input, ViewEncapsulation } from '@angular/core'

import RenderFeature from 'ol/render/Feature'
import { WKT } from 'ol/format'

type propRow = {
  title: string
  value: string
}

@Component({
  selector: 'app-object-info',
  encapsulation: ViewEncapsulation.ShadowDom,
  imports: [],
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
          if (val === 'geometry') {
            const wktFormat = new WKT()
            const wktString = wktFormat.writeGeometry(prop[val])
            const geop: propRow = { title: val, value: wktString }
            propTable.push(geop)
          } else {
            const p: propRow = { title: val, value: prop[val] }
            propTable.push(p)
          }
        }
      }

      return propTable
    } else {
      return []
    }
  }
}
