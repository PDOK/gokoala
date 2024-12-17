import { HttpClientModule } from '@angular/common/http'
import { LoggerModule } from 'ngx-logger'
import { environment } from './../src/environments/environment'
import { LocationSearchComponent } from './../src/app/location-search/location-search.component'
import { checkAccessibility, injectAxe } from './shared'

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
  cy.intercept('GET', 'https://visualisation.example.com/locationapi/search', { fixture: 'search-den-wgs84.json' }).as('search')

  cy.mount(LocationSearchComponent, {
    imports: [
      HttpClientModule,
      LoggerModule.forRoot({
        level: environment.loglevel,
      }),
    ],
    componentProperties: {
      url: 'https://visualisation.example.com/locationapi',
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
    cy.get('input[type="checkbox"').each($checkbox => {
      cy.wrap($checkbox).should('be.checked').should('be.enabled')
    })
  })

  it('verify all titles from collections', () => {
    loadLocationSearchWithUrl()
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
    cy.get(':nth-child(3) >  label > input[type=checkbox]').uncheck()
    cy.get(':nth-child(6) >  label > input[type=checkbox]').uncheck()
    cy.get('#searchBox').should('be.visible').should('be.enabled').type('den')
    cy.wait('@search')
    cy.wait('@search')
    cy.wait('@search')
    cy.contains('Beatrixlaan').focus()

  })
})
