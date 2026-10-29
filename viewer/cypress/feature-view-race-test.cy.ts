import { provideHttpClient } from '@angular/common/http'
import { SimpleChange } from '@angular/core'
import { LoggerModule, NgxLoggerLevel } from 'ngx-logger'
import { FeatureViewComponent } from 'src/app/feature-view/feature-view.component'
import { FeatureService } from 'src/app/shared/services/feature.service'

const SLOW_URL = 'https://test/slow/items'
const FAST_URL = 'https://test/fast/items'
const SLOW_DELAY_MS = 400
const FAST_DELAY_MS = 900

describe('feature view race condition when itemUrls change', () => {
  beforeEach(() => {
    cy.viewport(550, 750)
    cy.intercept('GET', 'https://tile.openstreetmap.org/*/*/*.png', { fixture: '172300.png' })
    cy.intercept('GET', `${SLOW_URL}*`, { fixture: 'amsterdam-wgs84.json', delay: SLOW_DELAY_MS }).as('slowItems')
    cy.intercept('GET', `${FAST_URL}*`, { fixture: 'amsterdam.json', delay: FAST_DELAY_MS }).as('fastItems')
  })

  function mountAndSwitchItemUrls(): Cypress.Chainable<FeatureViewComponent> {
    return cy
      .mount(FeatureViewComponent, {
        providers: [provideHttpClient(), FeatureService],
        imports: [LoggerModule.forRoot({ level: NgxLoggerLevel.DEBUG })],
        componentProperties: { backgroundMap: 'OSM', mode: 'default', itemUrls: [SLOW_URL] },
      })
      .then(({ component }) => {
        cy.get('@slowItems.all').should('have.length', 1)
        cy.then(() => {
          component.itemUrls = [FAST_URL]
          component.ngOnChanges({ itemUrls: new SimpleChange([SLOW_URL], [FAST_URL], false) } as never)
        })
        cy.then(() => component)
      }) as unknown as Cypress.Chainable<FeatureViewComponent>
  }

  it('ignores the response of the previous itemUrls', () => {
    mountAndSwitchItemUrls().then(component => {
      cy.wait(SLOW_DELAY_MS + 200).then(() => {
        expect(component.features.length).to.equal(0)
      })
      cy.wait('@fastItems')
        .then(() => cy.wait(100))
        .then(() => {
          expect(component.features.length).to.equal(1)
        })
    })
  })

  it('keeps loading until the latest request completes', () => {
    mountAndSwitchItemUrls().then(component => {
      cy.wait(SLOW_DELAY_MS + 200).then(() => {
        expect(component.loading()).to.equal(true)
      })
      cy.wait('@fastItems')
        .then(() => cy.wait(100))
        .then(() => {
          expect(component.loading()).to.equal(false)
        })
    })
  })
})
