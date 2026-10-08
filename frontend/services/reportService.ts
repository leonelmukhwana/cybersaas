const API_URL = (
  process.env.NEXT_PUBLIC_API_URL ||
  "http://localhost:8080"
).replace(/\/+$/, "");

const API_BASE_URL = `${API_URL}/api`;

export type ComplianceReportType = "cak_compliance";

export interface GenerateComplianceReportRequest {
  branch_id: string;
  report_type: ComplianceReportType;
  period_start: string;
  period_end: string;
}

export interface ComplianceReportSnapshot {
  id: string;
  tenant_id: string;
  branch_id?: string | null;
  generated_by?: string | null;
  report_type: string;
  period_start: string;
  period_end: string;
  generated_at: string;
  retention_until: string;
  snapshot: Record<string, unknown>;
  snapshot_hash?: string | null;
  created_at: string;
}

export interface ComplianceReportResponse {
  snapshot: ComplianceReportSnapshot;
}

export interface ComplianceReportsResponse {
  reports: ComplianceReportSnapshot[];
}

class ReportService {
  private getHeaders(token: string): HeadersInit {
    return {
      "Content-Type": "application/json",
      Authorization: `Bearer ${token}`,
    };
  }

  private async getErrorMessage(
    response: Response,
    fallback: string,
  ): Promise<string> {
    try {
      const text = await response.text();

      if (!text.trim()) {
        return fallback;
      }

      try {
        const data = JSON.parse(text);

        if (data?.error) {
          return data.error;
        }

        if (data?.message) {
          return data.message;
        }
      } catch {
        return text;
      }
    } catch {
      // Ignore response parsing errors.
    }

    return fallback;
  }

  async generateComplianceReport(
    token: string,
    data: GenerateComplianceReportRequest,
  ): Promise<ComplianceReportResponse> {
    const response = await fetch(
      `${API_BASE_URL}/reports/compliance`,
      {
        method: "POST",
        headers: this.getHeaders(token),
        body: JSON.stringify(data),
      },
    );

    if (!response.ok) {
      throw new Error(
        await this.getErrorMessage(
          response,
          "Failed to generate compliance report",
        ),
      );
    }

    return response.json();
  }

  async listComplianceReports(
    token: string,
    branchId?: string,
  ): Promise<ComplianceReportsResponse> {
    const params = new URLSearchParams();

    if (branchId) {
      params.set("branch_id", branchId);
    }

    const query = params.toString();

    const response = await fetch(
      `${API_BASE_URL}/reports/compliance${
        query ? `?${query}` : ""
      }`,
      {
        method: "GET",
        headers: this.getHeaders(token),
      },
    );

    if (!response.ok) {
      throw new Error(
        await this.getErrorMessage(
          response,
          "Failed to load compliance reports",
        ),
      );
    }

    return response.json();
  }

  async getComplianceReport(
    token: string,
    reportId: string,
  ): Promise<ComplianceReportResponse> {
    const response = await fetch(
      `${API_BASE_URL}/reports/compliance/${reportId}`,
      {
        method: "GET",
        headers: this.getHeaders(token),
      },
    );

    if (!response.ok) {
      throw new Error(
        await this.getErrorMessage(
          response,
          "Failed to load compliance report",
        ),
      );
    }

    return response.json();
  }
}

export const reportService = new ReportService();