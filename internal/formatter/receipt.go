package formatter

import (
	"encoding/json"
	"fmt"
	"io"
	"os"
	"slices"
	"strings"
	"time"

	"github.com/olekukonko/tablewriter"
	"github.com/olekukonko/tablewriter/tw"

	freeeapigen "github.com/micheam/freee-filebox-ctl/freeeapi/gen"
)

// Receipt formats a receipt in a human-readable format similar to `gh pr view`.
type Receipt struct {
	w io.Writer
}

// NewReceipt creates a new Receipt formatter.
func NewReceipt(w io.Writer) *Receipt {
	return &Receipt{w: w}
}

// Format writes the receipt in a human-readable format.
func (f *Receipt) Format(r *freeeapigen.Receipt) error {
	if r == nil {
		return fmt.Errorf("receipt is nil")
	}

	createdAt, err := formatDateTime(r.CreatedAt, "2006-01-02 15:04:05")
	if err != nil {
		return err
	}

	var b strings.Builder

	// Basic information
	fmt.Fprintf(&b, "ID:              %d\n", r.Id)
	fmt.Fprintf(&b, "Status:          %s\n", formatStatus(r.Status))
	fmt.Fprintf(&b, "Created:         %s\n", createdAt)
	fmt.Fprintf(&b, "Origin:          %s\n", r.Origin)
	fmt.Fprintf(&b, "MIME Type:       %s\n", r.MimeType)
	fmt.Fprintf(&b, "Description:     %s\n", formatString(r.Description))

	// Optional fields
	if r.DocumentType != nil {
		fmt.Fprintf(&b, "Document Type:   %s\n", *r.DocumentType)
	}
	if r.InvoiceRegistrationNumber != nil {
		fmt.Fprintf(&b, "Invoice Reg No:  %s\n", *r.InvoiceRegistrationNumber)
	}
	if r.QualifiedInvoice != nil {
		fmt.Fprintf(&b, "Qualified:       %s\n", *r.QualifiedInvoice)
	}

	// Receipt Information section
	if r.ReceiptMetadatum != nil {
		b.WriteString("\nReceipt Information\n")
		if r.ReceiptMetadatum.PartnerName != nil {
			fmt.Fprintf(&b, "  Partner:       %s\n", *r.ReceiptMetadatum.PartnerName)
		}
		if r.ReceiptMetadatum.Amount != nil {
			fmt.Fprintf(&b, "  Amount:        %s\n", formatAmount(*r.ReceiptMetadatum.Amount))
		}
		if r.ReceiptMetadatum.IssueDate != nil {
			fmt.Fprintf(&b, "  Issue Date:    %s\n", *r.ReceiptMetadatum.IssueDate)
		}
	}

	// User section
	b.WriteString("\nUser\n")
	fmt.Fprintf(&b, "  Name:          %s\n", formatString(r.User.DisplayName))
	fmt.Fprintf(&b, "  Email:         %s\n", r.User.Email)
	fmt.Fprintf(&b, "  ID:            %d\n", r.User.Id)

	_, err = f.w.Write([]byte(b.String()))
	return err
}

// formatString formats a string pointer, returning "(none)" for nil values.
func formatString(s *string) string {
	if s == nil || *s == "" {
		return "(none)"
	}
	return *s
}

// formatAmount formats an amount with thousands separator and currency symbol.
func formatAmount(amount int64) string {
	// Format with thousands separator
	sign := ""
	if amount < 0 {
		sign = "-"
		amount = -amount
	}

	str := fmt.Sprintf("%d", amount)
	var result strings.Builder

	for i, c := range str {
		if i > 0 && (len(str)-i)%3 == 0 {
			result.WriteRune(',')
		}
		result.WriteRune(c)
	}

	return fmt.Sprintf("%s¥%s", sign, result.String())
}

// formatStatus formats the status enum to a more readable string.
func formatStatus(status freeeapigen.ReceiptStatus) string {
	return string(status)
}

// formatDateTime formats a datetime string from srcFormat to destFormat in local timezone.
func formatDateTime(v string, destFormat string) (string, error) {
	t, err := time.Parse(time.RFC3339, v)
	if err != nil {
		return "", err
	}
	return t.Local().Format(destFormat), nil
}

// ExtractReceiptFields extracts specified fields from a receipt and returns a map.
// If fields is empty, returns all fields.
func ExtractReceiptFields(receipt *freeeapigen.Receipt, fields []string) map[string]any {
	result := make(map[string]any)

	// If no fields specified, return all fields as a map
	if len(fields) == 0 {
		// Marshal and unmarshal to get a map with all fields
		b, _ := json.Marshal(receipt)
		json.Unmarshal(b, &result)
		return result
	}

	// Extract only specified fields
	for _, field := range fields {
		switch field {
		case "id":
			result["id"] = receipt.Id
		case "status":
			result["status"] = receipt.Status
		case "created_at":
			result["created_at"] = receipt.CreatedAt
		case "description":
			result["description"] = receipt.Description
		case "document_type":
			result["document_type"] = receipt.DocumentType
		case "origin":
			result["origin"] = receipt.Origin
		case "mime_type":
			result["mime_type"] = receipt.MimeType
		case "qualified_invoice":
			result["qualified_invoice"] = receipt.QualifiedInvoice
		case "invoice_registration_number":
			result["invoice_registration_number"] = receipt.InvoiceRegistrationNumber
		case "partner_name", "partner":
			if receipt.ReceiptMetadatum != nil {
				result["partner_name"] = receipt.ReceiptMetadatum.PartnerName
			}
		case "amount":
			if receipt.ReceiptMetadatum != nil {
				result["amount"] = receipt.ReceiptMetadatum.Amount
			}
		case "issue_date":
			if receipt.ReceiptMetadatum != nil {
				result["issue_date"] = receipt.ReceiptMetadatum.IssueDate
			}
		case "receipt_metadatum":
			result["receipt_metadatum"] = receipt.ReceiptMetadatum
		case "user":
			result["user"] = receipt.User
		case "user_id":
			result["user_id"] = receipt.User.Id
		case "user_email":
			result["user_email"] = receipt.User.Email
		case "user_name", "user_display_name":
			result["user_name"] = receipt.User.DisplayName
		}
	}

	return result
}

// ReceiptList formats a list of receipts in a table format.
type ReceiptList struct {
	w io.Writer
}

// NewReceiptList creates a new ReceiptList formatter.
func NewReceiptList(w io.Writer) *ReceiptList {
	return &ReceiptList{w: w}
}

// Format writes the receipts in a table format using tablewriter.
func (f *ReceiptList) Format(receipts []freeeapigen.Receipt) error {
	return f.FormatWithFields(receipts, nil)
}

// FormatWithFields writes the receipts in a table format with only the specified fields.
// If fields is empty, displays all default fields.
func (f *ReceiptList) FormatWithFields(receipts []freeeapigen.Receipt, fields []string) error {
	if len(receipts) == 0 {
		return nil
	}
	selectedFields, err := determineReceiptFields(fields)
	if err != nil {
		return fmt.Errorf("determining receipt fields: %w", err)
	}

	header := make([]any, len(selectedFields))
	headerAlignments := make([]tw.Align, len(selectedFields))
	for i, fd := range selectedFields {
		header[i] = fd.Header
		headerAlignments[i] = fd.Alignment
	}

	rows := make([][]any, 0, len(receipts))
	for _, receipt := range receipts {
		row := make([]any, len(selectedFields))
		for i, fd := range selectedFields {
			val, err := fd.Extractor(&receipt)
			if err != nil {
				return err
			}
			row[i] = val
		}
		rows = append(rows, row)
	}

	table := tablewriter.NewTable(os.Stdout,
		tablewriter.WithConfig(tablewriter.Config{
			Row: tw.CellConfig{Alignment: tw.CellAlignment{PerColumn: headerAlignments}},
		}))
	table.Header(header...)
	table.Bulk(rows)
	return table.Render()
}

// determineReceiptFields determines which fields to display based on requested fields.
func determineReceiptFields(requestedFields []string) ([]fieldDef, error) {
	if len(requestedFields) == 0 {
		requestedFields = slices.Clone(defaultReceiptFieldNames)
	}
	var selectedFields = make([]fieldDef, 0, len(requestedFields))
	for i, fieldName := range requestedFields {
		fd, ok := allFields[fieldName]
		if !ok {
			return nil, &UnsupportedFieldError{idx: uint(i), fieldName: fieldName}
		}
		selectedFields = append(selectedFields, fd)
	}
	return selectedFields, nil
}

// AvailableReceiptFields returns a list of all available receipt fields.
func AvailableReceiptFields() []string {
	return slices.Clone(allReceiptFieldNames)
}

// allReceiptFieldNames defines the available fields for receipts.
// this comes from [freeeapigen.Receipt] and its nested structs
var allReceiptFieldNames = []string{
	"id",
	"created_at",
	"description",
	"document_type",
	"invoice_registration_number",
	"mime_type",
	"origin",
	"qualified_invoice",
	"receipt_metadatum.amount",
	"receipt_metadatum.issue_date",
	"receipt_metadatum.partner_name",
	"status",
	"user.display_name",
	"user.email",
	"user.id",
}

// defaultReceiptFieldNames defines the default fields to display in receipt lists.
var defaultReceiptFieldNames = []string{
	"id",
	"status",
	"created_at",
	"description",
	"receipt_metadatum.partner_name",
	"receipt_metadatum.amount",
	"receipt_metadatum.issue_date",
}

type fieldDef struct {
	Name      string
	Header    string
	Alignment tw.Align
	Extractor func(*freeeapigen.Receipt) (any, error)
}

var allFields = map[string]fieldDef{
	"id": {
		Header:    "ID",
		Alignment: tw.AlignLeft,
		Extractor: func(r *freeeapigen.Receipt) (any, error) {
			return fmt.Sprintf("%d", r.Id), nil
		},
	},
	"status": {
		Header:    "Status",
		Alignment: tw.AlignLeft,
		Extractor: func(r *freeeapigen.Receipt) (any, error) {
			return string(r.Status), nil
		},
	},
	"created_at": {
		Header:    "Created At",
		Alignment: tw.AlignLeft,
		Extractor: func(r *freeeapigen.Receipt) (any, error) {
			return formatDateTime(r.CreatedAt, "2006-01-02 15:04:05")
		},
	},
	"description": {
		Header:    "Description",
		Alignment: tw.AlignLeft,
		Extractor: func(r *freeeapigen.Receipt) (any, error) {
			return formatString(r.Description), nil
		},
	},
	"document_type": {
		Header:    "Document Type",
		Alignment: tw.AlignLeft,
		Extractor: func(r *freeeapigen.Receipt) (any, error) {
			if r.DocumentType == nil {
				return "(none)", nil
			}
			return string(*r.DocumentType), nil
		},
	},
	"invoice_registration_number": {
		Header:    "Invoice Reg No",
		Alignment: tw.AlignLeft,
		Extractor: func(r *freeeapigen.Receipt) (any, error) {
			return formatString(r.InvoiceRegistrationNumber), nil
		},
	},
	"mime_type": {
		Header:    "MIME Type",
		Alignment: tw.AlignLeft,
		Extractor: func(r *freeeapigen.Receipt) (any, error) {
			return r.MimeType, nil
		},
	},
	"origin": {
		Header:    "Origin",
		Alignment: tw.AlignLeft,
		Extractor: func(r *freeeapigen.Receipt) (any, error) {
			return r.Origin, nil
		},
	},
	"qualified_invoice": {
		Header:    "Qualified Invoice",
		Alignment: tw.AlignLeft,
		Extractor: func(r *freeeapigen.Receipt) (any, error) {
			if r.QualifiedInvoice == nil {
				return "(none)", nil
			}
			return string(*r.QualifiedInvoice), nil
		},
	},
	"receipt_metadatum.amount": {
		Header:    "Amount",
		Alignment: tw.AlignRight,
		Extractor: func(r *freeeapigen.Receipt) (any, error) {
			if r.ReceiptMetadatum == nil || r.ReceiptMetadatum.Amount == nil {
				return "(none)", nil
			}
			return formatAmount(*r.ReceiptMetadatum.Amount), nil
		},
	},
	"receipt_metadatum.issue_date": {
		Header:    "Issue Date",
		Alignment: tw.AlignLeft,
		Extractor: func(r *freeeapigen.Receipt) (any, error) {
			if r.ReceiptMetadatum == nil {
				return "(none)", nil
			}
			return formatString(r.ReceiptMetadatum.IssueDate), nil
		},
	},
	"receipt_metadatum.partner_name": {
		Header:    "Partner",
		Alignment: tw.AlignLeft,
		Extractor: func(r *freeeapigen.Receipt) (any, error) {
			if r.ReceiptMetadatum == nil {
				return "(none)", nil
			}
			return formatString(r.ReceiptMetadatum.PartnerName), nil
		},
	},
	"user.display_name": {
		Header:    "User Name",
		Alignment: tw.AlignLeft,
		Extractor: func(r *freeeapigen.Receipt) (any, error) {
			return formatString(r.User.DisplayName), nil
		},
	},
	"user.email": {
		Header:    "User Email",
		Alignment: tw.AlignLeft,
		Extractor: func(r *freeeapigen.Receipt) (any, error) {
			return r.User.Email, nil
		},
	},
	"user.id": {
		Header:    "User ID",
		Alignment: tw.AlignLeft,
		Extractor: func(r *freeeapigen.Receipt) (any, error) {
			return fmt.Sprintf("%d", r.User.Id), nil
		},
	},
}
