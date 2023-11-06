import { ComponentFixture, TestBed } from '@angular/core/testing'

import { LegendViewComponent } from './legend-view.component'

describe('LegendViewComponent', () => {
  let component: LegendViewComponent
  let fixture: ComponentFixture<LegendViewComponent>

  beforeEach(() => {
    TestBed.configureTestingModule({
      declarations: [LegendViewComponent],
    })
    fixture = TestBed.createComponent(LegendViewComponent)
    component = fixture.componentInstance
    fixture.detectChanges()
  })

  it('should create', () => {
    expect(component).toBeTruthy()
  })
})
