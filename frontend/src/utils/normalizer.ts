import type { Route, RouteAPI, Trip, TripAPI } from '@/types/mta'

export function normalizeRouteFields(route: RouteAPI): Route {
  return {
    id: route.id,
    shortName: route.short_name,
    longName: route.long_name,
    type: route.type,
    color: route.color,
  }
}

export function normalizeTripFields(trip: TripAPI): Trip {
  return {
    id: trip.id,
    routeId: trip.route_id,
    serviceId: trip.service_id,
    headsign: trip.headsign,
    directionId: trip.direction_id,
    shapeId: trip.shape_id,
  }
}
