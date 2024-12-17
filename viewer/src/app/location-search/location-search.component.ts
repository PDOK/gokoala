import { Component, Input } from '@angular/core'

import { CommonModule } from '@angular/common'

import { CollectionsService } from '../api/services'
import { NGXLogger } from 'ngx-logger'
import { LocationSearchService } from '../location-search.service'
import { defaultMapping, ProjectionMapping } from '../feature.service'
import { Feature } from 'ol'
import { Geometry } from 'ol/geom'
import { Observable } from 'rxjs'
import { SafeHtmlPipe } from '../safe-html.pipe'
import { BackgroundMap, FeatureViewComponent } from '../feature-view/feature-view.component'
import { SearchOptionsComponent } from './search-options/search-options.component'

@Component({
  selector: 'app-location-search',
  imports: [CommonModule, SafeHtmlPipe, FeatureViewComponent, SearchOptionsComponent],
  templateUrl: './location-search.component.html',
  styleUrl: './location-search.component.css',
})
export class LocationSearchComponent {
  selectedResultUrl: string | undefined = undefined

  @Input() url: string | undefined = undefined
  @Input() label: string = 'Search location'
  @Input() title: string = 'Enter the location you want to search for'
  @Input() placeholder: string = 'Enter location to search'
  @Input() backgroundmap: BackgroundMap = 'OSM'

  searchLocation: string = ''

  projection: ProjectionMapping = defaultMapping
  $results: Observable<Feature<Geometry>[]> | undefined = undefined

  constructor(
    private logger: NGXLogger,
    private collectionService: CollectionsService,
    private locationSearchService: LocationSearchService
  ) {}

  updateSearchField(event: KeyboardEvent) {
    const inputValue = (event.target as HTMLInputElement).value
    this.logger.log(inputValue)
    if (this.url) {
      //Todo replace by generated aanroep openapi call and use inputvalue as search param:
      this.$results = this.locationSearchService.getResults({ url: this.url + '/search', dataMapping: this.projection })
    }
  }

  selectResult(item: Feature<Geometry>) {
    this.logger.log(item)
    this.selectedResultUrl = item.getProperties()['href'] as string
  }
}
