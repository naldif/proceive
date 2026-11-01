package purchaseorder

import (
	"bytes"
	"fmt"
	"strings"
	"time"
)

// statusName mengembalikan representasi teks status untuk tampilan PDF.
func statusName(code string) string {
	switch strings.ToUpper(strings.TrimSpace(code)) {
	case StatusDraft:
		return "Draft"
	case StatusSentToVendor:
		return "Terkirim ke Vendor"
	case StatusSJPendingVerification:
		return "Menunggu Verifikasi Surat Jalan"
	case StatusSJRevised:
		return "Perlu Revisi Surat Jalan"
	case StatusReadyForGR:
		return "Siap Diterima Gudang"
	case StatusGRCompleted:
		return "Selesai Diterima (GR Selesai)"
	default:
		if code == "" {
			return "-"
		}
		return code
	}
}

// GeneratePOPDF membuat dokumen Purchase Order dalam format PDF 1.4.
func GeneratePOPDF(po *PurchaseOrder) []byte {
	var body bytes.Buffer

	escape := func(s string) string {
		s = strings.ReplaceAll(s, "\\", "\\\\")
		s = strings.ReplaceAll(s, "(", "\\(")
		s = strings.ReplaceAll(s, ")", "\\)")
		return s
	}

	formatRupiah := func(amount float64) string {
		s := fmt.Sprintf("%.0f", amount)
		n := len(s)
		if n <= 3 {
			return "Rp " + s
		}
		var res []string
		rem := n % 3
		if rem > 0 {
			res = append(res, s[:rem])
		}
		for i := rem; i < n; i += 3 {
			res = append(res, s[i:i+3])
		}
		return "Rp " + strings.Join(res, ".")
	}

	var stream bytes.Buffer

	// drawText mencetak teks pada koordinat (x, y) tertentu.
	drawText := func(x, y float64, font string, size float64, r, g, b float64, text string) {
		stream.WriteString("BT\n")
		stream.WriteString(fmt.Sprintf("%.2f %.2f %.2f rg\n", r, g, b))
		stream.WriteString(fmt.Sprintf("/%s %.1f Tf\n", font, size))
		stream.WriteString(fmt.Sprintf("%.1f %.1f Td\n", x, y))
		stream.WriteString(fmt.Sprintf("(%s) Tj\n", escape(text)))
		stream.WriteString("ET\n")
	}

	// Header dokumen
	stream.WriteString("0.12 0.23 0.38 rg\n")
	stream.WriteString("40 735 515 50 re f\n")

	drawText(55, 758, "F2", 18, 1, 1, 1, "PROCEIVE - PURCHASE ORDER")
	drawText(55, 744, "F1", 9, 0.85, 0.9, 1, "Dokumen Resmi Pesanan Pengadaan Barang (Procurement Order)")

	// Informasi vendor dan pesanan
	drawText(40, 710, "F2", 10, 0.12, 0.23, 0.38, "KEPADA YTH (VENDOR):")

	vendorName := "-"
	vendorEmail := "-"
	vendorAddress := "-"
	vendorContact := "-"
	if po.Vendor != nil {
		if po.Vendor.Name != "" {
			vendorName = po.Vendor.Name
		}
		if po.Vendor.Email != "" {
			vendorEmail = po.Vendor.Email
		}
		if po.Vendor.Address != "" {
			vendorAddress = po.Vendor.Address
		}
		if po.Vendor.ContactPerson != "" {
			vendorContact = po.Vendor.ContactPerson
		}
	}

	drawText(40, 694, "F1", 9, 0.2, 0.2, 0.2, fmt.Sprintf("Nama       : %s", vendorName))
	drawText(40, 680, "F1", 9, 0.2, 0.2, 0.2, fmt.Sprintf("Kontak PIC : %s", vendorContact))
	drawText(40, 666, "F1", 9, 0.2, 0.2, 0.2, fmt.Sprintf("Email      : %s", vendorEmail))
	// Batasi panjang alamat agar rapi
	if len(vendorAddress) > 42 {
		vendorAddress = vendorAddress[:39] + "..."
	}
	drawText(40, 652, "F1", 9, 0.2, 0.2, 0.2, fmt.Sprintf("Alamat     : %s", vendorAddress))

	// Sisi Kanan: Informasi PO
	drawText(330, 710, "F2", 10, 0.12, 0.23, 0.38, "INFORMASI PESANAN:")
	drawText(330, 694, "F1", 9, 0.2, 0.2, 0.2, fmt.Sprintf("Nomor PO   : %s", po.PONumber))

	prNo := "-"
	if po.PR != nil && po.PR.PRNumber != "" {
		prNo = po.PR.PRNumber
	}
	drawText(330, 680, "F1", 9, 0.2, 0.2, 0.2, fmt.Sprintf("Ref. PR    : %s", prNo))

	createdAtStr := po.CreatedAt.Format("02 Jan 2006, 15:04 WIB")
	if po.CreatedAt.IsZero() {
		createdAtStr = time.Now().Format("02 Jan 2006, 15:04 WIB")
	}
	drawText(330, 666, "F1", 9, 0.2, 0.2, 0.2, fmt.Sprintf("Tanggal    : %s", createdAtStr))

	// Status pesanan
	readableStatus := statusName(po.Status)
	drawText(330, 652, "F2", 9, 0.1, 0.5, 0.2, fmt.Sprintf("Status     : %s", readableStatus))

	// Garis pemisah
	stream.WriteString("0.8 0.85 0.9 RG\n1 w\n40 635 m 555 635 l S\n")

	// Header tabel item
	stream.WriteString("0.18 0.28 0.45 rg\n40 605 515 22 re f\n")
	drawText(48, 612, "F2", 9, 1, 1, 1, "NO")
	drawText(75, 612, "F2", 9, 1, 1, 1, "KODE")
	drawText(140, 612, "F2", 9, 1, 1, 1, "NAMA BARANG")
	drawText(310, 612, "F2", 9, 1, 1, 1, "JUMLAH")
	drawText(375, 612, "F2", 9, 1, 1, 1, "HARGA SATUAN")
	drawText(460, 612, "F2", 9, 1, 1, 1, "SUBTOTAL")

	// Daftar item pesanan
	y := 585.0
	for idx, line := range po.Lines {
		itemCode := "-"
		itemName := "-"
		uom := "UNIT"
		if line.Item != nil {
			if line.Item.Code != "" {
				itemCode = line.Item.Code
			}
			if line.Item.Name != "" {
				itemName = line.Item.Name
			}
			if line.Item.UOM != "" {
				uom = line.Item.UOM
			}
		}

		subtotal := float64(line.Qty) * line.UnitPrice

		if idx%2 == 1 {
			stream.WriteString(fmt.Sprintf("0.96 0.97 0.99 rg\n40 %.1f 515 20 re f\n", y-5))
		}

		displayName := itemName
		if len(displayName) > 30 {
			displayName = displayName[:27] + "..."
		}

		drawText(48, y, "F1", 9, 0.2, 0.2, 0.2, fmt.Sprintf("%d", idx+1))
		drawText(75, y, "F1", 9, 0.2, 0.2, 0.2, itemCode)
		drawText(140, y, "F1", 9, 0.1, 0.1, 0.1, displayName)
		drawText(310, y, "F1", 9, 0.2, 0.2, 0.2, fmt.Sprintf("%d %s", line.Qty, uom))
		drawText(375, y, "F1", 9, 0.2, 0.2, 0.2, formatRupiah(line.UnitPrice))
		drawText(460, y, "F2", 9, 0.12, 0.23, 0.38, formatRupiah(subtotal))

		y -= 22
		if y < 160 {
			break
		}
	}

	// Garis penutup tabel
	stream.WriteString(fmt.Sprintf("0.8 0.85 0.9 RG\n1 w\n40 %.1f m 555 %.1f l S\n", y+5, y+5))

	// Total pembelian
	y -= 25
	stream.WriteString(fmt.Sprintf("0.93 0.96 1 rg\n330 %.1f 225 28 re f\n", y-5))
	stream.WriteString(fmt.Sprintf("0.18 0.28 0.45 RG\n1.2 w\n330 %.1f 225 28 re S\n", y-5))

	drawText(340, y+4, "F2", 10, 0.12, 0.23, 0.38, "TOTAL PEMBELIAN:")
	drawText(450, y+4, "F2", 11, 0.12, 0.23, 0.38, formatRupiah(po.TotalAmount))

	// Tanda tangan purchasing dan management
	y -= 55

	drawText(60, y, "F2", 10, 0.12, 0.23, 0.38, "Diterbitkan Oleh (PIC Purchasing):")
	drawText(370, y, "F2", 10, 0.12, 0.23, 0.38, "Disetujui Oleh (Management):")

	// Tanda tangan PIC purchasing
	stream.WriteString("0.12 0.25 0.65 RG\n")
	stream.WriteString("1.6 w\n")
	stream.WriteString(fmt.Sprintf("70 %.1f m\n", y-35))
	stream.WriteString(fmt.Sprintf("78 %.1f 85 %.1f 92 %.1f c\n", y-15, y-10, y-28))
	stream.WriteString(fmt.Sprintf("98 %.1f 82 %.1f 80 %.1f c\n", y-42, y-38, y-22))
	stream.WriteString(fmt.Sprintf("85 %.1f 105 %.1f 115 %.1f c\n", y-15, y-28, y-35))
	stream.WriteString(fmt.Sprintf("122 %.1f 130 %.1f 138 %.1f c\n", y-22, y-32, y-26))
	stream.WriteString(fmt.Sprintf("145 %.1f 152 %.1f 162 %.1f c\n", y-20, y-30, y-24))
	stream.WriteString(fmt.Sprintf("140 %.1f 90 %.1f 75 %.1f c\n", y-45, y-48, y-42))
	stream.WriteString(fmt.Sprintf("100 %.1f 165 %.1f 178 %.1f c\n", y-45, y-42, y-38))
	stream.WriteString("S\n")

	// Stempel verified purchasing
	stream.WriteString(fmt.Sprintf("0.15 0.65 0.35 RG\n0.8 w\n185 %.1f 55 18 re S\n", y-32))
	drawText(188, y-24, "F2", 6.5, 0.15, 0.65, 0.35, "VERIFIED")
	drawText(188, y-30, "F1", 5.5, 0.15, 0.65, 0.35, "DIGITALLY")

	// Tanda tangan management
	stream.WriteString("0.12 0.25 0.65 RG\n")
	stream.WriteString("1.6 w\n")
	stream.WriteString(fmt.Sprintf("380 %.1f m\n", y-35))
	stream.WriteString(fmt.Sprintf("388 %.1f 398 %.1f 405 %.1f c\n", y-12, y-8, y-26))
	stream.WriteString(fmt.Sprintf("412 %.1f 396 %.1f 394 %.1f c\n", y-40, y-35, y-20))
	stream.WriteString(fmt.Sprintf("400 %.1f 420 %.1f 430 %.1f c\n", y-14, y-26, y-32))
	stream.WriteString(fmt.Sprintf("438 %.1f 445 %.1f 452 %.1f c\n", y-22, y-30, y-25))
	stream.WriteString(fmt.Sprintf("458 %.1f 465 %.1f 475 %.1f c\n", y-18, y-28, y-22))
	stream.WriteString(fmt.Sprintf("450 %.1f 405 %.1f 385 %.1f c\n", y-44, y-46, y-40))
	stream.WriteString(fmt.Sprintf("410 %.1f 480 %.1f 495 %.1f c\n", y-42, y-39, y-36))
	stream.WriteString("S\n")

	// Stempel approved management
	stream.WriteString(fmt.Sprintf("0.18 0.35 0.7 RG\n0.8 w\n500 %.1f 50 18 re S\n", y-32))
	drawText(504, y-24, "F2", 6.5, 0.18, 0.35, 0.7, "APPROVED")
	drawText(504, y-30, "F1", 5.5, 0.18, 0.35, 0.7, "OFFICIAL")

	// Garis tanda tangan
	stream.WriteString(fmt.Sprintf("0.5 0.5 0.5 RG\n0.8 w\n60 %.1f m 240 %.1f l S\n", y-50, y-50))
	stream.WriteString(fmt.Sprintf("370 %.1f m 550 %.1f l S\n", y-50, y-50))

	// Data PIC purchasing
	purchName := "Citra Purchasing"
	purchEmail := "purchasing@proceive.com"
	if po.Purchasing != nil {
		if po.Purchasing.Name != "" {
			purchName = po.Purchasing.Name
		}
		if po.Purchasing.Email != "" {
			purchEmail = po.Purchasing.Email
		}
	}
	drawText(60, y-62, "F2", 9, 0.1, 0.1, 0.1, purchName)
	drawText(60, y-73, "F1", 8, 0.4, 0.4, 0.4, fmt.Sprintf("Staff Procurement (%s)", purchEmail))

	// Data approver management
	approverName := "Super Administrator"
	approverRole := "Management & Approver (admin@proceive.com)"
	if po.PR != nil && po.PR.Approver != nil {
		if po.PR.Approver.Name != "" {
			approverName = po.PR.Approver.Name
		}
		if po.PR.Approver.Email != "" {
			approverRole = fmt.Sprintf("Approver (%s)", po.PR.Approver.Email)
		}
	}
	drawText(370, y-62, "F2", 9, 0.1, 0.1, 0.1, approverName)
	drawText(370, y-73, "F1", 8, 0.4, 0.4, 0.4, approverRole)

	// Catatan footer
	drawText(40, 35, "F1", 7.5, 0.5, 0.5, 0.5, "Dokumen Purchase Order ini sah dan diterbitkan secara digital oleh Proceive Procurement.")
	drawText(40, 25, "F1", 7.5, 0.5, 0.5, 0.5, fmt.Sprintf("ID Dokumen: %s | Dicetak otomatis pada: %s", po.ID.String(), time.Now().Format("02/01/2006 15:04:05")))

	streamBytes := stream.Bytes()

	// Struktur objek PDF 1.4
	offsets := make([]int, 7)

	body.WriteString("%PDF-1.4\n")

	// Obj 1: Catalog
	offsets[1] = body.Len()
	body.WriteString("1 0 obj\n<< /Type /Catalog /Pages 2 0 R >>\nendobj\n")

	// Obj 2: Pages
	offsets[2] = body.Len()
	body.WriteString("2 0 obj\n<< /Type /Pages /Kids [3 0 R] /Count 1 >>\nendobj\n")

	// Obj 3: Page
	offsets[3] = body.Len()
	body.WriteString("3 0 obj\n<< /Type /Page /Parent 2 0 R /MediaBox [0 0 595 842] /Resources << /Font << /F1 4 0 R /F2 5 0 R >> >> /Contents 6 0 R >>\nendobj\n")

	// Obj 4: Font Regular (Helvetica)
	offsets[4] = body.Len()
	body.WriteString("4 0 obj\n<< /Type /Font /Subtype /Type1 /BaseFont /Helvetica >>\nendobj\n")

	// Obj 5: Font Bold (Helvetica-Bold)
	offsets[5] = body.Len()
	body.WriteString("5 0 obj\n<< /Type /Font /Subtype /Type1 /BaseFont /Helvetica-Bold >>\nendobj\n")

	// Obj 6: Stream Contents
	offsets[6] = body.Len()
	body.WriteString(fmt.Sprintf("6 0 obj\n<< /Length %d >>\nstream\n", len(streamBytes)))
	body.Write(streamBytes)
	body.WriteString("\nendstream\nendobj\n")

	// Cross-Reference Table
	xrefOffset := body.Len()
	body.WriteString("xref\n0 7\n")
	body.WriteString("0000000000 65535 f \n")
	for i := 1; i <= 6; i++ {
		body.WriteString(fmt.Sprintf("%010d 00000 n \n", offsets[i]))
	}

	// Trailer
	body.WriteString("trailer\n<< /Size 7 /Root 1 0 R >>\n")
	body.WriteString("startxref\n")
	body.WriteString(fmt.Sprintf("%d\n%%%%EOF\n", xrefOffset))

	return body.Bytes()
}
