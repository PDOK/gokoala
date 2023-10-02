import { Component, Input, ViewEncapsulation } from '@angular/core';
import { CommonModule } from '@angular/common';
import Feature, { FeatureLike } from 'ol/Feature';
import RenderFeature from 'ol/render/Feature';

type proprow = {
  title: string;
  value: string;
};

@Component({
  selector: 'app-object-info',
  standalone: true,
  encapsulation: ViewEncapsulation.ShadowDom,
  imports: [CommonModule],
  templateUrl: './object-info.component.html',
  styleUrls: ['./object-info.component.css'],
})
export class ObjectInfoComponent {
  @Input() feature!: RenderFeature;

  public getFeatureProperties(): proprow[] {
    let proptable: proprow[] = [];
    if (this.feature) {
      const prop = this.feature.getProperties();

      for (const val in prop) {
        if (val !== 'mapbox-layer') {
          const p: proprow = { title: val, value: prop[val] };
          proptable.push(p);
        }
      }
      return proptable;
    } else {
      // console.log('feature undefined')
      return [];
    }
  }
}
