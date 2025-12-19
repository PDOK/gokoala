import { Component, EventEmitter, inject, Input, OnDestroy, OnInit, Output } from '@angular/core'
import { AsyncPipe, NgClass } from '@angular/common'
import { map, Observable, startWith, Subject, takeUntil, tap, withLatestFrom } from 'rxjs'
import { Collection, CollectionsService } from '../../shared/services/collections.service'
import { FormArray, FormControl, FormGroup, ReactiveFormsModule } from '@angular/forms'

interface CollectionSetting {
  checked: FormControl<boolean>
  relevance: FormControl<number>
}

@Component({
  selector: 'app-collection-settings',
  standalone: true,
  imports: [NgClass, AsyncPipe, ReactiveFormsModule],
  templateUrl: './collection-settings.component.html',
  styleUrl: './collection-settings.component.css',
})
export class CollectionSettingsComponent implements OnInit, OnDestroy {
  @Input() open: boolean = false
  @Input() collectionText = 'Collection'
  @Input() relevanceText = 'Relevance'

  @Output() formChange = new EventEmitter<{ [key: string]: number }>()

  form!: FormArray<FormGroup<CollectionSetting>>
  collections$!: Observable<Collection[]>

  private _destroy$ = new Subject<void>()

  private _collectionsService = inject(CollectionsService)

  ngOnInit() {
    this.form = new FormArray<FormGroup<CollectionSetting>>([])
    this.collections$ = this._collectionsService.getCollections().pipe(takeUntil(this._destroy$))
    this.emitFormChanges()
    this.buildForm()
  }

  buildForm() {
    const url = new URL(window.location.href)
    let hasAnyParam = false
    this.collections$.subscribe(collections => {
      collections.forEach(collection => {
        const param = url.searchParams.get(`${collection.title}[relevance]`)
        let relevance = 0.5
        if (param) {
          hasAnyParam = true
          relevance = parseFloat(param)
        }
        this.form.push(
          new FormGroup<CollectionSetting>({
            checked: new FormControl<boolean>(!!param, { nonNullable: true }),
            relevance: new FormControl<number>(relevance, { nonNullable: true }),
          })
        )
      })

      if (!hasAnyParam) this.form.controls.forEach(control => control.patchValue({ checked: true }))
    })
  }

  emitFormChanges() {
    this.form.valueChanges
      .pipe(
        startWith(this.form.getRawValue()),
        withLatestFrom(this.collections$),
        map(([formValues, collections]) => {
          const searchParams: { [key: string]: number } = {}
          formValues.forEach((formValue, idx) => {
            if (formValue.checked && formValue.relevance) searchParams[collections[idx].title] = formValue.relevance
          })
          return searchParams
        }),
        tap(formValue => {
          this.formChange.emit(formValue)
        }),
        takeUntil(this._destroy$)
      )
      .subscribe(formValue => this.storeSettings(formValue))
  }

  private storeSettings(formValue: { [key: string]: number }) {
    const url = new URL(window.location.href)
    url.search = ''
    for (const key in formValue) {
      url.searchParams.append(`${key}[relevance]`, formValue[key].toString())
      url.searchParams.append(`${key}[version]`, '1')
    }
    history.replaceState({}, '', url.toString())
  }

  ngOnDestroy() {
    this._destroy$.next()
  }
}
