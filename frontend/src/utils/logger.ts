import log from 'loglevel'
log.setLevel(import.meta.env.PROD ? 'warn' : 'debug')
export default log
