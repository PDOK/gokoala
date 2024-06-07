import { BrowserModule } from '@angular/platform-browser'
import { VectortileViewComponent } from './vectortile-view/vectortile-view.component'
import { createCustomElement } from '@angular/elements'
import { ObjectInfoComponent } from './object-info/object-info.component'
import { NgModule, Injector } from '@angular/core'
import { provideHttpClient, withInterceptorsFromDi } from '@angular/common/http'
import { LegendViewComponent } from './legend-view/legend-view.component'
import { FeatureViewComponent } from './feature-view/feature-view.component'
import { LoggerModule, NgxLoggerLevel } from 'ngx-logger'
import { environment } from 'src/environments/environment'

@NgModule({
  declarations: [],
  bootstrap: [],
  imports: [
    BrowserModule,
    VectortileViewComponent,
    LoggerModule.forRoot({
      serverLoggingUrl: '/api/logs',
      level: environment.loglevel,
      serverLogLevel: NgxLoggerLevel.OFF,
    }),
  ],
  providers: [provideHttpClient(withInterceptorsFromDi())],
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
  }

  // eslint-disable-next-line
  ngDoBootstrap() {
    // noop
  }
}
