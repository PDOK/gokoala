import { BrowserModule } from '@angular/platform-browser'
import { VectortileViewComponent } from './vectortile-view/vectortile-view.component'
import { createCustomElement } from '@angular/elements'
import { ObjectInfoComponent } from './object-info/object-info.component'
import { NgModule, Injector } from '@angular/core'
import { HttpEvent, HttpEventType, HttpHandlerFn, HttpRequest, provideHttpClient, withInterceptors } from '@angular/common/http'
import { LegendViewComponent } from './legend-view/legend-view.component'
import { FeatureViewComponent } from './feature-view/feature-view.component'
import { LocationSearchComponent } from './location-search/location-search.component'

import { LoggerModule, NgxLoggerLevel } from 'ngx-logger'

import { Observable, tap } from 'rxjs'
import { environment } from 'src/environments/environment'

export class Global {
  static loggingInterceptor(req: HttpRequest<unknown>, next: HttpHandlerFn): Observable<HttpEvent<unknown>> {
    return next(req).pipe(
      tap(event => {
        if (event.type === HttpEventType.Response) {
          environment.currenturl = req.urlWithParams
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
      level: environment.loglevel,
      serverLogLevel: NgxLoggerLevel.OFF,
    }),
  ],

  providers: [provideHttpClient(withInterceptors([Global.loggingInterceptor]))],
})
export class AppModule {
  constructor(private injector: Injector) {
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
  }

  // eslint-disable-next-line
  ngDoBootstrap() {
    // noop
  }
}
