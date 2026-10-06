
import type {
  Service,
  ServiceListResponse,
  CreateServiceRequest,
  UpdateServiceRequest,
  ChangeServiceStatusRequest,
  ServiceQueryParams,
} from "@/types/service";

const API_URL =
  process.env.NEXT_PUBLIC_API_URL ||
  "http://localhost:8080/api";

class ServiceService {
  private getHeaders(token: string) {
    return {
      "Content-Type": "application/json",
      Authorization: `Bearer ${token}`,
    };
  }

  async list(
    token: string,
    params: ServiceQueryParams,
  ): Promise<ServiceListResponse> {
    const searchParams = new URLSearchParams();

    searchParams.set("branch_id", params.branch_id);
    searchParams.set(
      "active_only",
      String(params.active_only ?? true),
    );
    searchParams.set(
      "limit",
      String(params.limit ?? 50),
    );
    searchParams.set(
      "offset",
      String(params.offset ?? 0),
    );

    const response = await fetch(
      `${API_URL}/services?${searchParams.toString()}`,
      {
        method: "GET",
        headers: this.getHeaders(token),
      },
    );

    if (!response.ok) {
      throw new Error(
        await this.getErrorMessage(response),
      );
    }

    return response.json();
  }

  async get(
    token: string,
    serviceId: string,
    branchId: string,
  ): Promise<Service> {
    const searchParams = new URLSearchParams();

    searchParams.set("branch_id", branchId);

    const response = await fetch(
      `${API_URL}/services/${serviceId}?${searchParams.toString()}`,
      {
        method: "GET",
        headers: this.getHeaders(token),
      },
    );

    if (!response.ok) {
      throw new Error(
        await this.getErrorMessage(response),
      );
    }

    return response.json();
  }

  async create(
    token: string,
    data: CreateServiceRequest,
  ): Promise<Service> {
    const response = await fetch(
      `${API_URL}/services`,
      {
        method: "POST",
        headers: this.getHeaders(token),
        body: JSON.stringify(data),
      },
    );

    if (!response.ok) {
      throw new Error(
        await this.getErrorMessage(response),
      );
    }

    return response.json();
  }

  async update(
    token: string,
    serviceId: string,
    branchId: string,
    data: UpdateServiceRequest,
  ): Promise<Service> {
    const searchParams = new URLSearchParams();

    searchParams.set("branch_id", branchId);

    const response = await fetch(
      `${API_URL}/services/${serviceId}?${searchParams.toString()}`,
      {
        method: "PUT",
        headers: this.getHeaders(token),
        body: JSON.stringify(data),
      },
    );

    if (!response.ok) {
      throw new Error(
        await this.getErrorMessage(response),
      );
    }

    return response.json();
  }

  async changeStatus(
    token: string,
    serviceId: string,
    branchId: string,
    data: ChangeServiceStatusRequest,
  ): Promise<Service> {
    const searchParams = new URLSearchParams();

    searchParams.set("branch_id", branchId);

    const response = await fetch(
      `${API_URL}/services/${serviceId}/status?${searchParams.toString()}`,
      {
        method: "PATCH",
        headers: this.getHeaders(token),
        body: JSON.stringify(data),
      },
    );

    if (!response.ok) {
      throw new Error(
        await this.getErrorMessage(response),
      );
    }

    return response.json();
  }

  async delete(
    token: string,
    serviceId: string,
    branchId: string,
  ): Promise<Service> {
    // Backend has no DELETE endpoint.
    // Deactivation is used as the delete action.
    return this.changeStatus(
      token,
      serviceId,
      branchId,
      { active: false },
    );
  }

  private async getErrorMessage(
    response: Response,
  ): Promise<string> {
    try {
      const data = await response.json();

      if (data?.error) {
        return data.error;
      }

      if (data?.message) {
        return data.message;
      }
    } catch {
      // Ignore invalid JSON response.
    }

    return `Request failed with status ${response.status}`;
  }
}

export const serviceService =
  new ServiceService();
