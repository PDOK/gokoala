import { HttpClientModule } from '@angular/common/http'
import { LoggerModule } from 'ngx-logger'
import { LegendViewComponent } from 'src/app/legend-view/legend-view.component'
import { environment } from 'src/environments/environment'
import { downloadPng, screenshot } from './shared'

describe('Legend-view-test.cy.ts', () => {
  it('mounts and shows legend items', () => {
    cy.intercept('GET', 'https://teststyle*', { fixture: 'teststyle.json' }).as('geo')

    cy.mount(LegendViewComponent, {
      imports: [
        HttpClientModule,
        LoggerModule.forRoot({
          level: environment.loglevel,
        }),
      ],
      componentProperties: {
        styleUrl: 'https://teststyle/',
      },
    })

    cy.wait(1000)

    cy.wrap(['1', '2', '3']).each(n => {
      const textsel = ':nth-child(' + n + ') > .legendText'
      cy.get(textsel).then($value => {
        const textValue = $value.text()

        const sel = ':nth-child(' + n + ') > app-legend-item > #itemmap > .ol-viewport > .ol-unselectable > .ol-layer > canvas'
        downloadPng(sel, textValue + '.png')
      })
    })
    cy.get(':nth-child(1) > .legendText').contains('TestArea')
    cy.get(':nth-child(2) > .legendText').contains('Name')
    cy.get(':nth-child(3) > .legendText').contains('Testline')
    screenshot('legend')
  })
})
