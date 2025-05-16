export interface Search {
  type: string
  timeStamp: Date
  links: Link[]
  features: SearchFeature[]
  numberReturned: number
}

export interface SearchFeature {
  type: string
  properties: SearchProperties
  geometry: SearchGeometry
  id: string
  links: Link[]
}

export interface SearchGeometry {
  type: string
  coordinates: Array<Array<number[]>>
}

export interface Link {
  rel: string
  title: string
  type: string
  href: string
}

export interface SearchProperties {
  collectionGeometryType: string
  collectionId: string
  collectionVersion: string
  displayName: string
  highlight: string
  href: string
  score: number
}
