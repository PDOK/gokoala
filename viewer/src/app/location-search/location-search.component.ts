import { Component, ElementRef, EventEmitter, Input, OnInit, Output } from '@angular/core'

import { CommonModule } from '@angular/common'

import { NGXLogger } from 'ngx-logger'
import { FeaturesService } from '../api/services'

import { defaultMapping, ProjectionMapping } from '../feature.service'

import { BackgroundMap } from '../feature-view/feature-view.component'
import { SafeHtmlPipe } from '../safe-html.pipe'
import { SearchOptionsComponent } from './search-options/search-options.component'

import { FeatureLike } from 'ol/Feature'
import { GeoJSON } from 'ol/format'
import { Search$Json$Params } from '../api/fn/features/search-json'
import { FeatureCollectionJsonfg, FeatureJsonfg } from '../api/models'

import { HttpHeaders } from '@angular/common/http'

import { currentHttp, CurrentHttp } from '../app.module'
import { Observable } from 'rxjs'

@Component({
  selector: 'app-location-search',
  imports: [CommonModule, SafeHtmlPipe, SearchOptionsComponent],
  templateUrl: './location-search.component.html',
  styleUrl: './location-search.component.css',
  standalone: true,
})
export class LocationSearchComponent implements OnInit {
  selectedResultUrl: string | undefined = undefined
  @Output() activeFeatureHovered = new EventEmitter<FeatureLike>()
  @Output() activeFeatureSelected = new EventEmitter<FeatureLike>()
  @Output() activeSearchUrl = new EventEmitter<CurrentHttp>()
  @Output() activeSearchText = new EventEmitter<string>()
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
    perceel: this.defaultColparams,
    gemeentegebied: this.defaultColparams,
    provinciegebied: this.defaultColparams,
    limit: 100,
    'bbox-crs': 'http://www.opengis.net/def/crs/OGC/1.3/CRS84',
    crs: 'http://www.opengis.net/def/crs/OGC/1.3/CRS84',
  } as Required<Search$Json$Params>

  searchLocation: string = ''

  projection: ProjectionMapping = defaultMapping
  results: Observable<FeatureCollectionJsonfg> | undefined = undefined
  activeSearchUrlEmited: CurrentHttp = { url: '', headers: new HttpHeaders() }

  constructor(
    private logger: NGXLogger,
    private featuresService: FeaturesService,
    private elementRef: ElementRef
  ) {}
  ngOnInit(): void {
    this.logger.debug('LocationSearchComponent initialized with URL:', this.url)
    this.activeSearchUrl.emit(currentHttp)
  }

  updateSearchField(event: KeyboardEvent) {
    const inputValue = (event.target as HTMLInputElement).value
    this.searchParams.q = inputValue
    this.logger.debug(inputValue)
    this.activeSearchText.emit(inputValue)
    this.deSelectResult()

    this.lookup()
  }

  private emitCurrentUrl() {
    if (currentHttp.url.includes('search')) {
      if (this.activeSearchUrlEmited !== currentHttp) {
        this.activeSearchUrlEmited = currentHttp
        this.activeSearchUrl.emit(currentHttp)
      }
    }
  }

  private lookup() {
    if (this.url) {
      this.featuresService.rootUrl = this.url
      this.results = this.featuresService.search$Json(this.searchParams)
    }
  }

  selectResultHover(item: FeatureJsonfg) {
    //this.logger.log('lookup via link to api: ')
    //this.logger.log(item)
    const geoJsonFormat = new GeoJSON()

    // Read the GeoJSON data and create an OpenLayers feature
    const feature = geoJsonFormat.readFeature(item) //, { featureProjection: 'EPSG:3857'}//

    this.activeFeatureHovered.emit(feature as FeatureLike)
    //if (item.links![0].href) {
    // this.selectedResultUrl = item.links![0].href as string
    //e.g: this.selectedResultUrl =
    //  'https://api.pdok.nl/lv/bag/ogc/v1-demo/collections/verblijfsobject/items/80f96ef7-dfa4-5197-b681-cfd92b10757e'
    //}
  }

  selectResultClick(item: FeatureJsonfg) {
    const geoJsonFormat = new GeoJSON()
    const feature = geoJsonFormat.readFeature(item)
    this.activeFeatureSelected.emit(feature as FeatureLike)
  }

  deSelectResult() {
    this.activeFeatureHovered.emit(undefined)
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
    this.deSelectResult()
    this.logger.debug('paramchanged:')
    this.logger.debug(JSON.stringify(event))
    this.searchParams = event
    this.lookup()
  }

  getResults(f: FeatureCollectionJsonfg | null): FeatureJsonfg[] {
    this.emitCurrentUrl()
    if (f) {
      return f?.features
    } else {
      return []
    }
  }
}
