import {
  ChangeDetectionStrategy,
  Component,
  ElementRef,
  EventEmitter,
  HostListener,
  inject,
  Input,
  OnChanges,
  OnDestroy,
  OnInit,
  Output,
  signal,
  SimpleChanges,
} from '@angular/core'
import { FormControl, FormGroup, ReactiveFormsModule } from '@angular/forms'
import { debounceTime, distinctUntilChanged, filter, map, Observable, Subject, switchMap, takeUntil, tap } from 'rxjs'
import { AsyncPipe, NgClass, NgIf } from '@angular/common'
import { PropertyValuePipe } from './property-value.pipe'
import { CollectionSettingsComponent } from './collection-settings/collection-settings.component'
import { FeatureGeoJSON, FeatureService } from '../shared/services/feature.service'

interface LocationForm {
  location: FormControl<string | null>
}

@Component({
  selector: 'app-location-search-view',
  standalone: true,
  imports: [ReactiveFormsModule, AsyncPipe, PropertyValuePipe, NgClass, CollectionSettingsComponent, NgIf],
  templateUrl: './location-search-view.component.html',
  styleUrl: './location-search-view.component.css',
  changeDetection: ChangeDetectionStrategy.OnPush,
})
export class LocationSearchViewComponent implements OnInit, OnDestroy, OnChanges {
  @Input() projection: string = 'http://www.opengis.net/def/crs/OGC/1.3/CRS84'

  @Input() placeholderText = 'Search by location'
  @Input() noResultsText = 'No results found'
  @Input() searchingText = 'Searching...'
  @Input() relevanceText = 'Relevance'
  @Input() collectionsText = 'Collections'
  @Input() searchHelpText = 'Search query must be at least three characters long.'
  @Input() noCollectionsSelectedText = 'A minimum of one collection must be selected.'

  @Output() locationSelected = new EventEmitter<string>()

  form!: FormGroup<LocationForm>
  features$?: Observable<FeatureGeoJSON[]>

  query: string = ''
  readonly MIN_QUERY_LENGTH = 3
  searchParams: { [key: string]: number } = {}

  searchOpen = signal(false)
  searching = signal(false)
  collectionSettingsOpen = signal(false)
  hasSearchParams = signal(true)

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

  ngOnChanges(changes: SimpleChanges) {
    if (!changes['projection'].isFirstChange() && changes['projection'].currentValue !== changes['projection'].previousValue) {
      this.form.controls.location.patchValue('', { emitEvent: true })
    }
  }

  initLocationListener() {
    this.features$ = this.form.controls.location.valueChanges.pipe(
      distinctUntilChanged(),
      filter(value => value !== null && value.length >= this.MIN_QUERY_LENGTH && this.hasSearchParams()),
      tap(() => this.searching.set(true)),
      debounceTime(200),
      tap(val => (this.query = val || '')),
      switchMap(val => this._featureService.queryFeatures(val || '', this.searchParams, this.projection)),
      tap(() => this.storeQuery()),
      tap(() => this.searching.set(false)),
      takeUntil(this._destroy$)
    )

    this.hasSearched$ = this.form.controls.location.valueChanges.pipe(map(value => value !== null && value.length >= this.MIN_QUERY_LENGTH))
  }

  onFormChange($event: { [p: string]: number }) {
    this.hasSearchParams.update(() => {
      return Object.keys($event).length > 0
    })
    this.searchParams = $event
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
