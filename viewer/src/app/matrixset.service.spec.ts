import { TestBed } from '@angular/core/testing'

import { MatrixsetService } from './matrixset.service'

describe('MatrixsetService', () => {
  let service: MatrixsetService

  beforeEach(() => {
    TestBed.configureTestingModule({})
    service = TestBed.inject(MatrixsetService)
  })

  it('should be created', () => {
    expect(service).toBeTruthy()
  })
})
