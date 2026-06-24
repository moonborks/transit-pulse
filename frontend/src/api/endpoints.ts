const BASE_URL = '/api'

export const endpoints = {
  mta: {
    stops: {
      getAll: `${BASE_URL}/mta/stops`,
    },
    shapes: {
      getAll: `${BASE_URL}/mta/shapes`,
      getAllSimplified: `${BASE_URL}/mta/shapes?simplify=true`,
    },
    trips: {
      getAll: `${BASE_URL}/mta/trips`,
      getMessages: `${BASE_URL}/mta/trips/messages`,
      getAllToday: `${BASE_URL}/mta/trips/today`,
      getLocations: `${BASE_URL}/mta/trips/positions`,
    },
    routes: {
      getAll: `${BASE_URL}/mta/routes`,
      getAllNextStops: `${BASE_URL}/mta/routes/next-stops`,
    },
  },
}
