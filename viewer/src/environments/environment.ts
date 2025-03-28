
import { NgxLoggerLevel } from 'ngx-logger'
import { initialCurrentHttp } from 'src/app/app.module'

export const environment = {
  bgtBackgroundUrl: 'https://service.pdok.nl/brt/achtergrondkaart/wmts/v2_0?',
  loglevel: NgxLoggerLevel.OFF,
  // Default configuration for HTTP requests
  currentHttp: initialCurrentHttp
}
