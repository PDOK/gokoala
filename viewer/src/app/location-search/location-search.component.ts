import { Component, Input, OnInit } from '@angular/core'
import { CommonModule } from '@angular/common'

import { CollectionsService } from '../api/services'
import { CollectionLink, Collection } from '../api/models'
import { NGXLogger } from 'ngx-logger'
import { LocationSearchService } from '../location-search.service'
import { defaultMapping, ProjectionMapping } from '../feature.service'
import { Feature } from 'ol'
import { Geometry } from 'ol/geom'
import { Observable } from 'rxjs'
import { SafeHtmlPipe } from '../safe-html.pipe'
import { FeatureViewComponent } from '../feature-view/feature-view.component'
type ActiveCollection = {
  name: string
  enabled: boolean
}

@Component({
  selector: 'app-location-search',
  imports: [CommonModule, SafeHtmlPipe, FeatureViewComponent],
  templateUrl: './location-search.component.html',
  styleUrl: './location-search.component.css',
})
export class LocationSearchComponent implements OnInit {
  selectedResultUrl: string | undefined = undefined

  @Input() url: string | undefined = undefined
  @Input() label: string = 'Search location'
  @Input() title: string = 'Enter the location you want to search for'
  @Input() placeholder: string = 'Enter location to search'

  searchLocation: string = ''
  collections: Array<Collection & CollectionLink> | undefined = undefined
  infomessage: string | undefined = undefined
  active: ActiveCollection[] = []
  projection: ProjectionMapping = defaultMapping
  $results: Observable<Feature<Geometry>[]> | undefined = undefined

  constructor(
    private logger: NGXLogger,
    private collectionService: CollectionsService,
    private locationSearchService: LocationSearchService
  ) {}

  ngOnInit() {
    if (this.url) {
      this.collectionService.rootUrl = this.url
      this.collectionService.getCollections$Json().subscribe(data => {
        this.collections = data.collections
        this.collections.forEach(x => {
          this.active.push({ name: x.id, enabled: true })
        })
        this.logger.log(this.collections)
      })
    } else {
      this.infomessage = 'please provide url to location url '
    }
  }

  isActiveCollection(name: string | undefined): boolean {
    if (name) {
      const item = this.active.find(item => item.name === name)
      return item ? item.enabled : false
    } else {
      return false
    }
  }
  updateSearchField(event: KeyboardEvent) {
    const inputValue = (event.target as HTMLInputElement).value
    this.logger.log(inputValue)
    if (this.url) {
      this.$results = this.locationSearchService.getResults({ url: this.url + '/search', dataMapping: this.projection })
    }
  }

  selectResult(item: Feature<Geometry>) {
    this.logger.log(item)
    this.selectedResultUrl = item.getProperties()['href'] as string
  }
}
