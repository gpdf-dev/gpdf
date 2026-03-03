# Components

gpdf includes pre-built document components for common business documents: **Invoice**, **Report**, and **Letter**. These create fully-styled, ready-to-generate PDFs from structured data.

## Invoice

Creates a professional invoice with company header, billing details, itemized table, tax calculation, and optional payment information.

### Usage

```go
import "github.com/gpdf-dev/gpdf/template"

doc := template.Invoice(template.InvoiceData{
    Number:  "#INV-2026-001",
    Date:    "March 1, 2026",
    DueDate: "March 31, 2026",
    From: template.InvoiceParty{
        Name:    "ACME Corporation",
        Address: []string{"123 Business Street", "Suite 100", "San Francisco, CA 94105"},
    },
    To: template.InvoiceParty{
        Name:    "John Smith",
        Address: []string{"Tech Solutions Inc.", "456 Client Avenue", "New York, NY 10001"},
    },
    Items: []template.InvoiceItem{
        {Description: "Web Development - Frontend", Quantity: "40 hrs", UnitPrice: 150.00, Amount: 6000.00},
        {Description: "Web Development - Backend",  Quantity: "60 hrs", UnitPrice: 150.00, Amount: 9000.00},
        {Description: "UI/UX Design",               Quantity: "20 hrs", UnitPrice: 120.00, Amount: 2400.00},
        {Description: "QA Testing",                  Quantity: "25 hrs", UnitPrice: 100.00, Amount: 2500.00},
    },
    TaxRate:  10,       // 10% tax
    Currency: "$",      // defaults to "$" if empty
    Notes:    "Thank you for your business!",
    Payment: &template.InvoicePayment{
        BankName: "First National Bank",
        Account:  "1234-5678-9012",
        Routing:  "021000021",
    },
})

data, err := doc.Generate()
```

### InvoiceData Fields

| Field | Type | Required | Description |
|---|---|---|---|
| `Number` | string | yes | Invoice identifier (e.g., "#INV-2026-001") |
| `Date` | string | no | Issue date |
| `DueDate` | string | no | Payment due date |
| `From` | `InvoiceParty` | yes | Sender/seller information |
| `To` | `InvoiceParty` | yes | Recipient/buyer information |
| `Items` | `[]InvoiceItem` | yes | Line items |
| `TaxRate` | float64 | no | Tax percentage (e.g., 10 for 10%). Zero = no tax |
| `Currency` | string | no | Currency symbol (default: `"$"`) |
| `Notes` | string | no | Footer notes text |
| `Payment` | `*InvoicePayment` | no | Payment details (nil = omitted) |

### InvoiceParty

| Field | Type | Description |
|---|---|---|
| `Name` | string | Company or person name |
| `Address` | []string | Address lines |

### InvoiceItem

| Field | Type | Description |
|---|---|---|
| `Description` | string | Item description |
| `Quantity` | string | Quantity string (e.g., "40 hrs", "5") |
| `UnitPrice` | float64 | Price per unit |
| `Amount` | float64 | Line total. If zero, uses `UnitPrice` as fallback |

### InvoicePayment

| Field | Type | Description |
|---|---|---|
| `BankName` | string | Bank name |
| `Account` | string | Account number |
| `Routing` | string | Routing number |

### Invoice Layout

The generated invoice includes:
1. Company header with name and address (left) + "INVOICE" title, number, dates (right)
2. Separator line
3. "Bill To" section (left) + Payment info (right, if provided)
4. Items table with columns: Description, Qty, Unit Price, Amount
5. Subtotal, Tax, and Total
6. Notes footer (if provided)

---

## Report

Creates a multi-section report with a title page, sections containing text and optional data tables/metrics, and automatic headers/footers with page numbers.

### Usage

```go
doc := template.Report(template.ReportData{
    Title:    "Quarterly Report - Q1 2026",
    Subtitle: "Financial Performance Summary",
    Author:   "Finance Department",
    Date:     "April 1, 2026",
    Sections: []template.ReportSection{
        {
            Title:   "Executive Summary",
            Content: "Revenue increased 25% year-over-year, driven by new enterprise offerings.",
        },
        {
            Title:   "Key Metrics",
            Content: "The following metrics highlight our Q1 performance:",
            Metrics: []template.ReportMetric{
                {Label: "Revenue",  Value: "$12.5M", ColorHex: 0x2E7D32},
                {Label: "Users",    Value: "85,000", ColorHex: 0x1565C0},
                {Label: "Growth",   Value: "+25%",   ColorHex: 0xFF6F00},
            },
        },
        {
            Title:   "Sales Breakdown",
            Content: "Regional sales performance for Q1:",
            Table: &template.ReportTable{
                Header:       []string{"Region", "Revenue", "Growth"},
                Rows:         [][]string{
                    {"North America", "$5.2M", "+30%"},
                    {"Europe",        "$3.8M", "+22%"},
                    {"Asia Pacific",  "$3.5M", "+20%"},
                },
                ColumnWidths: []float64{40, 30, 30},
            },
        },
    },
})

data, err := doc.Generate()
```

### ReportData Fields

| Field | Type | Required | Description |
|---|---|---|---|
| `Title` | string | yes | Report title |
| `Subtitle` | string | no | Displayed below title |
| `Author` | string | no | Report author |
| `Date` | string | no | Report date |
| `Sections` | `[]ReportSection` | yes | Content sections |

### ReportSection

| Field | Type | Description |
|---|---|---|
| `Title` | string | Section heading |
| `Content` | string | Paragraph text |
| `Table` | `*ReportTable` | Optional data table |
| `Metrics` | `[]ReportMetric` | Optional metric cards displayed in a grid |

### ReportMetric

| Field | Type | Description |
|---|---|---|
| `Label` | string | Metric label (e.g., "Revenue") |
| `Value` | string | Display value (e.g., "$12.5M") |
| `ColorHex` | uint32 | Hex color for value (e.g., `0x2E7D32`). Zero = default accent color |

### Report Layout

The generated report includes:
- **Header**: Author/title (repeated on every page) with accent line
- **Footer**: "Confidential" label + page number (repeated on every page)
- **Title page**: Large title, subtitle, date, separator
- **Sections**: Each section has heading, paragraph text, optional metrics grid, optional table

---

## Letter

Creates a formal business letter with sender header, recipient address, subject line, body paragraphs, and signature block.

### Usage

```go
doc := template.Letter(template.LetterData{
    From: template.LetterParty{
        Name:    "ACME Corporation",
        Address: []string{"123 Business Street", "San Francisco, CA 94105", "contact@acme.com"},
    },
    To: template.LetterParty{
        Name:    "Jane Doe",
        Address: []string{"Tech Solutions Inc.", "456 Innovation Blvd", "New York, NY 10001"},
    },
    Date:     "March 1, 2026",
    Subject:  "Partnership Proposal",
    Greeting: "Dear Ms. Doe,",
    Body: []string{
        "We are writing to express our interest in establishing a strategic partnership between ACME Corporation and Tech Solutions Inc.",
        "Our companies share a vision for innovation in the technology sector, and we believe a collaboration would be mutually beneficial.",
        "We would welcome the opportunity to discuss this proposal further at your convenience.",
    },
    Closing:    "Sincerely,",
    Signature:  "John Smith",
    SignerTitle: "Chief Executive Officer",
})

data, err := doc.Generate()
```

### LetterData Fields

| Field | Type | Required | Description |
|---|---|---|---|
| `From` | `LetterParty` | yes | Sender information |
| `To` | `LetterParty` | yes | Recipient information |
| `Date` | string | no | Letter date |
| `Subject` | string | no | Subject line (displayed as "Re: ...") |
| `Greeting` | string | no | Opening salutation |
| `Body` | []string | yes | Body paragraphs |
| `Closing` | string | no | Valediction (e.g., "Sincerely,") |
| `Signature` | string | no | Signer's name |
| `SignerTitle` | string | no | Signer's title or role |

### LetterParty

| Field | Type | Description |
|---|---|---|
| `Name` | string | Person or company name |
| `Address` | []string | Address lines |

### Letter Layout

The generated letter includes:
1. Sender header with name and address + separator line
2. Date (right-aligned)
3. Recipient name and address
4. Subject line (if provided)
5. Greeting
6. Body paragraphs
7. Closing + signature block

---

## Customizing Components

All components accept additional `Option` parameters to override defaults:

```go
fontData, _ := os.ReadFile("NotoSansJP-Regular.ttf")

doc := template.Invoice(invoiceData,
    template.WithFont("NotoSansJP", fontData),
    template.WithDefaultFont("NotoSansJP", 10),
    template.WithPageSize(document.Letter),
    template.WithMargins(document.UniformEdges(document.Mm(25))),
)
```

The component builds the document using default settings first, then your options override them.

## Using via the Facade

The top-level `gpdf` package provides convenient aliases:

```go
import gpdf "github.com/gpdf-dev/gpdf"

doc := gpdf.NewInvoice(invoiceData)
doc := gpdf.NewReport(reportData)
doc := gpdf.NewLetter(letterData)
```

## See Also

- [Builder API](02-builder-api.md) -- Build custom documents from scratch
- [Fonts](09-fonts.md) -- Using custom fonts with components
