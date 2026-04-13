type LogLevel = 'debug' | 'info' | 'warn' | 'error'

interface LogEntry {
  level: LogLevel
  component: string
  message: string
  data?: unknown
  timestamp: string
}

const LOG_LEVELS: Record<LogLevel, number> = {
  debug: 0,
  info: 1,
  warn: 2,
  error: 3,
}

const currentLevel = import.meta.env.DEV ? 'debug' : 'error'

function formatLogEntry(entry: LogEntry): string {
  const dataStr = entry.data ? ` ${JSON.stringify(entry.data)}` : ''
  return `[${entry.level.toUpperCase()}] ${entry.timestamp} ${entry.component}: ${entry.message}${dataStr}`
}

function log(level: LogLevel, component: string, message: string, data?: unknown): void {
  if (LOG_LEVELS[level] < LOG_LEVELS[currentLevel]) {
    return
  }

  const entry: LogEntry = {
    level,
    component,
    message,
    data,
    timestamp: new Date().toISOString(),
  }

  const formatted = formatLogEntry(entry)

  switch (level) {
    case 'debug':
      console.debug(formatted)
      break
    case 'info':
      console.info(formatted)
      break
    case 'warn':
      console.warn(formatted)
      break
    case 'error':
      console.error(formatted)
      break
  }
}

export const logger = {
  debug(component: string, message: string, data?: unknown): void {
    log('debug', component, message, data)
  },

  info(component: string, message: string, data?: unknown): void {
    log('info', component, message, data)
  },

  warn(component: string, message: string, data?: unknown): void {
    log('warn', component, message, data)
  },

  error(component: string, message: string, data?: unknown): void {
    log('error', component, message, data)
  },
}
