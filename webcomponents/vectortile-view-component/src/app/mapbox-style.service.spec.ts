import { TestBed } from '@angular/core/testing';

import { MapboxStyleService } from './mapbox-style.service';

describe('MapboxStyleService', () => {
  let service: MapboxStyleService;

  beforeEach(() => {
    TestBed.configureTestingModule({});
    service = TestBed.inject(MapboxStyleService);
  });

  it('should be created', () => {
    expect(service).toBeTruthy();
  });
});
