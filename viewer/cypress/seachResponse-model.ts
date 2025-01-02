export interface SeachResponse {
  id: string
  browserRequestId: string
  routeId: string
  request: Request
  state: string
  requestWaited: boolean
  responseWaited: boolean
  subscriptions: any[]
  response: Response
}

export interface Request {
  headers: RequestHeaders
  url: string
  method: string
  httpVersion: string
  resourceType: string
  query: Query
  body: string
  responseTimeout: number
  alias: string
}

export interface RequestHeaders {
  host: string
  connection: string
  pragma: string
  'cache-control': string
  'sec-ch-ua-platform': string
  'user-agent': string
  accept: string
  'sec-ch-ua': string
  'sec-ch-ua-mobile': string
  origin: string
  'sec-fetch-site': string
  'sec-fetch-mode': string
  'sec-fetch-dest': string
  referer: string
  'accept-encoding': string
  'accept-language': string
}

export interface Query {
  q: string
  functioneel_gebied: FunctioneelGebied
  geografisch_gebied: FunctioneelGebied
  ligplaats: FunctioneelGebied
  standplaats: FunctioneelGebied
  verblijfsobject: FunctioneelGebied
  woonplaats: FunctioneelGebied
}

export interface FunctioneelGebied {
  relevance: string
  version: string
}

export interface Response {
  headers: ResponseHeaders
  body: Body
  url: string
  method: null
  httpVersion: null
  statusCode: number
  statusMessage: null
}

export interface Body {
  type: string
  timeStamp: Date
  links: Link[]
  features: Feature[]
  numberReturned: number
}

export interface Feature {
  type: string
  properties: Properties
  geometry: Geometry
  id: string
  links: Link[]
}

export interface Geometry {
  type: string
  coordinates: Array<Array<number[]>>
}

export interface Link {
  rel: string
  title: string
  type: string
  href: string
}

export interface Properties {
  collectionGeometryType: string
  collectionId: string
  collectionVersion: string
  displayName: string
  highlight: string
  href: string
  score: number
}

export interface ResponseHeaders {
  'content-type': string
  'access-control-allow-origin': string
  'access-control-allow-credentials': string
}
