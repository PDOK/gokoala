import { createOutputSpy } from 'cypress/angular'
import { of } from 'rxjs'
import { LoggerModule, NgxLoggerLevel } from 'ngx-logger'
import { LocationSearchViewComponent } from 'src/app/location-search-view/location-search-view.component'
import { CollectionsService } from 'src/app/shared/services/collections.service'
import { FeatureService } from 'src/app/shared/services/feature.service'
import { screenshot } from './shared'

const stubCollections = [{ id: 'dutch-addresses', title: 'Dutch Addresses', version: 1, links: [] }]

const stubFeatures = [
  {
    type: 'Feature',
    id: '103',
    properties: {
      collection_geometry_type: 'POINT',
      collection_id: 'dutch-addresses',
      collection_version: 1,
      display_name: 'Amstel 3 1011PN Amsterdam',
      highlight: '<b>Amstel</b> 3 1011PN <b>Amsterdam</b>',
      href: ['http://localhost:8080/collections/dutch-addresses/items/103?f=json'],
      score: 0.92,
    },
    geometry: { type: 'Point', coordinates: [4.8998, 52.3676] },
  },
  {
    type: 'Feature',
    id: '104',
    properties: {
      collection_geometry_type: 'POINT',
      collection_id: 'dutch-addresses',
      collection_version: 1,
      display_name: 'Amsteldijk 1 1074HP Amsterdam',
      highlight: '<b>Amstel</b>dijk 1 1074HP <b>Amsterdam</b>',
      href: ['http://localhost:8080/collections/dutch-addresses/items/104?f=json'],
      score: 0.87,
    },
    geometry: { type: 'Point', coordinates: [4.9012, 52.3501] },
  },
]

function mountSearchComponent() {
  cy.viewport(800, 600)

  cy.mount(LocationSearchViewComponent, {
    providers: [
      { provide: CollectionsService, useValue: { getCollections: () => of(stubCollections) } },
      { provide: FeatureService, useValue: { queryFeatures: () => of(stubFeatures) } },
    ],
    imports: [
      LoggerModule.forRoot({
        level: NgxLoggerLevel.DEBUG,
      }),
    ],
    componentProperties: {
      locationSelected: createOutputSpy('locationSelectedSpy'),
    },
  })
}

describe('location-search-view', () => {
  it('shows results when typing and confirms selection on click', () => {
    mountSearchComponent()

    cy.get('#search-input').type('Amst')

    cy.get('[role="listbox"]').should('be.visible')
    cy.get('[role="option"]').should('have.length', 2)
    cy.get('[role="option"]').first().should('contain.text', 'Amstel')

    cy.get('[role="option"]').first().find('button').click()

    cy.get('#search-input').should('have.value', 'Amstel 3 1011PN Amsterdam')
    cy.get('@locationSelectedSpy')
      .should('have.been.called')
      // eslint-disable-next-line @typescript-eslint/no-explicit-any
      .should((spy: any) => {
        const emittedUrls: string[] = spy.lastCall.args[0]
        expect(emittedUrls[0]).to.include('/collections/dutch-addresses/items/103')
      })

    cy.get('[role="listbox"]').should('not.have.class', 'show')

    screenshot('search-result-selected')
  })

})
