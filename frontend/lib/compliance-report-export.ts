import jsPDF from "jspdf";
import * as XLSX from "xlsx";

type AnyRecord = Record<string, unknown>;

interface ComplianceReportSnapshot {
  id: string;
  tenant_id: string;
  branch_id?: string | null;
  generated_by?: string | null;
  report_type: string;
  period_start: string;
  period_end: string;
  generated_at: string;
  retention_until: string;
  snapshot: AnyRecord;
  snapshot_hash?: string | null;
  created_at: string;
}

interface PDFSection {
  title: string;
  records: AnyRecord[];
  fields: string[];
}

function asRecord(value: unknown): AnyRecord {
  if (
    typeof value === "object" &&
    value !== null &&
    !Array.isArray(value)
  ) {
    return value as AnyRecord;
  }

  return {};
}

function asArray(value: unknown): AnyRecord[] {
  if (!Array.isArray(value)) {
    return [];
  }

  return value.map((item) => asRecord(item));
}

function stringValue(value: unknown): string {
  if (value === null || value === undefined) {
    return "";
  }

  if (typeof value === "object") {
    return JSON.stringify(value);
  }

  return String(value);
}

function formatDate(value: unknown): string {
  if (!value) {
    return "";
  }

  const date = new Date(String(value));

  if (Number.isNaN(date.getTime())) {
    return String(value);
  }

  return date.toLocaleString("en-KE");
}

function formatDateOnly(value: unknown): string {
  if (!value) {
    return "";
  }

  const date = new Date(String(value));

  if (Number.isNaN(date.getTime())) {
    return String(value);
  }

  return date.toLocaleDateString("en-KE");
}

function getRecords(
  report: ComplianceReportSnapshot
) {
  const root = asRecord(report.snapshot);

  const records = asRecord(
    root.records
  );

  return {
    branch: asRecord(root.branch),

    period: asRecord(root.period),

    customers: asArray(
      records.customers ??
        root.customers
    ),

    terminals: asArray(
      records.terminals ??
        root.terminals
    ),

    sessions: asArray(
      records.sessions ??
        root.sessions
    ),

    sales: asArray(
      records.sales ??
        root.sales
    ),

    saleItems: asArray(
      records.sale_items ??
        root.sale_items
    ),

    payments: asArray(
      records.payments ??
        root.payments
    ),

    receipts: asArray(
      records.receipts ??
        root.receipts
    ),

    receiptReprints: asArray(
      records.receipt_reprints ??
        root.receipt_reprints
    ),

    expenses: asArray(
      records.expenses ??
        root.expenses
    ),

    auditLogs: asArray(
      records.audit_events ??
        records.audit_logs ??
        root.audit_events ??
        root.audit_logs
    ),

    limitations: Array.isArray(
      root.limitations
    )
      ? root.limitations.map(String)
      : [],
  };
}

function addSectionTitle(
  doc: jsPDF,
  title: string,
  y: number
): number {
  if (y > 270) {
    doc.addPage();
    y = 20;
  }

  doc.setFontSize(12);
  doc.setFont("helvetica", "bold");
  doc.text(title, 14, y);

  return y + 8;
}

function addTable(
  doc: jsPDF,
  section: PDFSection,
  startY: number
): number {
  let y = addSectionTitle(
    doc,
    section.title,
    startY
  );

  if (section.records.length === 0) {
    doc.setFont("helvetica", "normal");
    doc.setFontSize(9);
    doc.text(
      "No records found.",
      14,
      y
    );

    return y + 10;
  }

  const headers =
    section.fields;

  /*
   * Landscape A4:
   * page width = 297mm
   *
   * Margins:
   * left = 14
   * right = 14
   *
   * usable width = 269
   */
  const tableWidth = 269;

  const columnWidth =
    tableWidth /
    Math.max(headers.length, 1);

  const rowHeight = 7;

  const drawHeader = () => {
    doc.setFillColor(
      230,
      230,
      230
    );

    doc.rect(
      14,
      y - 5,
      tableWidth,
      rowHeight,
      "F"
    );

    doc.setFont(
      "helvetica",
      "bold"
    );

    doc.setFontSize(6.5);

    headers.forEach(
      (header, index) => {
        const x =
          14 +
          index * columnWidth;

        doc.text(
          header.substring(
            0,
            24
          ),
          x + 1,
          y
        );
      }
    );

    y += rowHeight;
  };

  drawHeader();

  doc.setFont(
    "helvetica",
    "normal"
  );

  doc.setFontSize(6);

  for (const record of section.records) {
    if (y > 285) {
      doc.addPage();

      y = 20;

      drawHeader();

      doc.setFont(
        "helvetica",
        "normal"
      );

      doc.setFontSize(6);
    }

    let maxLines = 1;

    headers.forEach(
      (header, index) => {
        const x =
          14 +
          index * columnWidth;

        const value =
          stringValue(
            record[header]
          );

        const wrapped =
          doc.splitTextToSize(
            value,
            Math.max(
              columnWidth - 2,
              10
            )
          );

        const lines =
          wrapped.slice(0, 2);

        maxLines = Math.max(
          maxLines,
          lines.length
        );

        doc.text(
          lines,
          x + 1,
          y
        );
      }
    );

    y +=
      Math.max(
        rowHeight,
        maxLines * 4
      );

    /*
     * Light row separator.
     */
    doc.setDrawColor(
      220,
      220,
      220
    );

    doc.line(
      14,
      y - 2,
      14 + tableWidth,
      y - 2
    );
  }

  return y + 6;
}

function createPDFSections(
  data: ReturnType<
    typeof getRecords
  >
): PDFSection[] {
  return [
    {
      title: "Customers",
      records: data.customers,
      fields: [
        "id",
        "full_name",
        "identification_number",
        "phone",
        "customer_type",
        "created_at",
      ],
    },

    {
      title: "Terminals",
      records: data.terminals,
      fields: [
        "id",
        "name",
        "device_identifier",
        "status",
      ],
    },

    {
      title: "Sessions",
      records: data.sessions,
      fields: [
        "id",
        "customer_name",
        "customer_identification_number",
        "terminal_name",
        "attendant_name",
        "session_type",
        "status",
        "started_at",
        "ended_at",
        "final_amount",
      ],
    },

    {
      title: "Sales",
      records: data.sales,
      fields: [
        "id",
        "customer_name",
        "attendant_name",
        "total_amount",
        "status",
        "created_at",
      ],
    },

    {
      title: "Sale Items",
      records: data.saleItems,
      fields: [
        "id",
        "sale_id",
        "description",
        "quantity",
        "unit_price",
        "total_price",
      ],
    },

    {
      title: "Payments",
      records: data.payments,
      fields: [
        "id",
        "sale_id",
        "method",
        "amount",
        "status",
        "paid_at",
      ],
    },

    {
      title: "Receipts",
      records: data.receipts,
      fields: [
        "id",
        "sale_id",
        "receipt_number",
        "receipt_type",
        "issued_at",
        "printed_at",
        "reprint_count",
      ],
    },

    {
      title: "Receipt Reprints",
      records:
        data.receiptReprints,
      fields: [
        "id",
        "receipt_id",
        "user_id",
        "terminal_id",
        "reason",
        "created_at",
      ],
    },

    {
      title: "Expenses",
      records: data.expenses,
      fields: [
        "id",
        "category",
        "description",
        "amount",
        "currency",
        "payment_method",
        "expense_date",
        "status",
      ],
    },

    {
      title: "Audit Logs",
      records: data.auditLogs,
      fields: [
        "id",
        "user_id",
        "action",
        "entity_type",
        "entity_id",
        "created_at",
      ],
    },
  ];
}

export function downloadCompliancePDF(
  report: ComplianceReportSnapshot
): void {
  const doc = new jsPDF({
    orientation: "landscape",
    unit: "mm",
    format: "a4",
  });

  const data =
    getRecords(report);

  let y = 18;

  /*
   * Report title
   */
  doc.setFont(
    "helvetica",
    "bold"
  );

  doc.setFontSize(18);

  doc.text(
    "CAK Compliance Report",
    14,
    y
  );

  y += 9;

  /*
   * Report metadata
   */
  doc.setFont(
    "helvetica",
    "normal"
  );

  doc.setFontSize(9);

  doc.text(
    `Report ID: ${report.id}`,
    14,
    y
  );

  y += 5;

  doc.text(
    `Generated: ${formatDate(
      report.generated_at
    )}`,
    14,
    y
  );

  y += 5;

  doc.text(
    `Period: ${formatDateOnly(
      report.period_start
    )} - ${formatDateOnly(
      report.period_end
    )}`,
    14,
    y
  );

  y += 8;

  /*
   * Branch information
   */
  doc.setFont(
    "helvetica",
    "bold"
  );

  doc.setFontSize(11);

  doc.text(
    `Branch: ${stringValue(
      data.branch.name
    )}`,
    14,
    y
  );

  y += 6;

  doc.setFont(
    "helvetica",
    "normal"
  );

  doc.setFontSize(8);

  doc.text(
    `Address: ${stringValue(
      data.branch.address
    )}`,
    14,
    y
  );

  y += 5;

  doc.text(
    `Phone: ${stringValue(
      data.branch.phone
    )}`,
    14,
    y
  );

  y += 10;

  /*
   * Data sections
   */
  const sections =
    createPDFSections(data);

  for (const section of sections) {
    y = addTable(
      doc,
      section,
      y
    );
  }

  /*
   * Report limitations
   */
  if (
    data.limitations.length >
    0
  ) {
    y = addSectionTitle(
      doc,
      "Report Limitations",
      y
    );

    doc.setFont(
      "helvetica",
      "normal"
    );

    doc.setFontSize(8);

    for (const limitation of data.limitations) {
      if (y > 285) {
        doc.addPage();
        y = 20;
      }

      const lines =
        doc.splitTextToSize(
          `• ${limitation}`,
          260
        );

      doc.text(
        lines,
        14,
        y
      );

      y += Math.max(
        lines.length * 4,
        6
      );
    }
  }

  /*
   * Page numbers
   */
  const pageCount =
    doc.getNumberOfPages();

  for (
    let page = 1;
    page <= pageCount;
    page++
  ) {
    doc.setPage(page);

    doc.setFontSize(7);

    doc.setFont(
      "helvetica",
      "normal"
    );

    doc.text(
      `CAK Compliance Report — Page ${page} of ${pageCount}`,
      14,
      202
    );
  }

  const branchName =
    stringValue(
      data.branch.name
    )
      .replace(
        /[^a-z0-9]+/gi,
        "-"
      )
      .replace(
        /^-+|-+$/g,
        ""
      ) || "branch";

  doc.save(
    `CAK-Compliance-${branchName}-${report.period_start}-${report.period_end}.pdf`
  );
}

export function downloadComplianceExcel(
  report: ComplianceReportSnapshot
): void {
  const data =
    getRecords(report);

  const workbook =
    XLSX.utils.book_new();

  const branch =
    data.branch;

  /*
   * Report information sheet
   */
  const reportInfo = [
    ["CAK Compliance Report"],
    [],
    ["Report ID", report.id],
    [
      "Report Type",
      report.report_type,
    ],
    [
      "Branch ID",
      report.branch_id ?? "",
    ],
    [
      "Branch Name",
      stringValue(
        branch.name
      ),
    ],
    [
      "Branch Address",
      stringValue(
        branch.address
      ),
    ],
    [
      "Branch Phone",
      stringValue(
        branch.phone
      ),
    ],
    [
      "Period Start",
      formatDateOnly(
        report.period_start
      ),
    ],
    [
      "Period End",
      formatDateOnly(
        report.period_end
      ),
    ],
    [
      "Generated At",
      formatDate(
        report.generated_at
      ),
    ],
    [
      "Retention Until",
      formatDate(
        report.retention_until
      ),
    ],
    [
      "Snapshot Hash",
      report.snapshot_hash ??
        "",
    ],
  ];

  XLSX.utils.book_append_sheet(
    workbook,
    XLSX.utils.aoa_to_sheet(
      reportInfo
    ),
    "Report Info"
  );

  /*
   * Add data worksheet.
   */
  const addWorksheet = (
    name: string,
    records: AnyRecord[]
  ) => {
    const rows =
      records.length > 0
        ? records
        : [
            {
              message:
                "No records found.",
            },
          ];

    XLSX.utils.book_append_sheet(
      workbook,
      XLSX.utils.json_to_sheet(
        rows
      ),
      name
    );
  };

  addWorksheet(
    "Customers",
    data.customers
  );

  addWorksheet(
    "Terminals",
    data.terminals
  );

  addWorksheet(
    "Sessions",
    data.sessions
  );

  addWorksheet(
    "Sales",
    data.sales
  );

  addWorksheet(
    "Sale Items",
    data.saleItems
  );

  addWorksheet(
    "Payments",
    data.payments
  );

  addWorksheet(
    "Receipts",
    data.receipts
  );

  addWorksheet(
    "Receipt Reprints",
    data.receiptReprints
  );

  addWorksheet(
    "Expenses",
    data.expenses
  );

  addWorksheet(
    "Audit Logs",
    data.auditLogs
  );

  /*
   * Limitations worksheet
   */
  const limitations =
    data.limitations.length > 0
      ? data.limitations.map(
          (item) => ({
            limitation: item,
          })
        )
      : [
          {
            limitation:
              "None recorded.",
          },
        ];

  XLSX.utils.book_append_sheet(
    workbook,
    XLSX.utils.json_to_sheet(
      limitations
    ),
    "Limitations"
  );

  const branchName =
    stringValue(
      branch.name
    )
      .replace(
        /[^a-z0-9]+/gi,
        "-"
      )
      .replace(
        /^-+|-+$/g,
        ""
      ) || "branch";

  XLSX.writeFile(
    workbook,
    `CAK-Compliance-${branchName}-${report.period_start}-${report.period_end}.xlsx`
  );
}