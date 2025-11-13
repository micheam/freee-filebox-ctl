package freeeapi

import (
	"bytes"
	"fmt"
	"io"
	"mime/multipart"
	"path/filepath"

	"github.com/oapi-codegen/runtime/types"

	"github.com/micheam/freee-filebox-ctl/internal/freeeapi/gen"
)

func NewReceiptCreateParams(companyID int64, filename string, r io.Reader) (*gen.ReceiptCreateParams, error) {
	data, err := io.ReadAll(r)
	if err != nil {
		return nil, fmt.Errorf("read file data: %w", err)
	}
	var receiptFile types.File
	receiptFile.InitFromBytes(data, filepath.Base(filename))

	return &gen.ReceiptCreateParams{
		CompanyId: companyID,
		Receipt:   receiptFile,
	}, nil
}

// EncodeReceiptCreateParams は gen.ReceiptCreateParams を multipart/form-data 形式の
// io.Reader とその Content-Type にエンコードします。
//
// 戻り値:
//   - io.Reader: multipart/form-data のボディ
//   - string: Content-Type ヘッダーの値（boundary を含む）
//   - error: エラーが発生した場合
func EncodeReceiptCreateParams(params *gen.ReceiptCreateParams) (io.Reader, string, error) {
	var body bytes.Buffer
	writer := multipart.NewWriter(&body)

	// 必須フィールド: company_id
	if err := writer.WriteField("company_id", fmt.Sprintf("%d", params.CompanyId)); err != nil {
		return nil, "", fmt.Errorf("write company_id: %w", err)
	}

	// 必須フィールド: receipt (ファイル)
	fileReader, err := params.Receipt.Reader()
	if err != nil {
		return nil, "", fmt.Errorf("get receipt reader: %w", err)
	}
	defer fileReader.Close()

	part, err := writer.CreateFormFile("receipt", params.Receipt.Filename())
	if err != nil {
		return nil, "", fmt.Errorf("create form file: %w", err)
	}

	if _, err := io.Copy(part, fileReader); err != nil {
		return nil, "", fmt.Errorf("copy receipt file: %w", err)
	}

	// オプショナルフィールド: description
	if params.Description != nil {
		if err := writer.WriteField("description", *params.Description); err != nil {
			return nil, "", fmt.Errorf("write description: %w", err)
		}
	}

	// オプショナルフィールド: document_type
	if params.DocumentType != nil {
		if err := writer.WriteField("document_type", string(*params.DocumentType)); err != nil {
			return nil, "", fmt.Errorf("write document_type: %w", err)
		}
	}

	// オプショナルフィールド: qualified_invoice
	if params.QualifiedInvoice != nil {
		if err := writer.WriteField("qualified_invoice", string(*params.QualifiedInvoice)); err != nil {
			return nil, "", fmt.Errorf("write qualified_invoice: %w", err)
		}
	}

	// オプショナルフィールド: receipt_metadatum_amount
	if params.ReceiptMetadatumAmount != nil {
		if err := writer.WriteField("receipt_metadatum_amount", fmt.Sprintf("%d", *params.ReceiptMetadatumAmount)); err != nil {
			return nil, "", fmt.Errorf("write receipt_metadatum_amount: %w", err)
		}
	}

	// オプショナルフィールド: receipt_metadatum_issue_date
	if params.ReceiptMetadatumIssueDate != nil {
		if err := writer.WriteField("receipt_metadatum_issue_date", *params.ReceiptMetadatumIssueDate); err != nil {
			return nil, "", fmt.Errorf("write receipt_metadatum_issue_date: %w", err)
		}
	}

	// オプショナルフィールド: receipt_metadatum_partner_name
	if params.ReceiptMetadatumPartnerName != nil {
		if err := writer.WriteField("receipt_metadatum_partner_name", *params.ReceiptMetadatumPartnerName); err != nil {
			return nil, "", fmt.Errorf("write receipt_metadatum_partner_name: %w", err)
		}
	}

	// マルチパートを完成させる
	if err := writer.Close(); err != nil {
		return nil, "", fmt.Errorf("close writer: %w", err)
	}

	// Content-Type を返す（boundary を含む）
	contentType := writer.FormDataContentType()

	return &body, contentType, nil
}
