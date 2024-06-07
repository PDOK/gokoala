import { Control } from 'ol/control.js'

import { EventEmitter } from '@angular/core'
import { emitBox } from './boxcontrol'
import { Geometry } from 'ol/geom'
import { fromExtent } from 'ol/geom/Polygon'

export class fullBoxControl extends Control {
  constructor(
    public boxEmitter: EventEmitter<string>,
    // eslint-disable-next-line @typescript-eslint/no-explicit-any
    opt_options: any
  ) {
    const options = opt_options || {}

    const button = document.createElement('button')

    const element = document.createElement('div')
    element.className = 'fullboxcontrol ol-unselectable ol-control'
    button.title = 'Get Features in map'
    element.appendChild(button)

    super({
      element: element,
      target: options.target,
    })

    button.innerHTML = `
        <svg xmlns="http://www.w3.org/2000/svg" viewBox="0 0 512 512"><!--!Font Awesome Free 6.5.2 by @fontawesome - https://fontawesome.com License - https://fontawesome.com/license/free Copyright 2024 Fonticons, Inc.--><path d="M86.6 9.4C74.1-3.1 53.9-3.1 41.4 9.4s-12.5 32.8 0 45.3L122.7 136 30.6 228.1c-37.5 37.5-37.5 98.3 0 135.8L148.1 481.4c37.5 37.5 98.3 37.5 135.8 0L474.3 290.9c28.1-28.1 28.1-73.7 0-101.8L322.9 37.7c-28.1-28.1-73.7-28.1-101.8 0L168 90.7 86.6 9.4zM168 181.3l49.4 49.4c12.5 12.5 32.8 12.5 45.3 0s12.5-32.8 0-45.3L213.3 136l53.1-53.1c3.1-3.1 8.2-3.1 11.3 0L429.1 234.3c3.1 3.1 3.1 8.2 0 11.3L386.7 288H67.5c1.4-5.4 4.2-10.4 8.4-14.6L168 181.3z"/></svg>      `

    button.addEventListener('click', this.addFullBox.bind(this), false)
  }

  addFullBox() {
    const map = this.getMap()!
    const extent = map.getView().calculateExtent(map.getSize())
    const extent2 = extent // transformExtent(extent, 'EPSG:3857', 'EPSG:4326')
    const polygon = fromExtent(extent2) as Geometry
    emitBox(map, polygon, this.boxEmitter)
  }
}
