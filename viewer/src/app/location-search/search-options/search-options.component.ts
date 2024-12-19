import { Component, Input, OnInit } from '@angular/core'
import { CommonModule } from '@angular/common'

import { CollectionsService } from '../../api/services'
import { CollectionLink, Collection } from '../../api/models'
import { NGXLogger } from 'ngx-logger'
import { LocationSearchService } from '../../location-search.service'

type ActiveCollection = {
  name: string
  enabled: boolean
}

@Component({
  selector: 'app-search-options',
  imports: [CommonModule],
  templateUrl: './search-options.component.html',
  styleUrl: './search-options.component.css',
})
export class SearchOptionsComponent implements OnInit {
  @Input() url: string | undefined = undefined

  collections: Array<Collection & CollectionLink> | undefined = undefined
  infomessage: string | undefined = undefined
  active: ActiveCollection[] = []
  visible: boolean = false

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

  toggleVisible() {
    this.visible = !this.visible
  }
}
