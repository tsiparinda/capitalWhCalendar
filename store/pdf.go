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
		//pdfFile := "order_" + o.OperID + ".pdf"

		// Генерируем PDF
		pdfBytes, err := generatePDFInMemory(o)
		if err != nil {
			logger.Log.Errorf("CreatePDF: Error generating PDF for order %s: %v", o.OperID, err)
			return err
		}
		logger.Log.WithFields(logrus.Fields{
			"OrderID":  o.OperID,
			"Articles": o.Articles,
		}).Trace("CreatePDF: PDF generated")

		fileID, err := uploadPDFBytesToDrive(pdfBytes, "order_"+o.OperID+".pdf", folderId, driveSrv)
		if err != nil {
			logger.Log.Errorf("CreatePDF: Upload failed for %s: %v", fileID, err)
			continue
		}

		o.FileURL = "https://drive.google.com/file/d/" + fileID

		//logger.Log.Infof("CreatePDF: Uploaded %s to Drive with file ID: %s", pdfFile, driveFile.Id)
	}

	return nil
}

func generatePDFInMemory(order *Order) ([]byte, error) {
	pdf := gofpdf.New("P", "mm", "A4", "")

	// Создаем временный TTF-файл для шрифта
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

	// Заголовок
	pdf.CellFormat(0, 10, fmt.Sprintf("Заказ #%s", order.OperID), "", 1, "", false, 0, "")

	// Поставщик
	pdf.SetFont("DejaVu", "", 12)
	pdf.CellFormat(0, 10, "Поставщик: "+order.Summary, "", 1, "", false, 0, "")

	// Таблица
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

	// Вывод в bytes.Buffer
	var buf bytes.Buffer
	err = pdf.Output(&buf)
	if err != nil {
		return nil, err
	}

	return buf.Bytes(), nil
}

func generatePDF(order *Order, filename string) error {
	pdf := gofpdf.New("P", "mm", "A4", "")

	// Временный файл для шрифта
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

	// Подключаем шрифт
	pdf.AddUTF8Font("DejaVu", "", tmpFile.Name())
	pdf.AddPage()
	pdf.SetFont("DejaVu", "", 16)

	// Заголовок
	pdf.CellFormat(0, 10, fmt.Sprintf("Заказ #%s", order.OperID), "", 1, "", false, 0, "")

	// Поставщик
	pdf.SetFont("DejaVu", "", 12)
	pdf.CellFormat(0, 10, "Поставщик: "+order.Summary, "", 1, "", false, 0, "")

	// Товары
	pdf.CellFormat(0, 10, "Товары:", "", 1, "", false, 0, "")
	// Заголовок таблицы

	pdf.SetFont("DejaVu", "", 12)
	pdf.CellFormat(10, 8, "№", "1", 0, "C", false, 0, "")
	pdf.CellFormat(25, 8, "Amount", "1", 0, "C", false, 0, "")
	pdf.CellFormat(0, 8, "Article", "1", 1, "L", false, 0, "") // ширина 0 = до конца страницы

	// Данные
	pdf.SetFont("DejaVu", "", 12)
	for i, item := range order.Articles {
		num := fmt.Sprintf("%d", i+1)
		amount := item.Amount.String()
		article := item.Article

		pdf.CellFormat(10, 8, num, "1", 0, "C", false, 0, "")
		pdf.CellFormat(25, 8, amount, "1", 0, "C", false, 0, "")
		pdf.CellFormat(0, 8, article, "1", 1, "L", false, 0, "")
	}

	// for _, item := range order.Articles {
	// 	line := fmt.Sprintf("- %s %s", item.Article, item.Amount.String())
	// 	pdf.CellFormat(0, 8, line, "", 1, "", false, 0, "")
	// }

	return pdf.OutputFileAndClose(filename)
}

// ---------- Загрузка в Google Drive ----------
func uploadPDFBytesToDrive(pdfBytes []byte, filename string, folderId string, srv *drive.Service) (string, error) {
	/////////////////////////////// query token
	// Читаем OAuth credentials
	// b, err := os.ReadFile("client_secret.json")
	// if err != nil {
	// 	return "", fmt.Errorf("unable to read client_secret.json: %v", err)
	// }

	// // Настраиваем конфиг
	// config, err := google.ConfigFromJSON(b, drive.DriveFileScope)
	// if err != nil {
	// 	return "", fmt.Errorf("unable to parse client secret file: %v", err)
	// }

	// client := getClient(config)

	// // Создаем сервис Drive
	// srv, err := drive.NewService(context.Background(), option.WithHTTPClient(client))
	// if err != nil {
	// 	return "", fmt.Errorf("unable to create Drive service: %v", err)
	// }

	// Используем bytes.Reader вместо os.File
	reader := bytes.NewReader(pdfBytes)

	file := &drive.File{
		Name:     filename,
		MimeType: "application/pdf",
		Parents:  []string{folderId}, // <<< ID папки
	}

	// Загружаем
	uploadedFile, err := srv.Files.Create(file).Media(reader).Do()
	if err != nil {
		return "", fmt.Errorf("unable to upload to Drive: %v", err)
	}

	// Делаем файл публичным
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

// func uploadToDrive(filename string) (string, error) {

// 	/////////////////////////////// query token
// 	// Читаем OAuth credentials
// 	b, err := os.ReadFile("client_secret.json")
// 	if err != nil {
// 		panic("Unable to read credentials.json: " + err.Error())
// 	}

// 	// Настраиваем конфиг
// 	config, err := google.ConfigFromJSON(b, drive.DriveFileScope)
// 	if err != nil {
// 		panic("Unable to parse client secret file: " + err.Error())
// 	}

// 	client := getClient(config)

// 	// Создаем сервис Drive с новым способом
// 	srv, err := drive.NewService(context.Background(), option.WithHTTPClient(client))
// 	if err != nil {
// 		panic(err)
// 	}

// 	// Файл для загрузки
// 	f, err := os.Open(filename)
// 	if err != nil {
// 		panic(err)
// 	}
// 	defer f.Close()

// 	file := &drive.File{
// 		Name:     filename,
// 		MimeType: "application/pdf",
// 		Parents:  []string{"1N1b8WIfccnb7SdWt8YXr8m-sbyH4t8wT"}, // <<< сюда ID папки
// 	}

// 	// Загружаем
// 	uploadedFile, err := srv.Files.Create(file).Media(f).Do()
// 	if err != nil {
// 		panic(err)
// 	}

// 	// Делаем файл публичным
// 	_, err = srv.Permissions.Create(uploadedFile.Id, &drive.Permission{
// 		Role: "reader",
// 		Type: "anyone",
// 	}).Do()
// 	if err != nil {
// 		return "", fmt.Errorf("unable to set permission: %v", err)
// 	}

// 	fmt.Printf("Uploaded file '%s' with ID: %s\n", uploadedFile.Name, uploadedFile.Id)
// 	return uploadedFile.Id, nil
// }

// Функция конвертации cp1251 -> utf8
func cp1251ToUtf8(s string) string {
	reader := transform.NewReader(strings.NewReader(s), charmap.Windows1251.NewDecoder())
	decoded, err := io.ReadAll(reader)
	if err != nil {
		return s // если ошибка - возвращаем как есть
	}
	return string(decoded)
}
