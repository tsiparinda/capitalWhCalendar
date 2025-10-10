package store

import (
	"bytes"
	"capitalWhCalendar/logger"
	_ "embed"
	"io"
	"os"
	"strings"

	"fmt"

	"github.com/jung-kurt/gofpdf"
	"github.com/sirupsen/logrus"
	"golang.org/x/text/encoding/charmap"
	"golang.org/x/text/transform"
	"google.golang.org/api/drive/v3"
)

//go:embed fonts/DejaVuSans.ttf
var fontData []byte

func CreatePDF(driveSrv *drive.Service, orders *[]Order, folderId string) error {

	for i := range *orders {
		o := &(*orders)[i]

		// generate PDF
		pdfBytes, err := generatePDFInMemory(o)
		if err != nil {
			logger.Log.Errorf("CreatePDF: Error generating PDF for order %s: %v", o.OperID, err)
			continue
		}
		logger.Log.WithFields(logrus.Fields{
			"OrderID":  o.OperID,
			"Articles": o.Articles,
		}).Trace("CreatePDF: PDFinMemory was generated")

		fileID, err := uploadPDFBytesToDrive(pdfBytes, "order_"+o.OperID+".pdf", folderId, driveSrv)
		if err != nil {
			logger.Log.Warnf("CreatePDF: Upload failed for %s: %v", fileID, err)
			o.FileURL = ""
			continue
		}

		o.FileURL = "https://drive.google.com/file/d/" + fileID
	}

	return nil
}

func generatePDFInMemory(order *Order) ([]byte, error) {
	pdf := gofpdf.New("P", "mm", "A4", "")

	// create a temporary TTF-file for font
	tmpFile, err := os.CreateTemp("", "embedded-*.ttf")
	if err != nil {
		return nil, err
	}
	defer os.Remove(tmpFile.Name())
	if _, err := tmpFile.Write(fontData); err != nil {
		tmpFile.Close()
		return nil, err
	}
	tmpFile.Close()

	pdf.AddUTF8Font("DejaVu", "", tmpFile.Name())
	pdf.AddPage()
	pdf.SetFont("DejaVu", "", 12)

	// Header
	pdf.CellFormat(0, 10, fmt.Sprintf("Заказ #%s", order.OperID), "", 1, "", false, 0, "")

	// Supplier
	pdf.SetFont("DejaVu", "", 12)
	pdf.CellFormat(0, 10, "Поставщик: "+order.Summary, "", 1, "", false, 0, "")

	// Table
	pdf.SetFont("DejaVu", "", 12)
	pdf.CellFormat(10, 8, "№", "1", 0, "C", false, 0, "")
	pdf.CellFormat(25, 8, "Amount", "1", 0, "C", false, 0, "")
	pdf.CellFormat(0, 8, "Article", "1", 1, "L", false, 0, "")

	pdf.SetFont("DejaVu", "", 12)
	for i, item := range order.Articles {
		pdf.CellFormat(10, 8, fmt.Sprintf("%d", i+1), "1", 0, "C", false, 0, "")
		pdf.CellFormat(25, 8, item.Amount.String(), "1", 0, "C", false, 0, "")
		pdf.CellFormat(0, 8, item.Article, "1", 1, "L", false, 0, "")
	}

	// Output to bytes.Buffer
	var buf bytes.Buffer
	err = pdf.Output(&buf)
	if err != nil {
		return nil, err
	}

	return buf.Bytes(), nil
}

func generatePDF(order *Order, filename string) error {
	pdf := gofpdf.New("P", "mm", "A4", "")

	// create a temporary TTF-file for font
	tmpFile, err := os.CreateTemp("", "embedded-*.ttf")
	if err != nil {
		return fmt.Errorf("cannot create temp font file: %v", err)
	}
	defer os.Remove(tmpFile.Name())

	if _, err := tmpFile.Write(fontData); err != nil {
		tmpFile.Close()
		return fmt.Errorf("cannot write font data: %v", err)
	}
	tmpFile.Close()

	// connect font
	pdf.AddUTF8Font("DejaVu", "", tmpFile.Name())
	pdf.AddPage()
	pdf.SetFont("DejaVu", "", 16)

	// Header
	pdf.CellFormat(0, 10, fmt.Sprintf("Заказ #%s", order.OperID), "", 1, "", false, 0, "")

	// Supplier
	pdf.SetFont("DejaVu", "", 12)
	pdf.CellFormat(0, 10, "Поставщик: "+order.Summary, "", 1, "", false, 0, "")

	// Atricles
	pdf.CellFormat(0, 10, "Товары:", "", 1, "", false, 0, "")

	// Table header
	pdf.SetFont("DejaVu", "", 12)
	pdf.CellFormat(10, 8, "№", "1", 0, "C", false, 0, "")
	pdf.CellFormat(25, 8, "Amount", "1", 0, "C", false, 0, "")
	pdf.CellFormat(0, 8, "Article", "1", 1, "L", false, 0, "") // ширина 0 = до конца страницы

	// Table lines
	pdf.SetFont("DejaVu", "", 12)
	for i, item := range order.Articles {
		num := fmt.Sprintf("%d", i+1)
		amount := item.Amount.String()
		article := item.Article

		pdf.CellFormat(10, 8, num, "1", 0, "C", false, 0, "")
		pdf.CellFormat(25, 8, amount, "1", 0, "C", false, 0, "")
		pdf.CellFormat(0, 8, article, "1", 1, "L", false, 0, "")
	}

	return pdf.OutputFileAndClose(filename)
}

// ---------- Upload data to Google Drive ----------
func uploadPDFBytesToDrive(pdfBytes []byte, filename string, folderId string, srv *drive.Service) (string, error) {

	// Using bytes.Reader instead of os.File
	reader := bytes.NewReader(pdfBytes)

	file := &drive.File{
		Name:     filename,
		MimeType: "application/pdf",
		Parents:  []string{folderId}, // <<< ID папки
	}

	// Upload
	uploadedFile, err := srv.Files.Create(file).Media(reader).Do()
	if err != nil {
		return "", fmt.Errorf("unable to upload to Drive: %v", err)
	}

	// Make a file public
	_, err = srv.Permissions.Create(uploadedFile.Id, &drive.Permission{
		Role: "reader",
		Type: "anyone",
	}).Do()
	if err != nil {
		return "", fmt.Errorf("unable to set permission: %v", err)
	}

	fmt.Printf("Uploaded file '%s' with ID: %s\n", uploadedFile.Name, uploadedFile.Id)
	return uploadedFile.Id, nil
}

// convertation cp1251 -> utf8
func cp1251ToUtf8(s string) string {
	reader := transform.NewReader(strings.NewReader(s), charmap.Windows1251.NewDecoder())
	decoded, err := io.ReadAll(reader)
	if err != nil {
		return s // if error - return as is
	}
	return string(decoded)
}
