import { HttpClientModule } from '@angular/common/http'
import { LoggerModule } from 'ngx-logger'
import { environment } from 'src/environments/environment'
import { VectortileViewComponent } from 'src/app/vectortile-view/vectortile-view.component'

describe('Vectortiled-view-test.cy.ts', () => {
  it.skip('Skipped unable to supply vectortile as feature yet, schould Write File Test', function () {
    cy.request('GET', 'https://api.pdok.nl/kadaster/bestuurlijkegebieden/ogc/v1_0-preprod/tiles/NetherlandsRDNewQuad/2/1/1?f=mvt').then(
      resp => {
        cy.log(resp.body.length)
        cy.log(JSON.stringify(resp.headers))
        cy.writeFile('cypress/fixtures/vt1.mvt1', resp, null)
      }
    )
  })

  it.skip('Skipped unable to supply vectortile as featureyet, should mounts and shows tiles', () => {
    cy.intercept('GET', 'https://api.pdok.nl/kadaster/bestuurlijkegebieden/ogc/v1_0-preprod/tiles/NetherlandsRDNewQuad/*/*/1?f=mvt', {
      fixture: 'fix-todo',
      statusCode: 200,
      headers: {
        'access-control-allow-origin': '*',
        'access-control-expose-headers': 'Content-Crs,Link',
        'api-version': '0.1.0',
        'content-encoding': 'gzip',
        'content-length': '9475',
        'content-type': 'application/vnd.mapbox-vector-tile',
        date: 'Mon, 15 Jan 2024 16:31:49 GMT',
        etag: '0x8DBBD9891CEC887',
        'last-modified': 'Mon, 25 Sep 2023 07:25:18 GMT',
        'x-ms-meta-md5sum': '0ea1c6818378a22ebe1b1fd2280f27a5',
        'strict-transport-security': 'max-age=31536000; includeSubDomains; preload',
      },
    }).as('vt1')

    cy.mount(VectortileViewComponent, {
      imports: [
        HttpClientModule,
        LoggerModule.forRoot({
          level: environment.loglevel,
        }),
      ],
      componentProperties: {
        tileUrl: 'https://api.pdok.nl/kadaster/bestuurlijkegebieden/ogc/v1_0-preprod/tiles/NetherlandsRDNewQuad',
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
})
