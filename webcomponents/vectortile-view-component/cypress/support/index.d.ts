declare namespace Cypress {
  interface Chainable {
    /**
     * Custom command to ... add your description here
     * @example cy.clickOnMyJourneyInCandidateCabinet()
     */
    simulateOpenLayersEvent(map, type, x, y): Chainable
  }
}
