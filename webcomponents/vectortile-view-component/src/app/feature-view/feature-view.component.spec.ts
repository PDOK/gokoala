import { ComponentFixture, TestBed } from '@angular/core/testing';

import { FeatureViewComponent } from './feature-view.component';

describe('FeatureViewComponent', () => {
  let component: FeatureViewComponent;
  let fixture: ComponentFixture<FeatureViewComponent>;

  beforeEach(() => {
    TestBed.configureTestingModule({
      declarations: [FeatureViewComponent]
    });
    fixture = TestBed.createComponent(FeatureViewComponent);
    component = fixture.componentInstance;
    fixture.detectChanges();
  });

  it('should create', () => {
    expect(component).toBeTruthy();
  });
});
