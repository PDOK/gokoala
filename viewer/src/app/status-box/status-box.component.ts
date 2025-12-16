import { AfterViewInit, Component, ErrorHandler, inject } from '@angular/core'
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
  private logger = inject(NGXLogger)
  private globalErrorHandlerService = inject(GlobalErrorHandlerService)

  errorDetail: ErrorDetail | undefined = undefined

  ngAfterViewInit(): void {
    this.globalErrorHandlerService.errorDetailStream$.subscribe(errorDetail => {
      this.logger.log('errorDetail received in status-box')
      this.errorDetail = errorDetail
    })
  }
}
