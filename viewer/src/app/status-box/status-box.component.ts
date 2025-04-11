import { AfterViewInit, Component, ErrorHandler } from '@angular/core'
import { ErrorDetail, GlobalErrorHandlerService } from '../global-error-handler.service'
import { NGXLogger } from 'ngx-logger'

@Component({
  selector: 'app-status-box',
  imports: [],
  templateUrl: './status-box.component.html',
  styleUrl: './status-box.component.css',
  providers: [{ provide: ErrorHandler, useClass: GlobalErrorHandlerService }],
})
export class StatusBoxComponent implements AfterViewInit {
  errorDetail: ErrorDetail | undefined = undefined

  constructor(
    private logger: NGXLogger,
    private globalErrorHandlerService: GlobalErrorHandlerService
  ) {}

  ngAfterViewInit(): void {
    this.globalErrorHandlerService.errorDetailStream$.subscribe(errorDetail => {
      this.logger.log('errorDetail received in status-box')
      this.errorDetail = errorDetail
    })
  }
}
