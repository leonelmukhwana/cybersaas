import { api } from "@/lib/api";

import type {
  AttendantReportResponse,
} from "@/types/attendant-report";

const BASE_PATH = "/api/reports/attendant";

export const attendantReportService = {
  async getSummary(
    periodStart: string,
    periodEnd: string
  ): Promise<AttendantReportResponse> {
    const params = new URLSearchParams();

    params.set("period_start", periodStart);
    params.set("period_end", periodEnd);

    return api<AttendantReportResponse>(
      `${BASE_PATH}/summary?${params.toString()}`,
      {
        method: "GET",
      }
    );
  },
};