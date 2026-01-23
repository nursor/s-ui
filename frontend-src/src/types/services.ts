import { Listen } from "./inbounds"
import { iTls } from "./tls"

export const SrvTypes = {
  DERP: 'derp',
  Resolved: 'resolved',
  SSMAPI: 'ssm-api',
  FRP: 'frp',
}

type SrvType = typeof SrvTypes[keyof typeof SrvTypes]

interface SrvBasics extends Listen {
  id: number
  type: SrvType
  tag: string
  tls_id: number
}

export interface DERP extends SrvBasics {
  tls: iTls
  config_path: string
  verify_client_endpoint?: string[]
  verify_client_url?: any[]
  home?: string
  mesh_with?: any[]
  mesh_psk?: string
  mesh_psk_file?: string
  stun?: any
}

export interface Resolved extends SrvBasics {}

export interface SSMAPI extends SrvBasics {
  servers: any
  tls?: iTls
}

// FRP Proxy configuration
export interface FRPProxy {
  name: string
  type: 'tcp' | 'udp' | 'http' | 'https' | 'stcp' | 'xtcp'
  local_ip: string
  local_port: number
  remote_port?: number
  custom_domains?: string
  subdomain?: string
  locations?: string
  host_header_rewrite?: string
  use_encryption?: boolean
  use_compression?: boolean
  bandwidth_limit?: string
  sk?: string
  health_check_type?: string
  health_check_url?: string
  health_check_interval?: number
  health_check_max_failed?: number
}

// FRP base configuration (without Listen)
interface FRPBasics {
  id: number
  type: 'frp'
  tag: string
  tls_id: number
}

// FRP Server configuration
export interface FRPServer extends FRPBasics {
  mode: 'server'
  bind_port: number
  vhost_http_port?: number
  vhost_https_port?: number
  dashboard_addr?: string
  dashboard_port?: number
  dashboard_user?: string
  dashboard_pwd?: string
  token?: string
  tcp_mux?: boolean
  max_pool_count?: number
  heartbeat_timeout?: number
  user_conn_timeout?: number
  enable_prometheus?: boolean
  metrics_addr?: string
  metrics_port?: number
}

// FRP Client configuration
export interface FRPClient extends FRPBasics {
  mode: 'client'
  server_addr: string
  server_port: number
  token?: string
  user?: string
  dns_server?: string
  tcp_mux?: boolean
  pool_count?: number
  heartbeat_interval?: number
  heartbeat_timeout?: number
  login_fail_exit?: boolean
  proxies: FRPProxy[]
}

// FRP union type (Server or Client)
export type FRP = FRPServer | FRPClient

type InterfaceMap = {
  derp: DERP
  resolved: Resolved
  'ssm-api': SSMAPI
  frp: FRP
}

export type Srv = InterfaceMap[keyof InterfaceMap]

const defaultValues: any = {
  derp: { type: 'derp', config_path: '', tls_id:0 },
  resolved: { type: 'resolved', listen: '::', listen_port: 53 },
  'ssm-api': { type: 'ssm-api', tls_id: 0, servers: {} },
  frp: {
    type: 'frp',
    mode: 'server',
    bind_port: 7000,
    tls_id: 0,
    proxies: []
  },
}

export function createSrv<T extends Srv>(type: string, json?: Partial<T>): Srv {
  const defaultObject: Srv = { ...defaultValues[type], ...(json || {}) }
  return defaultObject as any
}