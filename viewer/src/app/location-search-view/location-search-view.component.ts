import {
  ChangeDetectionStrategy,
  ChangeDetectorRef,
  Component,
  computed,
  ElementRef,
  EventEmitter,
  HostListener,
  inject,
  Input,
  OnDestroy,
  OnInit,
  Output,
  output,
  signal,
} from '@angular/core'

import { FormControl, FormGroup, ReactiveFormsModule } from '@angular/forms'
import { debounceTime, distinctUntilChanged, filter, map, Observable, of, Subject, Subscription, switchMap, takeUntil, tap } from 'rxjs'
import { AsyncPipe, JsonPipe, NgClass } from '@angular/common'
import { PropertyValuePipe } from './property-value.pipe'
import { CollectionSettingsComponent } from './collection-settings/collection-settings.component'
import { FeatureGeoJSON, FeatureService } from '../shared/services/feature.service'

interface LocationForm {
  location: FormControl<string | null>
}

@Component({
  selector: 'app-location-search-view',
  standalone: true,
  imports: [ReactiveFormsModule, AsyncPipe, PropertyValuePipe, NgClass, CollectionSettingsComponent, JsonPipe],
  templateUrl: './location-search-view.component.html',
  styleUrl: './location-search-view.component.css',
  changeDetection: ChangeDetectionStrategy.OnPush,
})
export class LocationSearchViewComponent implements OnInit, OnDestroy {
  @Input() projection?: string

  @Input() placeholder = 'Search by location'
  @Input() noResultsText = 'No results found'
  @Input() searchingText = 'Searching...'
  @Input() relevanceText = 'Relevance'
  @Input() collectionText = 'Collection'

  @Output() locationSelected = new EventEmitter<string>()

  form!: FormGroup<LocationForm>
  features$?: Observable<FeatureGeoJSON[]>

  query: string = ''
  searchParams: { [key: string]: number } = {}

  searchOpen = signal(false)
  searching = signal(false)
  collectionSettingsOpen = signal(false)

  hasSearched$!: Observable<boolean>

  private _featureService = inject(FeatureService)
  private _destroy$ = new Subject<void>()

  constructor(private host: ElementRef<HTMLElement>) {}

  ngOnInit() {
    const url = new URL(window.location.href)
    this.query = url.searchParams.get('q') || ''
    this.form = new FormGroup<LocationForm>({
      location: new FormControl(this.query),
    })

    this.initLocationListener()
  }

  initLocationListener() {
    this.features$ = this.form.controls.location.valueChanges.pipe(
      distinctUntilChanged(),
      filter(value => value !== null && value.length >= 3),
      tap(() => {
        this.searching.set(true)
      }),
      debounceTime(200),
      tap(val => (this.query = val || '')),
      switchMap(val => this._featureService.queryFeatures(val || '', this.searchParams, this.projection)),
      tap(() => {
        this.storeQuery()
      }),
      tap(() => {
        this.searching.set(false)
      }),
      takeUntil(this._destroy$)
    )

    this.hasSearched$ = this.form.controls.location.valueChanges.pipe(map(value => value !== null && value.length > 3))
  }

  selectFeature(feature: FeatureGeoJSON) {
    const propertyValuePipe = new PropertyValuePipe()
    this.locationSelected.emit(propertyValuePipe.transform(feature.properties, 'href'))
    this.searchOpen.set(false)
  }

  openSearchIfNot() {
    if (!this.searchOpen()) this.searchOpen.set(true)
  }

  closeSearch() {
    this.searchOpen.set(false)
  }

  toggleCollectionSettings() {
    this.collectionSettingsOpen.update(val => !val)
  }

  private storeQuery() {
    const url = new URL(window.location.href)
    url.searchParams.set('q', this.query)
    history.pushState({}, '', url.toString())
  }

  @HostListener('document:mousedown', ['$event'])
  onGlobalMouseDown(ev: MouseEvent) {
    const root = this.host.nativeElement
    const target = ev.target as Node | null
    if (target && !root.contains(target)) {
      this.closeSearch()
    }
  }

  // Close on Escape
  @HostListener('document:keydown.escape')
  onEscape() {
    this.closeSearch()
  }

  ngOnDestroy() {
    this._destroy$.next()
  }
}
