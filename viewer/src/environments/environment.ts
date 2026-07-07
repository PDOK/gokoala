import { NgxLoggerLevel } from 'ngx-logger'

export const environment = {
  brt: {
    backgroundUrl: 'https://service.pdok.nl/brt/achtergrondkaart/wmts/v2_0?',
    projections: ['EPSG:28992', 'EPSG:3035', 'EPSG:3857'],
  },
  loglevel: NgxLoggerLevel.DEBUG,
}
