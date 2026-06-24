import type {
  NextStop,
  NextStopAPI,
  Route,
  RouteAPI,
  TrainLocation,
  TrainLocationAPI,
  Trip,
  TripAPI,
} from '@/types/mta'

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
    routeId: trip.route_id,
    headsign: trip.headsign,
    shapeId: trip.shape_id,
  }
}

export function normalizeNextStopFields(nextStop: NextStopAPI): NextStop {
  return {
    stopId: nextStop.stop_id,
    tripId: nextStop.trip_id,
    routeId: nextStop.route_id,
    arrivalTime: nextStop.arrival_time,
  }
}

export function normalizeTrainLocationFields(trainLocation: TrainLocationAPI): TrainLocation {
  return {
    tripId: trainLocation.trip_id,
    routeId: trainLocation.route_id,
    lat: trainLocation.lat,
    lon: trainLocation.lon,
    bearing: trainLocation.bearing,
    nextStopId: trainLocation.next_stop_id,
  }
}
