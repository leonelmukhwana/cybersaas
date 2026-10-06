import { create } from "zustand"

import type {
  Service,
  CreateServiceRequest,
  UpdateServiceRequest,
} from "@/types/service"

import { serviceService } from "@/services/service.service"

interface ServiceStore {
  services: Service[]
  total: number

  loading: boolean
  saving: boolean

  error: string | null

  selectedService: Service | null

  fetchServices: (
    token: string,
    branchId: string,
    activeOnly?: boolean,
  ) => Promise<void>

  createService: (
    token: string,
    data: CreateServiceRequest,
  ) => Promise<Service>

  updateService: (
    token: string,
    serviceId: string,
    branchId: string,
    data: UpdateServiceRequest,
  ) => Promise<Service>

  deleteService: (
    token: string,
    serviceId: string,
    branchId: string,
  ) => Promise<void>

  toggleStatus: (
    token: string,
    serviceId: string,
    branchId: string,
    active: boolean,
  ) => Promise<void>

  selectService: (
    service: Service | null,
  ) => void

  clearError: () => void
}

export const useServiceStore = create<ServiceStore>(
  (set) => ({
    services: [],
    total: 0,

    loading: false,
    saving: false,

    error: null,

    selectedService: null,

    fetchServices: async (
      token,
      branchId,
      activeOnly = false,
    ) => {
      set({
        loading: true,
        error: null,
      })

      try {
        const response =
          await serviceService.list(
            token,
            {
              branch_id: branchId,
              active_only: activeOnly,
              limit: 100,
              offset: 0,
            },
          )

        set({
          services: response.services ?? [],
          total: response.total ?? 0,
          loading: false,
        })
      } catch (error) {
        set({
          loading: false,
          error:
            error instanceof Error
              ? error.message
              : "Failed to load services",
        })

        throw error
      }
    },

    createService: async (
      token,
      data,
    ) => {
      set({
        saving: true,
        error: null,
      })

      try {
        const service =
          await serviceService.create(
            token,
            data,
          )

        set((state) => ({
          services: [
            service,
            ...state.services,
          ],
          total: state.total + 1,
          saving: false,
        }))

        return service
      } catch (error) {
        set({
          saving: false,
          error:
            error instanceof Error
              ? error.message
              : "Failed to create service",
        })

        throw error
      }
    },

    updateService: async (
      token,
      serviceId,
      branchId,
      data,
    ) => {
      set({
        saving: true,
        error: null,
      })

      try {
        const updated =
          await serviceService.update(
            token,
            serviceId,
            branchId,
            data,
          )

        set((state) => ({
          services: state.services.map(
            (service) =>
              service.id === serviceId
                ? updated
                : service,
          ),
          selectedService: null,
          saving: false,
        }))

        return updated
      } catch (error) {
        set({
          saving: false,
          error:
            error instanceof Error
              ? error.message
              : "Failed to update service",
        })

        throw error
      }
    },

    deleteService: async (
      token,
      serviceId,
      branchId,
    ) => {
      set({
        saving: true,
        error: null,
      })

      try {
        const updated =
          await serviceService.delete(
            token,
            serviceId,
            branchId,
          )

        set((state) => ({
          services: state.services.map(
            (service) =>
              service.id === serviceId
                ? updated
                : service,
          ),
          saving: false,
        }))
      } catch (error) {
        set({
          saving: false,
          error:
            error instanceof Error
              ? error.message
              : "Failed to deactivate service",
        })

        throw error
      }
    },

    toggleStatus: async (
      token,
      serviceId,
      branchId,
      active,
    ) => {
      set({
        saving: true,
        error: null,
      })

      try {
        const updated =
          await serviceService.changeStatus(
            token,
            serviceId,
            branchId,
            { active },
          )

        set((state) => ({
          services: state.services.map(
            (service) =>
              service.id === serviceId
                ? updated
                : service,
          ),
          saving: false,
        }))
      } catch (error) {
        set({
          saving: false,
          error:
            error instanceof Error
              ? error.message
              : "Failed to change service status",
        })

        throw error
      }
    },

    selectService: (service) => {
      set({
        selectedService: service,
      })
    },

    clearError: () => {
      set({
        error: null,
      })
    },
  }),
)