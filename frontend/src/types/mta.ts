export interface Route {
  id: string
  shortName: string
  longName: string
  type: number
  color: string
}

export interface RouteAPI {
  id: string
  short_name: string
  long_name: string
  type: number
  color: string
}

export interface Stop {
  id: string
  name: string
  lat: number
  lon: number
}

export interface ShapePoint {
  id: string
  sequence: number
  lat: number
  lon: number
}

export interface Trip {
  routeId: string
  headsign: string
  shapeId: string | null
}

export interface TripAPI {
  route_id: string
  headsign: string
  shape_id: string | null
}
