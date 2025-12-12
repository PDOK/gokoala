import { Component, Input, Output, OnInit, EventEmitter } from '@angular/core'

import { CollectionsService } from '../../api/services'
import { CollectionLink, Collection } from '../../api/models'
import { NGXLogger } from 'ngx-logger'
import { Search$Json$Params } from 'src/app/api/fn/features/search-json'

type ActiveCollection = {
  name: string
  enabled: boolean
  relevance: number
  version: number
}

@Component({
  selector: 'app-search-options',
  imports: [],
  templateUrl: './search-options.component.html',
  styleUrl: './search-options.component.css',
})
export class SearchOptionsComponent implements OnInit {
  @Input() url: string | undefined = undefined
  defaultColparams = { relevance: 0.5, version: 1 }
  @Input() searchParams: Search$Json$Params = {
    q: '',
  }
  @Output() searchParamsEvent = new EventEmitter<Search$Json$Params>()
  collections: Array<Collection & CollectionLink> | undefined = undefined
  infomessage: string | undefined = undefined
  active: ActiveCollection[] = []
  visible: boolean = false

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
          this.active.push({ name: x.id, enabled: true, relevance: 0.5, version: 1 })
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

  getRelevance(name: string | undefined): number {
    if (name) {
      const item = this.active.find(item => item.name === name)
      return item ? item.relevance : 0.5
    } else {
      return 0.5
    }
  }

  SetActiveCollection(name: string | undefined, enabled: boolean) {
    if (name) {
      const item = this.active.find(item => item.name === name)
      if (item) {
        item.enabled = enabled
      } else {
        this.logger.error(name + ' collection not found')
      }
    }
  }

  SetRelevance(name: string | undefined, relevance: number) {
    if (name) {
      const item = this.active.find(item => item.name === name)
      if (item) {
        item.relevance = relevance
      } else {
        this.logger.error(name + ' collection not found')
      }
    }
  }

  toggleVisible() {
    this.visible = !this.visible
  }

  checboxChanged(title: string | undefined, event: Event) {
    if (title) {
      this.SetActiveCollection(title, (<HTMLInputElement>event.target).checked)
      this.emit()
    }
  }

  relevanceChanged(title: string | undefined, event: Event) {
    if (title) {
      this.SetRelevance(title, Number((<HTMLInputElement>event.target).value))
      this.emit()
    }
  }

  emit() {
    const newSearchParams = { q: this.searchParams.q } as Search$Json$Params
    this.active.forEach(x => {
      if (x.name === 'functioneel_gebied' && x.enabled) {
        newSearchParams.functioneel_gebied = { relevance: x.relevance, version: x.version }
      }
      if (x.name === 'geografisch_gebied' && x.enabled) {
        newSearchParams.geografisch_gebied = { relevance: x.relevance, version: x.version }
      }
      if (x.name === 'ligplaats' && x.enabled) {
        newSearchParams.ligplaats = { relevance: x.relevance, version: x.version }
      }
      if (x.name === 'standplaats' && x.enabled) {
        newSearchParams.standplaats = { relevance: x.relevance, version: x.version }
      }

      if (x.name === 'verblijfsobject' && x.enabled) {
        newSearchParams.verblijfsobject = { relevance: x.relevance, version: x.version }
      }

      if (x.name === 'woonplaats' && x.enabled) {
        newSearchParams.woonplaats = { relevance: x.relevance, version: x.version }
      }

      if (x.name === 'gemeentegebied' && x.enabled) {
        newSearchParams.gemeentegebied = { relevance: x.relevance, version: x.version }
      }
      if (x.name === 'provinciegebied' && x.enabled) {
        newSearchParams.provinciegebied = { relevance: x.relevance, version: x.version }
      }
      if (x.name === 'perceel' && x.enabled) {
        newSearchParams.perceel = { relevance: x.relevance, version: x.version }
      }
    })

    this.logger.log(JSON.stringify(newSearchParams))

    this.searchParamsEvent.emit(newSearchParams)
  }
}
