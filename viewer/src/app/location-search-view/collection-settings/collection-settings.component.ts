import { Component, inject, Input, OnInit } from '@angular/core'
import { AsyncPipe, NgClass } from '@angular/common'
import { Observable } from 'rxjs'
import { Collection, CollectionsService } from '../../shared/services/collections.service'
import { take } from 'rxjs/operators'

@Component({
  selector: 'app-collection-settings',
  standalone: true,
  imports: [NgClass, AsyncPipe],
  templateUrl: './collection-settings.component.html',
  styleUrl: './collection-settings.component.css',
})
export class CollectionSettingsComponent implements OnInit {
  @Input() open: boolean = false

  collections$!: Observable<Collection[]>

  private _collectionsService = inject(CollectionsService)

  getAvailableCollections() {
    this.collections$ = this._collectionsService.getCollections()
  }

  ngOnInit() {
    this.getAvailableCollections()
  }
}
