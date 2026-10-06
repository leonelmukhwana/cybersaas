
import {
  FetchOwnerReportFilters,
  OwnerReportSummaryResponse,
} from "@/types/reports";

const API_BASE_URL =
  process.env.NEXT_PUBLIC_API_URL || "http://localhost:8080/api";

export const reportService = {
  async getOwnerSummary(
    token: string,
    filters: FetchOwnerReportFilters = {}
  ): Promise<OwnerReportSummaryResponse> {
    const params = new URLSearchParams();

    if (filters.branch_id) {
      params.append("branch_id", filters.branch_id);
    }

    if (filters.period_start) {
      params.append("period_start", filters.period_start);
    }

    if (filters.period_end) {
      params.append("period_end", filters.period_end);
    }

    const queryString = params.toString();

    const url =
      `${API_BASE_URL}/reports/owner/summary` +
      (queryString ? `?${queryString}` : "");

    const response = await fetch(url, {
      method: "GET",
      headers: {
        "Content-Type": "application/json",
        Authorization: `Bearer ${token}`,
      },
    });

    if (!response.ok) {
      const errorData = await response.json().catch(() => ({}));

      throw new Error(
        errorData.message || "Failed to fetch report summary."
      );
    }

    return response.json();
  },
};
