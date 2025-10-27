package formatter

import (
	"fmt"
	"io"
	"strings"

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

	var b strings.Builder

	// Basic information
	fmt.Fprintf(&b, "ID:              %d\n", r.Id)
	fmt.Fprintf(&b, "Status:          %s\n", formatStatus(r.Status))
	fmt.Fprintf(&b, "Created:         %s\n", r.CreatedAt)
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

	_, err := f.w.Write([]byte(b.String()))
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
