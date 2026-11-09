import { getTopLeft } from 'ol/extent'
import TileLayer from 'ol/layer/Tile'
import { Projection } from 'ol/proj'
import { OSM, WMTS } from 'ol/source'
import WMTSTileGrid from 'ol/tilegrid/WMTS'
import { environment } from 'src/environments/environment'

export type BackgroundMap = 'BRT' | 'OSM'

export function osmLayer() {
  return new TileLayer({ source: new OSM() })
}

export function brtLayer(p: { projection: Projection; resolutions: number[]; matrixIds: string[] }) {
  return new TileLayer({
    source: new WMTS({
      attributions: 'Kaartgegevens: &copy; <a href="https://www.kadaster.nl">Kadaster</a>',
      url: environment.brt.backgroundUrl,
      layer: 'grijs',
      matrixSet: p.projection.getCode(),
      format: 'image/png',
      projection: p.projection,
      tileGrid: new WMTSTileGrid({
        origin: getTopLeft(p.projection.getExtent()),
        resolutions: p.resolutions,
        matrixIds: p.matrixIds,
      }),
      style: 'default',
      wrapX: false,
    }),
  })
}
