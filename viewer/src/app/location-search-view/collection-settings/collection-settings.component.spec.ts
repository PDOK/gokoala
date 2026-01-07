import { ComponentFixture, TestBed } from '@angular/core/testing'

import { CollectionSettingsComponent } from './collection-settings.component'

describe('CollectionSettingsComponent', () => {
  let component: CollectionSettingsComponent
  let fixture: ComponentFixture<CollectionSettingsComponent>

  beforeEach(async () => {
    await TestBed.configureTestingModule({
      imports: [CollectionSettingsComponent],
    }).compileComponents()

    fixture = TestBed.createComponent(CollectionSettingsComponent)
    component = fixture.componentInstance
    fixture.detectChanges()
  })

  it('should create', () => {
    expect(component).toBeTruthy()
  })
})
