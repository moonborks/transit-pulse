import { defineStore } from 'pinia'
import {
  type Stop,
  type ShapePoint,
  type Route,
  type Trip,
  type RouteAPI,
  type TripAPI,
  type NextStopAPI,
  type NextStop,
} from '@/types/mta'
import { useFetch } from '@/composables/api/useFetch'
import { computed } from 'vue'
import {
  normalizeNextStopFields,
  normalizeRouteFields,
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
    data: nextStops,
    loading: nextStopsLoading,
    error: nextStopsError,
    fetchData: fetchNextStops,
  } = useFetch<NextStopAPI[], NextStop[]>((data) => data.map(normalizeNextStopFields))

  const load = async () => {
    await Promise.all([
      fetchStops(endpoints.mta.stops.getAll),
      fetchShapes(endpoints.mta.shapes.getAllSimplified),
      fetchTrips(endpoints.mta.trips.getAllToday),
      fetchRoutes(endpoints.mta.routes.getAll),
      fetchNextStops(endpoints.mta.routes.getAllNextStops),
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
    nextStops,
    stopLocationLookup,
    stopsLoading,
    stopsError,
    shapesLoading,
    shapesError,
    routesLoading,
    routesError,
    tripsLoading,
    tripsError,
    nextStopsLoading,
    nextStopsError,
    load,
    getRouteColor,
  }
})
