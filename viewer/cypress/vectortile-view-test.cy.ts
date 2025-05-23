import { HttpClientModule } from '@angular/common/http'
import { LoggerModule } from 'ngx-logger'
import { environment } from '../src/environments/environment'
import { VectortileViewComponent } from '../src/app/vectortile-view/vectortile-view.component'

describe('Vectortiled-view-test.cy.ts', () => {
  it.skip('Skipped unable to supply vectortile as feature yet, schould Write File Test', function () {
    cy.request('GET', 'https://data.example.com/dataset/ogc/v1/tiles/NetherlandsRDNewQuad/2/1/1?f=mvt').then(resp => {
      cy.log(resp.body.length)
      cy.log(JSON.stringify(resp.headers))
      cy.writeFile('cypress/fixtures/vt1.mvt1', resp, null)
    })
  })

  it.skip('Skipped unable to supply vectortile as featureyet, should mounts and shows tiles', () => {
    cy.intercept('GET', 'https://data.example.com/dataset/ogc/v1/tiles/NetherlandsRDNewQuad/*/*/1?f=mvt', {
      fixture: 'fix-todo',
      statusCode: 200,
      headers: { 'content-encoding': 'gzip', 'content-type': 'application/vnd.mapbox-vector-tile' },
    }).as('vt1')

    cy.mount(VectortileViewComponent, {
      imports: [
        HttpClientModule,
        LoggerModule.forRoot({
          level: environment.loglevel,
        }),
      ],
      componentProperties: {
        tileUrl: 'https://data.example.com/dataset/ogc/v1/tiles/NetherlandsRDNewQuad',
        centerX: 5.3896944,
        centerY: 52.1562499,
        showGrid: true,
        showObjectInfo: true,
      },
    }).then(comp1 => {
      const map = comp1.component.map
      map.addEventListener('loadend', cy.stub().as('MapLoaded'))
      const viewport = map.getViewport()
      const position = viewport.getBoundingClientRect()
      cy.log(`left: ${position.left}, top: ${position.top}, width: ${position.width}, height: ${position.height}`)
    })
  })
  it('show achtergrond', () => {
    cy.intercept('GET', 'https://visualisation.example.com/teststyle*', { fixture: 'teststyle-fonts.json' }).as('style')
    cy.mount(VectortileViewComponent, {
      imports: [
        HttpClientModule,
        LoggerModule.forRoot({
          level: environment.loglevel,
        }),
      ],
      componentProperties: {
        id: 'test',
        tileUrl: 'https://api.pdok.nl/kadaster/kadastralekaart/ogc/v1-demo/tiles/NetherlandsRDNewQuad',
        styleUrl: 'https://visualisation.example.com/teststyle/',
        zoom: 12,
        centerX: 5.3896944,
        centerY: 52.1562499,
        showGrid: true,
        showObjectInfo: false,
      },
    }).then(comp1 => {
      const map = comp1.component.map
      map.addEventListener('loadend', cy.stub().as('MapLoaded'))
      const viewport = map.getViewport()
      const position = viewport.getBoundingClientRect()
      cy.log(`left: ${position.left}, top: ${position.top}, width: ${position.width}, height: ${position.height}`)
    })
  })
})
