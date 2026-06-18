export interface Route {
  id: string;
  shortName: string;
  longName: string;
  type: number;
  color: string;
}

export interface RouteAPI {
  id: string;
  short_name: string;
  long_name: string;
  type: number;
  color: string;
}

export interface Stop {
  id: string;
  name: string;
  lat: number;
  lon: number;
}

export interface ShapePoint {
  id: string;
  sequence: number;
  lat: number;
  lon: number;
}

export interface Trip {
  id: string;
  routeId: string;
  serviceId: string;
  headsign: string;
  directionId: 0 | 1;
  shapeId: string | null;
}

export interface TripAPI {
  id: string;
  route_id: string;
  service_id: string;
  headsign: string;
  direction_id: 0 | 1;
  shape_id: string | null;
}
