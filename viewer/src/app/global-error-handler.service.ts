import { HttpErrorResponse } from '@angular/common/http'
import { ErrorHandler, Injectable } from '@angular/core'
import { NGXLogger } from 'ngx-logger'
import { Subject } from 'rxjs'

export type ErrorDetail = {
  title: string
  detail: string
  error: unknown
  type: 'httpError' | 'unknownError'
}

@Injectable({
  providedIn: 'root',
})
export class GlobalErrorHandlerService implements ErrorHandler {
  initialErrorDetail: ErrorDetail = {
    title: 'unknown error',
    detail: 'unknown error',
    error: undefined,
    type: 'unknownError',
  }

  private _errorDetailSource = new Subject<ErrorDetail>()
  public errorDetailStream$ = this._errorDetailSource.asObservable()

  constructor(private logger: NGXLogger) {
    this._errorDetailSource.next(this.initialErrorDetail)
  }

  handleError(error: unknown): void {
    const errorDetail: ErrorDetail = JSON.parse(JSON.stringify(this.initialErrorDetail))

    if (error instanceof HttpErrorResponse) {
      this.logger.log('http error detected')
      const errorResponse = error as HttpErrorResponse
      this.logger.log(errorResponse.error)

      errorDetail.detail = errorResponse.error?.detail ?? 'No detail available'
      errorDetail.title = errorResponse.error?.title ?? 'No title available'
      errorDetail.type = 'httpError'
    } else {
      this.logger.error('error detected')
      this.logger.error(error)
      errorDetail.type = 'unknownError'
    }

    this._errorDetailSource.next(errorDetail)
  }
}
