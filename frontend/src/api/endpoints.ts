const BASE_URL = '/api'

export const endpoints = {
  mta: {
    stops: {
      getAll: `${BASE_URL}/mta/stops`,
    },
    shapes: {
      getAll: `${BASE_URL}/mta/shapes`,
    },
    trips: {
      getAll: `${BASE_URL}/mta/trips`,
    },
    routes: {
      getAll: `${BASE_URL}/mta/routes`,
      getAllNextStops: `${BASE_URL}/mta/routes/next-stops`,
    },
  },
}
