import { Component, ElementRef, EventEmitter, HostListener, inject, Input, OnDestroy, OnInit, Output, signal } from '@angular/core'

import { FormControl, FormGroup, ReactiveFormsModule } from '@angular/forms'
import { debounceTime, distinctUntilChanged, map, Observable, Subscription, switchMap, tap } from 'rxjs'
import { FeaturesService } from '../api/services'
import { filter } from 'rxjs/operators'
import { AsyncPipe, NgClass } from '@angular/common'
import { FeatureJsonfg } from '../api/models/feature-jsonfg'
import { PropertyValuePipe } from './property-value.pipe'

interface LocationForm {
  location: FormControl<string | null>
}

@Component({
  selector: 'app-location-search-view',
  standalone: true,
  imports: [ReactiveFormsModule, AsyncPipe, PropertyValuePipe, NgClass],
  templateUrl: './location-search-view.component.html',
  styleUrl: './location-search-view.component.css',
})
export class LocationSearchViewComponent implements OnInit, OnDestroy {
  @Input() placeholder = 'Search by location'

  @Output() locationSelected = new EventEmitter<string>()

  form!: FormGroup<LocationForm>
  features$?: Observable<FeatureJsonfg[]>
  defaultColparams = { relevance: 0.5, version: 1 }

  open = signal(false)
  searching = signal(false)
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
      filter(val => val !== null && val.length > 3),
      debounceTime(50),
      distinctUntilChanged(),
      tap(val => (this.searchParams.q = val ? val : '')),
      tap(() => this.searching.set(true)),
      switchMap(() => this._featureService.search$Json(this.searchParams)),
      tap(() => this.searching.set(false)),
      map(val => val.features)
    )
  }

  ngOnDestroy() {
    this._locationChangesSub$.unsubscribe()
  }

  selectFeature(feature: FeatureJsonfg) {
    const propertyValuePipe = new PropertyValuePipe()
    this.locationSelected.emit(propertyValuePipe.transform(feature.properties, 'href'))
    this.open.set(false)
  }

  toggle() {
    this.open.update(open => !open)
  }

  openIfNot() {
    if (!this.open()) this.open.set(true)
  }

  close() {
    this.open.set(false)
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
