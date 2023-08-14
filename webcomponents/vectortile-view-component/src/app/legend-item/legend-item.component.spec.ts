import { ComponentFixture, TestBed } from '@angular/core/testing';

import { LegendItemComponent } from './legend-item.component';

describe('LegendItemComponent', () => {
  let component: LegendItemComponent;
  let fixture: ComponentFixture<LegendItemComponent>;

  beforeEach(() => {
    TestBed.configureTestingModule({
      declarations: [LegendItemComponent]
    });
    fixture = TestBed.createComponent(LegendItemComponent);
    component = fixture.componentInstance;
    fixture.detectChanges();
  });

  it('should create', () => {
    expect(component).toBeTruthy();
  });
});
