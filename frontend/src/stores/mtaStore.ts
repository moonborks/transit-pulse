import { defineStore } from 'pinia'
import {
  type Stop,
  type ShapePoint,
  type Route,
  type Trip,
  type RouteAPI,
  type TripAPI,
  type TrainLocation,
  type TrainLocationAPI,
} from '@/types/mta'
import { useFetch } from '@/composables/api/useFetch'
import { computed } from 'vue'
import {
  normalizeRouteFields,
  normalizeTrainLocationFields,
  normalizeTripFields,
} from '@/utils/normalizer'
import { endpoints } from '@/api/endpoints'

export const useMtaStore = defineStore('mta', () => {
  const {
    data: stops,
    loading: stopsLoading,
    error: stopsError,
    fetchData: fetchStops,
  } = useFetch<Stop[]>()
  const {
    data: shapes,
    loading: shapesLoading,
    error: shapesError,
    fetchData: fetchShapes,
  } = useFetch<ShapePoint[]>()
  const {
    data: trips,
    loading: tripsLoading,
    error: tripsError,
    fetchData: fetchTrips,
  } = useFetch<TripAPI[], Trip[]>((data) => data.map(normalizeTripFields))
  const {
    data: routes,
    loading: routesLoading,
    error: routesError,
    fetchData: fetchRoutes,
  } = useFetch<RouteAPI[], Route[]>((data) => data.map(normalizeRouteFields))

  const {
    data: trainLocations,
    loading: trainLocationsLoading,
    error: trainLocationsError,
    fetchData: fetchTrainLocations,
  } = useFetch<TrainLocationAPI[], TrainLocation[]>((data) =>
    data.map(normalizeTrainLocationFields),
  )

  const load = async () => {
    await Promise.all([
      fetchStops(endpoints.mta.stops.getAll),
      fetchShapes(endpoints.mta.shapes.getAll),
      fetchTrips(endpoints.mta.trips.getAllToday),
      fetchRoutes(endpoints.mta.routes.getAll),
      // fetchNextStops(endpoints.mta.routes.getAllNextStops),
      fetchTrainLocations(endpoints.mta.trips.getLocations),
    ])
  }

  const groupedShapes = computed(() => {
    return groupShapePoints(shapes.value)
  })

  function groupShapePoints(points: ShapePoint[] | null): Record<string, [number, number][]> {
    const grouped: Record<string, [number, number][]> = {}
    if (points === null) {
      return grouped
    }
    for (const p of points) {
      const pts = grouped[p.id] ?? []
      pts.push([p.lon, p.lat])
      grouped[p.id] = pts
    }
    return grouped
  }

  const routeColorMap = computed(() => {
    const map = new Map<string, string>()
    for (const route of routes.value ?? []) {
      map.set(route.id, `#${route.color}`)
    }
    return map
  })

  function getRouteColor(routeId: string): string {
    return routeColorMap.value.get(routeId) ?? '#888888'
  }

  const stopLocationLookup = computed(() => {
    const lookup = new Map<string, Stop>()
    for (const stop of stops.value ?? []) {
      lookup.set(stop.id, stop)
    }
    return lookup
  })

  return {
    stops,
    groupedShapes,
    routes,
    trips,
    trainLocations,
    stopLocationLookup,
    stopsLoading,
    stopsError,
    shapesLoading,
    shapesError,
    routesLoading,
    routesError,
    tripsLoading,
    tripsError,
    trainLocationsLoading,
    trainLocationsError,
    load,
    getRouteColor,
  }
})
