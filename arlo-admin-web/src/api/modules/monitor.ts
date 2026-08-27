import request from '../request'

export interface OnlineSessionItem {
  userId: number
  username: string
  sessionId: string
  ip: string
  browser: string
  os: string
  loginAt: string
}

export interface OnlineListResult {
  list: OnlineSessionItem[]
  total: number
  page: number
  pageSize: number
}

export function getOnlineList(params: { username?: string; page: number; pageSize: number }) {
  return request.get<OnlineListResult>('/v1/monitor/online/list', params)
}

export function kickOnlineUser(data: { userId: number; sessionId?: string }) {
  return request.post('/v1/monitor/online/kick', data)
}

export interface ServerMonitorInfo {
  cpu: {
    cores: number
    usagePct: number
    load1: number
    load5: number
    load15: number
  }
  mem: {
    total: number
    used: number
    available: number
    usagePct: number
  }
  disk: Array<{
    mount: string
    fsType: string
    total: number
    used: number
    free: number
    usagePct: number
  }>
  sys: {
    os: string
    arch: string
    hostname: string
    uptime: number
  }
  go: {
    version: string
    goroutines: number
    gomaxprocs: number
    heapAlloc: number
    heapSys: number
    heapInuse: number
    gcCPUFraction: number
    numGC: number
    lastGC: string
  }
  app: {
    name: string
    mode: string
    startTime: string
    runSeconds: number
    goVersion: string
  }
  db: {
    status: string
    pingMs: number
    open: number
    inUse: number
    idle: number
    maxOpen: number
  }
  redis: {
    status: string
    pingMs: number
  }
}

export function getServerMonitor() {
  return request.get<ServerMonitorInfo>('/v1/monitor/server')
}
