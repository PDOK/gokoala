import { PROJJSONDefinition } from 'proj4/dist/lib/core'
import { ProjectionDefinition } from 'proj4/dist/lib/defs'

export interface CrsMap {
  [key: string]: string | ProjectionDefinition | PROJJSONDefinition
}
