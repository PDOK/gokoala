import { Component, Input } from '@angular/core'

import { CommonModule } from '@angular/common'

import { CollectionsService, FeaturesService } from '../api/services'
import { NGXLogger } from 'ngx-logger'

import { defaultMapping, ProjectionMapping } from '../feature.service'

import { SafeHtmlPipe } from '../safe-html.pipe'
import { BackgroundMap, FeatureViewComponent } from '../feature-view/feature-view.component'
import { SearchOptionsComponent } from './search-options/search-options.component'

import { FeatureJsonfg } from '../api/models'
import { Search$Json$Params } from '../api/fn/features/search-json'

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
  defaultColparams = { relevance: 0.5, version: 1 }
  @Input() searchParams: Search$Json$Params = {
    q: '',
    functioneel_gebied: this.defaultColparams,
    geografisch_gebied: this.defaultColparams,
    ligplaats: this.defaultColparams,
    standplaats: this.defaultColparams,
    verblijfsobject: this.defaultColparams,
    woonplaats: this.defaultColparams,
  }

  searchLocation: string = ''

  projection: ProjectionMapping = defaultMapping
  results: FeatureJsonfg[] = []

  constructor(
    private logger: NGXLogger,
    private collectionService: CollectionsService,
    private featuresService: FeaturesService
  ) {}

  updateSearchField(event: KeyboardEvent) {
    const inputValue = (event.target as HTMLInputElement).value
    this.searchParams.q = inputValue
    this.logger.log(inputValue)
    this.lookup()
  }

  private lookup() {
    if (this.url) {
      this.featuresService.rootUrl = this.url
      this.results = []
      this.featuresService.search$Json(this.searchParams).subscribe(x => {
        this.results = x.features
      })
    }
  }

  selectResult(item: FeatureJsonfg) {
    this.logger.log('lookup via link to api: ')
    this.logger.log(item)
    if (item.links![0].href) {
      // this.selectedResultUrl = item.links![0].href as string
      //e.g: this.selectedResultUrl =
      //  'https://api.pdok.nl/lv/bag/ogc/v1-demo/collections/verblijfsobject/items/80f96ef7-dfa4-5197-b681-cfd92b10757e'
    }
  }
  getHighLight(r: { properties: unknown }): string {
    return this.getProperty(r, 'highlight')
  }

  getDisplayname(r: { properties: unknown }): string {
    return this.getProperty(r, 'displayName')
  }

  getProperty(r: { properties: unknown }, propertyName: string): string {
    const p = r.properties as { [key: string]: unknown }

    if (p[propertyName]) {
      return p[propertyName] as string
    } else {
      return ''
    }
  }

  paramChanged(event: Search$Json$Params) {
    this.logger.log('paramchanged:')
    this.logger.log(JSON.stringify(event))
    this.searchParams = event
    this.lookup()
  }
}
