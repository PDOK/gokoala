import { Component, ElementRef, EventEmitter, HostListener, inject, Input, OnDestroy, OnInit, Output, Signal, signal } from '@angular/core'

import { FormControl, FormGroup, ReactiveFormsModule } from '@angular/forms'
import { debounceTime, distinctUntilChanged, map, Observable, Subscription, switchMap, tap } from 'rxjs'
import { FeaturesService } from '../api/services'
import { AsyncPipe, JsonPipe, NgClass } from '@angular/common'
import { FeatureJsonfg } from '../api/models/feature-jsonfg'
import { PropertyValuePipe } from './property-value.pipe'
import { CollectionSettingsComponent } from './collection-settings/collection-settings.component'

interface LocationForm {
  location: FormControl<string | null>
}

@Component({
  selector: 'app-location-search-view',
  standalone: true,
  imports: [ReactiveFormsModule, AsyncPipe, PropertyValuePipe, NgClass, CollectionSettingsComponent],
  templateUrl: './location-search-view.component.html',
  styleUrl: './location-search-view.component.css',
})
export class LocationSearchViewComponent implements OnInit, OnDestroy {
  @Input() placeholder = 'Search by location'

  @Output() locationSelected = new EventEmitter<string>()

  form!: FormGroup<LocationForm>
  features$?: Observable<FeatureJsonfg[]>

  defaultColparams = { relevance: 0.5, version: 1 }

  searchOpen = signal(false)
  searching = signal(false)
  hasSearched$!: Observable<boolean>

  // eslint-disable-next-line
  searchParams: any = {
    q: '',
    functioneel_gebied: this.defaultColparams,
    geografisch_gebied: this.defaultColparams,
    ligplaats: this.defaultColparams,
    standplaats: this.defaultColparams,
    verblijfsobject: this.defaultColparams,
    woonplaats: this.defaultColparams,
  }

  private _locationChangesSub$!: Subscription
  private _featureService = inject(FeaturesService)

  constructor(private host: ElementRef<HTMLElement>) {}

  ngOnInit() {
    this.form = new FormGroup<LocationForm>({
      location: new FormControl(''),
    })

    this.initLocationListener()
  }

  initLocationListener() {
    this.features$ = this.form.controls.location.valueChanges.pipe(
      distinctUntilChanged(),
      tap(() => this.searching.set(true)),
      debounceTime(200),
      tap(val => (this.searchParams.q = val ? val : '')),
      switchMap(() => this._featureService.search$Json(this.searchParams)),
      tap(() => this.searching.set(false)),
      map(val => val?.features || [])
    )
    this.hasSearched$ = this.form.controls.location.valueChanges.pipe(map(value => value !== null && value.length > 3))
  }

  ngOnDestroy() {
    this._locationChangesSub$.unsubscribe()
  }

  selectFeature(feature: FeatureJsonfg) {
    const propertyValuePipe = new PropertyValuePipe()
    this.locationSelected.emit(propertyValuePipe.transform(feature.properties, 'href'))
    this.searchOpen.set(false)
  }

  openIfNot() {
    if (!this.searchOpen()) this.searchOpen.set(true)
  }

  close() {
    this.searchOpen.set(false)
  }

  @HostListener('document:mousedown', ['$event'])
  onGlobalMouseDown(ev: MouseEvent) {
    const root = this.host.nativeElement
    const target = ev.target as Node | null
    if (target && !root.contains(target)) {
      this.close()
    }
  }

  // Close on Escape
  @HostListener('document:keydown.escape')
  onEscape() {
    this.close()
  }
}
