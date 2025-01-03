import { HttpClientModule } from '@angular/common/http'
import { LoggerModule } from 'ngx-logger'
import { environment } from './../src/environments/environment'
import { LocationSearchComponent } from './../src/app/location-search/location-search.component'
import { checkAccessibility, injectAxe } from './shared'
import { SeachResponse } from './seachResponse-model'

function loadLocationSearchEmpty() {
  cy.mount(LocationSearchComponent, {
    imports: [
      HttpClientModule,
      LoggerModule.forRoot({
        level: environment.loglevel,
      }),
    ],
  })
}
function loadLocationSearchWithUrl() {
  cy.intercept('GET', 'https://visualisation.example.com/locationapi/collections', { fixture: 'collectionfix.json' }).as('col')
  cy.intercept('GET', 'https://visualisation.example.com/locationapi/search?*', { fixture: 'search-den-wgs84.json' }).as('search')
  cy.intercept('GET', 'https://example.com/ogc/v1/collections/addresses/items/827*', { fixture: 'amsterdam-wgs84.json' }).as('geo')
  cy.intercept('GET', 'https://example.com/ogc/v1/collections/addresses/items/22215*', { fixture: 'grid-amsterdam-wgs84.json' }).as('geo2')
  cy.intercept('GET', 'https://tile.openstreetmap.org/**/*', { fixture: 'backgroundstub.png' }).as('background')

  cy.mount(LocationSearchComponent, {
    imports: [
      HttpClientModule,
      LoggerModule.forRoot({
        level: environment.loglevel,
      }),
    ],
    componentProperties: {
      url: 'https://visualisation.example.com/locationapi',

      //    backgroundmap: 'BRT',
    },
  })
  cy.wait('@col')
}

function loadLocationSearch(fixture: string, labelText: string, placeholder: string, title: string) {
  cy.intercept('GET', 'https://visualisation.example.com/locationapi*', { fixture: 'collectionfix.json' }).as('col')
  cy.mount(LocationSearchComponent, {
    imports: [
      HttpClientModule,
      LoggerModule.forRoot({
        level: environment.loglevel,
      }),
    ],
    componentProperties: {
      url: 'https://visualisation.example.com/locationapi',
      label: labelText,
      placeholder,
      title,
    },
  })
}

describe('location-search-test', () => {
  it('it show no url message', () => {
    loadLocationSearchEmpty()
    cy.get('body').contains('please provide url to location url')
  })

  it('Has no detectable a11y violations on mount and show default values', () => {
    injectAxe()
    loadLocationSearchWithUrl()
    cy.get('#searchboxlabel').should('have.text', 'Search location')
    cy.get('#searchBox')
      .should('have.attr', 'placeholder', 'Enter location to search')
      .should('have.attr', 'title', 'Enter the location you want to search for')
    checkAccessibility('body')
  })

  it('Label can be changed', () => {
    injectAxe()
    const expectedText = 'label can be changed'
    const expectedPlaceholder = 'placeholder'
    const expectedTitle = 'titel'
    loadLocationSearch('', expectedText, expectedPlaceholder, expectedTitle)
    cy.get('label').should('have.text', expectedText)
    cy.get('#searchBox').should('have.attr', 'placeholder', expectedPlaceholder).should('have.attr', 'title', expectedTitle)
    checkAccessibility('body')
  })

  it('can have search input', () => {
    loadLocationSearchWithUrl()
    cy.get('#searchBox').should('be.visible').should('be.enabled').type('A')
  })

  it('should verify all checkboxes are checked', () => {
    loadLocationSearchWithUrl()
    cy.get('button').should('have.attr', 'title', 'show/hide search options').click()
    cy.get('input[type="checkbox"').each($checkbox => {
      cy.wrap($checkbox).should('be.checked').should('be.enabled')
    })
  })

  it('verify all titles from collections', () => {
    loadLocationSearchWithUrl()
    cy.get('button').should('have.attr', 'title', 'show/hide search options').click()
    const expectedLabels = ['functioneel_gebied', 'geografisch_gebied', 'ligplaats', 'standplaats', 'verblijfsobject', 'woonplaats']
    expectedLabels.forEach(label => {
      // Verify the checkbox is checked
      //cy.get(`input[type="checkbox"][value="${label}"]`).should('be.checked')

      // Verify the label text
      cy.get('body').contains(label)
    })
  })

  it('disable collection and typeahead', () => {
    loadLocationSearchWithUrl()
    cy.get('button').should('have.attr', 'title', 'show/hide search options').click()
    cy.contains('ligplaats').find('input[type="checkbox"]').uncheck()
    cy.contains('standplaats').find('input[type="checkbox"]').uncheck()
    cy.contains('verblijfsobject').find('input[type="checkbox"]').uncheck()
    cy.get('input[placeholder="Enter Relevance for woonplaats"]').type('{backspace}{backspace}0.8')
    cy.get('#searchBox').should('be.visible').should('be.enabled').type('den')
    cy.wait('@search')
    cy.wait('@search')
    cy.wait('@search')
    cy.get('@search').then(res => {
      const result = res as unknown as SeachResponse
      const r = result.request.query
      expect(r.q).to.equal('den')
      expect(r.functioneel_gebied.version).to.equal('1')
      expect(r.geografisch_gebied.version).to.equal('1')
      expect(r.woonplaats.version).to.equal('1')
      expect(r.woonplaats.relevance).to.equal('0.8')
      expect(result.request.url).to.equal(
        'https://visualisation.example.com/locationapi/search?q=den&functioneel_gebied%5Brelevance%5D=0.5&functioneel_gebied%5Bversion%5D=1&geografisch_gebied%5Brelevance%5D=0.5&geografisch_gebied%5Bversion%5D=1&woonplaats%5Brelevance%5D=0.8&woonplaats%5Bversion%5D=1'
      )
    })

    cy.contains('Beatrixlaan').focus()
    //cy.wait('@geo')
    //cy.wait('@background')
  })
})
