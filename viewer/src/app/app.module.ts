import {
  HttpEvent,
  HttpEventType,
  HttpHandlerFn,
  HttpHeaders,
  HttpRequest,
  provideHttpClient,
  withInterceptors,
} from '@angular/common/http'
import { ErrorHandler, Injector, NgModule, inject } from '@angular/core'
import { createCustomElement } from '@angular/elements'
import { BrowserModule } from '@angular/platform-browser'
import { FeatureViewComponent } from './feature-view/feature-view.component'
import { LegendViewComponent } from './legend-view/legend-view.component'
import { LocationSearchComponent } from './location-search/location-search.component'
import { ObjectInfoComponent } from './object-info/object-info.component'
import { VectortileViewComponent } from './vectortile-view/vectortile-view.component'

import { LoggerModule, NgxLoggerLevel } from 'ngx-logger'

import { Observable, tap } from 'rxjs'

import { GlobalErrorHandlerService } from './global-error-handler.service'
import { StatusBoxComponent } from './status-box/status-box.component'
export type CurrentHttp = {
  url: string
  headers: HttpHeaders
}
export const initialCurrentHttp: CurrentHttp = { url: '', headers: new HttpHeaders() }
export let currentHttp: CurrentHttp = initialCurrentHttp

export class GlobalHttpInterceptor {
  static loggingInterceptor(req: HttpRequest<unknown>, next: HttpHandlerFn): Observable<HttpEvent<unknown>> {
    return next(req).pipe(
      tap(event => {
        if (event.type === HttpEventType.Response) {
          currentHttp = { url: req.urlWithParams, headers: event.headers }
        }
        if (event.type === HttpEventType.ResponseHeader) {
          currentHttp = { url: req.urlWithParams, headers: event.headers }
        }
      })
    )
  }
}

@NgModule({
  declarations: [],
  bootstrap: [],
  imports: [
    BrowserModule,
    LoggerModule.forRoot({
      serverLoggingUrl: '/api/logs',
      level: NgxLoggerLevel.LOG,
      serverLogLevel: NgxLoggerLevel.OFF,
    }),
  ],

  providers: [
    { provide: ErrorHandler, useClass: GlobalErrorHandlerService },
    provideHttpClient(withInterceptors([GlobalHttpInterceptor.loggingInterceptor])),
  ],
})
export class AppModule {
  private injector = inject(Injector)

  constructor() {
    const injector = this.injector

    const vectorTileView = createCustomElement(VectortileViewComponent, { injector })
    customElements.define('app-vectortile-view', vectorTileView)

    const objectInfo = createCustomElement(ObjectInfoComponent, { injector })
    customElements.define('app-objectinfo-view', objectInfo)

    const legendView = createCustomElement(LegendViewComponent, { injector })
    customElements.define('app-legend-view', legendView)

    const featureView = createCustomElement(FeatureViewComponent, { injector })
    customElements.define('app-feature-view', featureView)

    const locationSearch = createCustomElement(LocationSearchComponent, { injector })
    customElements.define('app-location-search', locationSearch)

    const statusBox = createCustomElement(StatusBoxComponent, { injector })
    customElements.define('app-status-box', statusBox)
  }

  // eslint-disable-next-line
  ngDoBootstrap() {
    // noop
  }
}
