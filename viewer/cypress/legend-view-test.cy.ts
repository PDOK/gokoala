import { HttpClientModule } from '@angular/common/http'
import { LoggerModule } from 'ngx-logger'
import { LegendViewComponent } from './../src/app/legend-view/legend-view.component'

import { checkAccessibility, downloadPng, injectAxe, screenshot } from './shared'
import { environment } from '../src/environments/environment'

function loadlegend(fixture: string) {
  cy.intercept('GET', 'https://visualisation.example.com/teststyle*', { fixture: fixture }).as('geo')
  cy.mount(LegendViewComponent, {
    imports: [
      HttpClientModule,
      LoggerModule.forRoot({
        level: environment.loglevel,
      }),
    ],
    componentProperties: {
      styleUrl: 'https://visualisation.example.com/teststyle/',
    },
  })
  cy.wrap(['1', '2', '3']).each(n => {
    const textsel = ':nth-child(' + n + ') > .legendText'
    cy.get(textsel).then($value => {
      const textValue = $value.text()

      const sel = ':nth-child(' + n + ') > app-legend-item > #itemmap > .ol-viewport > .ol-unselectable > .ol-layer > canvas'
      downloadPng(sel, textValue + '.png')
    })
  })
}

describe('Legend-view-test', () => {
  it('Has no detectable a11y violations on mount', () => {
    cy.intercept('GET', 'https://visualisation.example.com/teststyle*', { fixture: 'teststyle.json' }).as('geo')
    injectAxe()

    cy.mount(LegendViewComponent, {
      imports: [
        HttpClientModule,
        LoggerModule.forRoot({
          level: environment.loglevel,
        }),
      ],
      componentProperties: {
        styleUrl: 'https://visualisation.example.com/teststyle/',
      },
    })
    checkAccessibility('body')
  })

  it('mounts and shows legend items from style without metadata', () => {
    loadlegend('teststyle.json')
    cy.get(':nth-child(1) > .legendText').contains('TestArea')
    cy.get(':nth-child(2) > .legendText').contains('Name')
    cy.get(':nth-child(3) > .legendText').contains('Testline')
    screenshot('legend')
  })

  it('mounts and shows legend items from style with metadata with "gokoala:title-items": "id" ', () => {
    loadlegend('teststyle-id.json')
    cy.get(':nth-child(1) > .legendText').contains('Area label Name')
    cy.get(':nth-child(2) > .legendText').contains('Area print border')
    cy.get(':nth-child(3) > .legendText').contains('circle')
    cy.get(':nth-child(4) > .legendText').contains('line')
    screenshot('legend')
  })

  it('mounts and shows legend items from style with metadata with "gokoala:title-items": "color,function" ', () => {
    loadlegend('teststyle-filter.json')
    cy.get(':nth-child(1) > .legendText').contains('Green B')
    cy.get(':nth-child(2) > .legendText').contains('Red A')
    cy.get(':nth-child(3) > .legendText').contains('Red A Name')

    screenshot('legend')
  })
})
