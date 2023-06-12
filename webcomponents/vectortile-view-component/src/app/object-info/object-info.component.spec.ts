import { ComponentFixture, TestBed } from '@angular/core/testing';

import { ObjectInfoComponent } from './object-info.component';

describe('ObjectInfoComponent', () => {
  let component: ObjectInfoComponent;
  let fixture: ComponentFixture<ObjectInfoComponent>;

  beforeEach(() => {
    TestBed.configureTestingModule({
      imports: [ObjectInfoComponent]
    });
    fixture = TestBed.createComponent(ObjectInfoComponent);
    component = fixture.componentInstance;
    fixture.detectChanges();
  });

  it('should create', () => {
    expect(component).toBeTruthy();
  });
});
