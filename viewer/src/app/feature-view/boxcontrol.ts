import { Control } from 'ol/control.js'
import { Draw } from 'ol/interaction'
import { createBox } from 'ol/interaction/Draw'
import VectorSource from 'ol/source/Vector'

import { EventEmitter } from '@angular/core'
import { Fill, Stroke, Style } from 'ol/style'
import VectorLayer from 'ol/layer/Vector'

export class boxControl extends Control {
  /**
   * @param {Object} [opt_options] Control options.
   */
  // eslint-disable-next-line @typescript-eslint/no-explicit-any
  constructor(
    public boxEmitter: EventEmitter<string>,

    // eslint-disable-next-line @typescript-eslint/no-explicit-any
    opt_options: any
    // eslint-disable-next-line @typescript-eslint/no-explicit-any
  ) {
    const options = opt_options || {}

    const button = document.createElement('button')

    const element = document.createElement('div')
    element.className = 'boundingboxcontrol ol-unselectable ol-control'
    button.title = 'Draw boundingbox'
    element.appendChild(button)

    super({
      element: element,
      target: options.target,
    })

    button.innerHTML = `
      <svg xmlns="http://www.w3.org/2000/svg" class = "innersvg" viewBox="0 0 448 512"><!--! Font Awesome Free 6.2.0 by @fontawesome - https://fontawesome.com License - https://fontawesome.com/license/free (Icons: CC BY 4.0, Fonts: SIL OFL 1.1, Code: MIT License) Copyright 2022 Fonticons, Inc. --><path d="M368 80h32v32H368V80zM352 32c-17.7 0-32 14.3-32 32H128c0-17.7-14.3-32-32-32H32C14.3 32 0 46.3 0 64v64c0 17.7 14.3 32 32 32V352c-17.7 0-32 14.3-32 32v64c0 17.7 14.3 32 32 32H96c17.7 0 32-14.3 32-32H320c0 17.7 14.3 32 32 32h64c17.7 0 32-14.3 32-32V384c0-17.7-14.3-32-32-32V160c17.7 0 32-14.3 32-32V64c0-17.7-14.3-32-32-32H352zM96 160c17.7 0 32-14.3 32-32H320c0 17.7 14.3 32 32 32V352c-17.7 0-32 14.3-32 32H128c0-17.7-14.3-32-32-32V160zM48 400H80v32H48V400zm320 32V400h32v32H368zM48 112V80H80v32H48z"/></svg>
      `

    button.addEventListener('click', this.addBox.bind(this, this), false)
  }
  addBox(control: boxControl) {
    const draw = new Draw({
      source: new VectorSource({}),
      type: 'Circle',
      geometryFunction: createBox(),
    })
    draw.on('drawend', e => {
      const map = this.getMap()!
      const bbox = e.feature //this is the feature fired the event
      console.log(JSON.stringify(bbox))
      const bboxGeometry = bbox.getGeometry()
      if (bboxGeometry) {
        if (map.getView().getProjection().getCode() === 'EPSG:3857') {
          const box84 = bboxGeometry.transform(map.getView().getProjection(), 'EPSG:4326').getExtent()
          const extString = box84.join(',')
          this.boxEmitter.emit(extString)
        } else {
          const box = bboxGeometry.getExtent()
          const extString = box.join(',')
          this.boxEmitter.emit(extString)
        }

        const bboxStyle = new Style({
          stroke: new Stroke({
            color: 'blue',
            width: 3,
          }),
          fill: new Fill({
            color: 'rgba(0, 0, 255, 0.06)',
          }),
        })
        const bboxsource = new VectorSource({})
        bboxsource.addFeature(bbox)
        const boxLayer = new VectorLayer({ source: bboxsource, style: bboxStyle })
        map.addLayer(boxLayer)
        map.removeInteraction(draw)
      }
    })

    this.getMap()!.addInteraction(draw)
  }
}
