import { NgxLoggerLevel } from 'ngx-logger'

export const environment = {
  bgt: {
    backgroundUrl: 'https://service.pdok.nl/brt/achtergrondkaart/wmts/v2_0?',
    projections: ['EPSG:28992', 'EPSG:25831', 'EPSG:3857'],
  },
  loglevel: NgxLoggerLevel.OFF,
}
