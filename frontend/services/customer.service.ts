import type {
  Customer,
  CustomerListResponse,
  CreateCustomerRequest,
  UpdateCustomerRequest,
  CustomerQueryParams,
} from "@/types/customer";

const API_URL =
  process.env.NEXT_PUBLIC_API_URL ||
  "http://localhost:8080/api";

class CustomerService {
  private getHeaders(
    token: string,
    clientOperationId?: string,
  ) {
    return {
      "Content-Type": "application/json",
      Authorization: `Bearer ${token}`,
      ...(clientOperationId
        ? {
            "X-Client-Operation-ID":
              clientOperationId,
          }
        : {}),
    };
  }

  async list(
    token: string,
    params: CustomerQueryParams,
  ): Promise<CustomerListResponse> {
    const searchParams = new URLSearchParams();

    searchParams.set(
      "branch_id",
      params.branch_id,
    );

    if (params.search) {
      searchParams.set(
        "search",
        params.search,
      );
    }

    searchParams.set(
      "limit",
      String(params.limit ?? 50),
    );

    searchParams.set(
      "offset",
      String(params.offset ?? 0),
    );

    const response = await fetch(
      `${API_URL}/customers?${searchParams.toString()}`,
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
    customerId: string,
    branchId: string,
  ): Promise<Customer> {
    const searchParams = new URLSearchParams();

    searchParams.set(
      "branch_id",
      branchId,
    );

    const response = await fetch(
      `${API_URL}/customers/${customerId}?${searchParams.toString()}`,
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

    const data = await response.json();

    return data.customer;
  }

  async create(
    token: string,
    data: CreateCustomerRequest,
  ): Promise<Customer> {
    const clientOperationId =
      crypto.randomUUID();

    const response = await fetch(
      `${API_URL}/customers`,
      {
        method: "POST",
        headers: this.getHeaders(
          token,
          clientOperationId,
        ),
        body: JSON.stringify(data),
      },
    );

    if (!response.ok) {
      throw new Error(
        await this.getErrorMessage(response),
      );
    }

    const result = await response.json();

    return result.customer;
  }

  async update(
    token: string,
    customerId: string,
    branchId: string,
    data: UpdateCustomerRequest,
  ): Promise<Customer> {
    const searchParams = new URLSearchParams();

    searchParams.set(
      "branch_id",
      branchId,
    );

    const clientOperationId =
      crypto.randomUUID();

    const response = await fetch(
      `${API_URL}/customers/${customerId}?${searchParams.toString()}`,
      {
        method: "PUT",
        headers: this.getHeaders(
          token,
          clientOperationId,
        ),
        body: JSON.stringify(data),
      },
    );

    if (!response.ok) {
      throw new Error(
        await this.getErrorMessage(response),
      );
    }

    const result = await response.json();

    return result.customer;
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
      // Ignore invalid JSON.
    }

    return `Request failed with status ${response.status}`;
  }
}

export const customerService =
  new CustomerService();