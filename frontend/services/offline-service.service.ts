import {
  getLocalService,
  getLocalServices,
  replaceLocalServices,
  type LocalService,
} from "@/lib/offline/db";

import type { Service } from "@/types/service";

class OfflineServiceService {
  /**
   * Convert backend services into local services.
   *
   * The branchId passed to this method is authoritative.
   * This makes sure the cached records are stored under
   * the exact branch used by the current attendant.
   */
  private toLocalServices(
    branchId: string,
    services: Service[],
  ): LocalService[] {
    const cachedAt = new Date().toISOString();

    return services.map(
      (service): LocalService => ({
        id: service.id,
        tenant_id: service.tenant_id,
        branch_id: branchId || service.branch_id,
        name: service.name,
        description: service.description ?? null,
        price: service.price,
        active: service.active,
        created_at: service.created_at,
        updated_at: service.updated_at,
        cached_at: cachedAt,
      }),
    );
  }

  /**
   * Cache services for a branch.
   */
  async cacheServices(
    branchId: string,
    services: Service[],
  ): Promise<void> {
    const localServices = this.toLocalServices(branchId, services);

    console.log("[Offline Services] Caching services", {
      branchId,
      received: services.length,
      saving: localServices.length,
    });

    // Replace cached services in IndexedDB for this branch
    await replaceLocalServices(branchId, localServices);

    if (localServices.length === 0) {
      console.warn(
        "[Offline Services] No services received from backend to cache.",
      );
      return;
    }

    /*
     * Verify that the records can immediately
     * be read back from IndexedDB.
     */
    const cached = await getLocalServices(branchId, false);

    console.log("[Offline Services] Cache verification", {
      branchId,
      cached: cached.length,
      services: cached,
    });

    if (cached.length === 0 && localServices.length > 0) {
      throw new Error(
        "Services were received from the server but could not be cached locally.",
      );
    }
  }

  /**
   * Replace the cached services for a branch.
   */
  async replaceServices(
    branchId: string,
    services: Service[],
  ): Promise<void> {
    await this.cacheServices(branchId, services);
  }

  /**
   * Get cached services for a branch.
   * Includes fallback strategy if branchId mismatch occurs offline.
   */
  async list(
    branchId: string,
    activeOnly = true,
  ): Promise<Service[]> {
    console.log("[Offline Services] Reading cache", {
      branchId,
      activeOnly,
    });

    let localServices = await getLocalServices(branchId, activeOnly);

    // Fallback 1: Try without active flag restriction if initial query returns empty
    if (localServices.length === 0 && activeOnly) {
      console.warn(
        "[Offline Services] No active services found for branchId. Attempting query without active-only filter.",
      );
      localServices = await getLocalServices(branchId, false);
    }

    // Fallback 2: If branchId changed or wasn't properly set offline, attempt empty branch ID query
    if (localServices.length === 0 && branchId) {
      console.warn(
        "[Offline Services] Branch-specific cache empty. Attempting global fallback query.",
      );
      localServices = await getLocalServices("", false);
    }

    console.log("[Offline Services] Cache result", {
      branchId,
      count: localServices.length,
      services: localServices,
    });

    return localServices.map(
      (service): Service => ({
        id: service.id,
        tenant_id: service.tenant_id,
        branch_id: service.branch_id,
        name: service.name,
        description: service.description ?? null,
        price: service.price,
        active: service.active,
        created_at: service.created_at,
        updated_at: service.updated_at,
      }),
    );
  }

  /**
   * Get one cached service.
   */
  async get(serviceId: string): Promise<Service | null> {
    const service = await getLocalService(serviceId);

    if (!service) {
      return null;
    }

    return {
      id: service.id,
      tenant_id: service.tenant_id,
      branch_id: service.branch_id,
      name: service.name,
      description: service.description ?? null,
      price: service.price,
      active: service.active,
      created_at: service.created_at,
      updated_at: service.updated_at,
    };
  }
}

export const offlineServiceService = new OfflineServiceService();