import { Component, Input, OnInit } from '@angular/core'
import { CommonModule } from '@angular/common'

import { CollectionsService } from '../api/services'
import { CollectionLink, Collection } from '../api/models'
import { NGXLogger } from 'ngx-logger'
type ActiveCollection = {
  name: string
  enabled: boolean
}

@Component({
  selector: 'app-location-search',
  imports: [CommonModule],
  templateUrl: './location-search.component.html',
  styleUrl: './location-search.component.css',
})
export class LocationSearchComponent implements OnInit {
  @Input() url: string | undefined = undefined
  @Input() label: string = 'Search location'
  @Input() title: string = 'Enter the location you want to search for'
  @Input() placeholder: string = 'Enter location to search'

  searchLocation: string = ''
  collections: Array<Collection & CollectionLink> | undefined = undefined
  infomessage: string | undefined = undefined
  active: ActiveCollection[] = []

  updateSearchField(event: KeyboardEvent) {
    this.logger.log(event)
  }

  constructor(
    private logger: NGXLogger,
    private collectionService: CollectionsService
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
}
