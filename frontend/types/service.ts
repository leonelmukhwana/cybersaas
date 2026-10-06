export interface Service {
  id: string
  tenant_id: string
  branch_id: string
  name: string
  description?: string | null
  price: string
  active: boolean
  created_at: string
  updated_at: string
}

export interface CreateServiceRequest {
  branch_id: string
  name: string
  description?: string | null
  price: string
}

export interface UpdateServiceRequest {
  name: string
  description?: string | null
  price: string
}

export interface ChangeServiceStatusRequest {
  active: boolean
}

export interface ServiceListResponse {
  services: Service[]
  total: number
}

export interface ServiceQueryParams {
  branch_id: string
  active_only?: boolean
  limit?: number
  offset?: number
}