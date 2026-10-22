import { AfterViewInit, ChangeDetectionStrategy, Component, ElementRef, Input, OnChanges, OnDestroy, inject } from '@angular/core'
import { NGXLogger } from 'ngx-logger'
import { Map as OLMap, View } from 'ol'
import { defaults as defaultControls } from 'ol/control'
import { Extent } from 'ol/extent'
import ImageLayer from 'ol/layer/Image'
import { addEquivalentProjections, get as getProjection, Projection } from 'ol/proj'
import OGCMapSource from 'ol/source/OGCMap'
import { getRijksdriehoek, initProj4WithTilesDefaults } from '../shared/model/map-projection'
import { NgChanges } from '../vectortile-view/vectortile-view.component'

export type MapFormat = 'png' | 'jpeg'

const EPSG_PREFIX = 'EPSG:'
const EPSG_URI_PREFIX = 'http://www.opengis.net/def/crs/EPSG/0/'
const RD_NEW_SRID = '28992'

@Component({
  selector: 'app-maps-view',
  templateUrl: './maps-view.component.html',
  styleUrls: ['./maps-view.component.css'],
  changeDetection: ChangeDetectionStrategy.OnPush,
  standalone: true,
})
export class MapsViewComponent implements OnChanges, AfterViewInit, OnDestroy {
  private readonly el: ElementRef = inject(ElementRef)
  private readonly logger = inject(NGXLogger)

  @Input() mapUrl!: string
  @Input() crs: string = EPSG_PREFIX + RD_NEW_SRID
  @Input() bbox: string | undefined
  @Input() format: MapFormat = 'png'

  map = new OLMap({
    controls: defaultControls(),
  })
  private _viewInitialized = false

  ngOnChanges(changes: NgChanges<MapsViewComponent>) {
    if (changes.mapUrl || changes.crs || changes.bbox || changes.format) {
      this.initialize()
    }
  }

  ngAfterViewInit() {
    this.map.setTarget(this.el.nativeElement.querySelector('.maps-area'))
    // prevent initialization if view isn't ready yet, depends on map#setTarget as defined above
    this._viewInitialized = true
    this.initialize()
  }

  ngOnDestroy() {
    this.map.setTarget(undefined)
  }

  private initialize() {
    if (!this._viewInitialized) {
      return
    }
    if (!this.mapUrl) {
      this.logger.error('No map-url was provided')
      return
    }
    const projection = this.getMapProjection(this.crs)
    if (!projection) {
      this.logger.error(`CRS ${this.crs} is not supported`)
      return
    }

    this.map.getLayers().clear()
    this.map.addLayer(
      new ImageLayer({
        source: new OGCMapSource({
          url: this.mapUrl,
          projection: projection,
          params: { f: this.format },
          hidpi: false,
          ratio: 1,
        }),
      })
    )

    const view = new View({
      projection: projection,
      enableRotation: false,
    })
    this.map.setView(view)
    const extent = this.parseBbox(this.bbox) ?? projection.getExtent()
    if (extent) {
      view.fit(extent, { size: this.map.getSize() })
    }
  }

  private getMapProjection(crs: string): Projection | undefined {
    const srid = crs.startsWith(EPSG_URI_PREFIX) ? crs.substring(EPSG_URI_PREFIX.length) : crs.replace(EPSG_PREFIX, '')
    const uri = EPSG_URI_PREFIX + srid
    // register shared projection definitions (RD, ETRS89), normally only done by MapProjection
    initProj4WithTilesDefaults(this.logger)
    const existing = getProjection(uri)
    if (existing) {
      return existing
    }

    const epsgProjection = srid === RD_NEW_SRID ? getRijksdriehoek().projection : getProjection(EPSG_PREFIX + srid)
    if (!epsgProjection) {
      return undefined
    }
    // ol needs degrees, proj gives 'degree'
    const units = (epsgProjection.getUnits() as string) === 'degree' ? 'degrees' : epsgProjection.getUnits()
    const uriProjection = new Projection({
      code: uri,
      units: units,
      extent: epsgProjection.getExtent(),
      axisOrientation: epsgProjection.getAxisOrientation(),
      global: epsgProjection.isGlobal(),
    })
    addEquivalentProjections([epsgProjection, uriProjection])
    return uriProjection
  }

  private parseBbox(bbox: string | undefined): Extent | undefined {
    if (!bbox) {
      return undefined
    }
    const extent = bbox.split(',').map(Number)
    if (extent.length !== 4 || extent.some(isNaN)) {
      this.logger.error(`Invalid bbox ${bbox} provided, expected minx,miny,maxx,maxy`)
      return undefined
    }
    return extent
  }
}
