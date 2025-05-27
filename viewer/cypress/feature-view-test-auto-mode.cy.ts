import { intercept, mountFeatureComponent, screenshot, tests as projectenTests } from './shared'

function moveMap(x: number, y: number) {
  // Trigger a pointerdown event to start the panning action
  cy.get('.ol-viewport').trigger('pointerdown', {
    eventConstructor: 'MouseEvent',
    clientX: 100,
    clientY: 100,
    force: true,
    isPrimary: true,
  })

  // Calculate new coordinates 200 pixels away
  let newX = 100 + x
  let newY = 100 + y

  // Trigger pointermove event to simulate the panning action
  cy.get('.ol-viewport').trigger('pointermove', {
    eventConstructor: 'MouseEvent',
    clientX: newX,
    clientY: newY,
  })
  newX = newX + x
  newY = newY + y
  cy.get('.ol-viewport').trigger('pointermove', { clientX: newX, clientY: newY })

  cy.get('.ol-viewport').trigger('pointerup', {
    eventConstructor: 'MouseEvent',
    force: true,
    isPrimary: true,
  })
}

projectenTests
  // .filter(e => e.code === 'EPSG:28992')
  .forEach(i => {
    describe(i.geofix + '-feature-view-grid', () => {
      it('It emits boundingbox when moveing map in automode grid' + i.geofix + 'on BRT', () => {
        intercept('grid-' + i.geofix, true)

        const optionstring = JSON.stringify({
          font: 'bold 40px Arial, Verdana, Courier New',
        })
        const prop = {
          labelField: 'label',
          labelOptions: optionstring,
          fillColor: 'rgba(0,0,255,0)',
          itemsUrl: 'https://test/items',
        }
        mountFeatureComponent(i.projection, 'BRT', 'auto', prop)
        cy.get('.innersvg').should('not.exist')
        screenshot(i.code + '-1-auto-before-move')
        moveMap(-100, -100)
        screenshot(i.code + '-2-auto-after-move')
        cy.get('@boxSpy').should((spy: any) => {
          const firstCallArgs = spy.getCall(0).args[0].split(',')
          expect(firstCallArgs[0]).to.match(/^4./)
          expect(firstCallArgs[1]).to.match(/^52./)
        })
      })

      describe(i.geofix + '-feature-view-grid', () => {
        it('It emits boundingbox when moveing map in automode grid' + i.geofix + 'on BRT', () => {
          intercept('grid-' + i.geofix, true)

          const optionstring = JSON.stringify({
            font: 'bold 40px Arial, Verdana, Courier New',
          })
          const prop = {
            labelField: 'label',
            labelOptions: optionstring,
            fillColor: 'rgba(0,0,255,0)',
            itemsUrl: 'https://test/items',
          }
          mountFeatureComponent(i.projection, 'OSM', 'auto', prop)
          cy.get('.innersvg').should('not.exist')
          screenshot(i.code + '-1-auto-before-move-OSM')
          moveMap(-100, -100)
          screenshot(i.code + '-2-auto-after-move-OSM')
          cy.get('@boxSpy').should((spy: any) => {
            const firstCallArgs = spy.getCall(0).args[0].split(',')
            expect(firstCallArgs[0]).to.match(/^4./)
            expect(firstCallArgs[1]).to.match(/^52./)
          })
        })
      })

      it('It emits boundingbox when zoomin out in automode grid' + i.geofix + 'on BRT', () => {
        //  cy.intercept('GET', 'https://test*', generateRDSquareGrid(200000, 300000, 100, 1)).as('geo')
        intercept('grid-' + i.geofix, true)

        const optionstring = JSON.stringify({
          font: 'bold 40px Arial, Verdana, Courier New',
        })
        const prop = {
          labelField: 'label',
          labelOptions: optionstring,
          fillColor: 'rgba(0,0,255,0)',
          itemsUrl: 'https://test/items',
          maxFitScale: 3000,
        }
        mountFeatureComponent(i.projection, 'BRT', 'auto', prop)
        cy.get('.innersvg').should('not.exist')
        screenshot(i.code + '-3-BRT-bbox-auto-before-zoom')
        cy.get('.ol-zoom-out').click()
        screenshot(i.code + '-4-BRT-bbox-auto-after-zoom')
        cy.get('@boxSpy').should('have.been.calledTwice')
      })
    })
  })
