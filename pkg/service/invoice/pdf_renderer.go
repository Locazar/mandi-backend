package invoice

import (
	"bytes"
	"context"
	_ "embed"
	"fmt"
	"strconv"
	"strings"

	"github.com/go-pdf/fpdf"
	"github.com/rohit221990/mandi-backend/pkg/domain"
)

// Noto Sans covers the ₹ glyph. fpdf's built-in core fonts are Latin-1 and
// render ₹ as garbage, so an embedded Unicode TTF is mandatory, not cosmetic.
//
//go:embed fonts/NotoSans-Regular.ttf
var notoRegular []byte

//go:embed fonts/NotoSans-Bold.ttf
var notoBold []byte

// Layout C1: white header, logo in its true blue, blue as an accent rule.
// A tax invoice gets printed, scanned and emailed to accountants, and a
// full-bleed colour band degrades at every one of those steps.
const (
	fontFamily = "NotoSans"

	marginLeft  = 15.0
	marginRight = 15.0
	marginTop   = 14.0
	pageWidth   = 210.0 // A4 portrait, mm
	contentW    = pageWidth - marginLeft - marginRight
)

// accentR/G/B is #4E9CD4, sampled from Khangaro-seller/assets/images/logo.png.
var accentR, accentG, accentB = 78, 156, 212

// headingR/G/B is a darkened accent (#2C6E9B) that stays legible as text.
var headingR, headingG, headingB = 44, 110, 155

type pdfRenderer struct{}

func NewPDFRenderer() Renderer { return &pdfRenderer{} }

func (r *pdfRenderer) Render(_ context.Context, inv domain.Invoice, logoPNG []byte) ([]byte, error) {
	pdf := fpdf.New("P", "mm", "A4", "")
	pdf.SetMargins(marginLeft, marginTop, marginRight)
	pdf.SetAutoPageBreak(true, 15)

	pdf.AddUTF8FontFromBytes(fontFamily, "", notoRegular)
	pdf.AddUTF8FontFromBytes(fontFamily, "B", notoBold)
	pdf.AddPage()

	drawHeader(pdf, inv, logoPNG)
	drawMetaRow(pdf, inv)
	drawParties(pdf, inv)
	drawLineItems(pdf, inv)
	drawTotals(pdf, inv)
	drawPaymentBlock(pdf, inv)
	drawFooter(pdf, inv)

	var buf bytes.Buffer
	if err := pdf.Output(&buf); err != nil {
		return nil, fmt.Errorf("render invoice pdf: %w", err)
	}
	return buf.Bytes(), nil
}

func drawHeader(pdf *fpdf.Fpdf, inv domain.Invoice, logoPNG []byte) {
	y := pdf.GetY()

	// Logo, if we have usable bytes. A broken logo must never stop an invoice.
	textX := marginLeft
	if len(logoPNG) > 0 {
		opts := fpdf.ImageOptions{ImageType: "PNG", ReadDpi: false}
		pdf.RegisterImageOptionsReader("logo", opts, bytes.NewReader(logoPNG))
		if pdf.Ok() {
			pdf.ImageOptions("logo", marginLeft, y, 16, 16, false, opts, 0, "")
			textX = marginLeft + 20
		} else {
			// Clear the error so a bad image doesn't poison Output().
			pdf.ClearError()
		}
	}

	pdf.SetXY(textX, y+1)
	pdf.SetFont(fontFamily, "B", 17)
	pdf.SetTextColor(headingR, headingG, headingB)
	pdf.CellFormat(80, 7, "Locazar", "", 2, "L", false, 0, "")

	pdf.SetFont(fontFamily, "", 8)
	pdf.SetTextColor(90, 90, 90)
	pdf.CellFormat(80, 4, inv.SellerLegalName, "", 0, "L", false, 0, "")

	// Right block: TAX INVOICE + number.
	pdf.SetXY(pageWidth-marginRight-70, y+1)
	pdf.SetFont(fontFamily, "B", 8)
	pdf.SetTextColor(headingR, headingG, headingB)
	pdf.CellFormat(70, 5, "TAX INVOICE", "", 2, "R", false, 0, "")
	pdf.SetFont(fontFamily, "", 10)
	pdf.SetTextColor(26, 26, 26)
	pdf.CellFormat(70, 5, inv.InvoiceNumber, "", 0, "R", false, 0, "")

	// Accent rule under the header.
	ruleY := y + 19
	pdf.SetDrawColor(accentR, accentG, accentB)
	pdf.SetLineWidth(0.9)
	pdf.Line(marginLeft, ruleY, pageWidth-marginRight, ruleY)
	pdf.SetY(ruleY + 5)
}

func drawMetaRow(pdf *fpdf.Fpdf, inv domain.Invoice) {
	y := pdf.GetY()
	col := contentW / 3

	label(pdf, marginLeft, y, col, "INVOICE DATE")
	value(pdf, marginLeft, y+4, col, inv.InvoiceDate.Format("02 Jan 2006"), "L")

	label(pdf, marginLeft+col, y, col, "PLACE OF SUPPLY")
	value(pdf, marginLeft+col, y+4, col, orDash(inv.PlaceOfSupply), "L")

	label(pdf, marginLeft+2*col, y, col, "AMOUNT PAID")
	pdf.SetXY(marginLeft+2*col, y+4)
	pdf.SetFont(fontFamily, "B", 11)
	pdf.SetTextColor(26, 26, 26)
	pdf.CellFormat(col, 5, "₹"+formatINR(inv.Total), "", 0, "R", false, 0, "")

	pdf.SetDrawColor(227, 227, 227)
	pdf.SetLineWidth(0.2)
	pdf.Line(marginLeft, y+11, pageWidth-marginRight, y+11)
	pdf.SetY(y + 16)
}

func drawParties(pdf *fpdf.Fpdf, inv domain.Invoice) {
	y := pdf.GetY()
	col := contentW / 2

	// From
	label(pdf, marginLeft, y, col, "FROM")
	pdf.SetXY(marginLeft, y+4)
	pdf.SetFont(fontFamily, "B", 9.5)
	pdf.SetTextColor(26, 26, 26)
	pdf.MultiCell(col-4, 4.4, inv.SellerLegalName, "", "L", false)
	pdf.SetX(marginLeft)
	pdf.SetFont(fontFamily, "", 8.5)
	pdf.SetTextColor(85, 85, 85)
	pdf.MultiCell(col-4, 4, inv.SellerAddress, "", "L", false)
	if inv.SellerGSTIN != "" {
		pdf.SetX(marginLeft)
		pdf.SetTextColor(26, 26, 26)
		pdf.MultiCell(col-4, 4.4, "GSTIN "+inv.SellerGSTIN, "", "L", false)
	}
	fromBottom := pdf.GetY()

	// Billed to
	label(pdf, marginLeft+col, y, col, "BILLED TO")
	pdf.SetXY(marginLeft+col, y+4)
	pdf.SetFont(fontFamily, "B", 9.5)
	pdf.SetTextColor(26, 26, 26)
	pdf.MultiCell(col-4, 4.4, orDash(inv.BuyerName), "", "L", false)

	buyerLines := inv.BuyerContactName
	if inv.BuyerAddress != "" {
		if buyerLines != "" {
			buyerLines += "\n"
		}
		buyerLines += inv.BuyerAddress
	}
	if buyerLines != "" {
		pdf.SetX(marginLeft + col)
		pdf.SetFont(fontFamily, "", 8.5)
		pdf.SetTextColor(85, 85, 85)
		pdf.MultiCell(col-4, 4, buyerLines, "", "L", false)
	}
	// Omitted entirely for sellers who onboarded with PAN/Aadhaar/licence.
	if inv.BuyerGSTIN != "" {
		pdf.SetX(marginLeft + col)
		pdf.SetFont(fontFamily, "", 8.5)
		pdf.SetTextColor(26, 26, 26)
		pdf.MultiCell(col-4, 4.4, "GSTIN "+inv.BuyerGSTIN, "", "L", false)
	}

	if pdf.GetY() < fromBottom {
		pdf.SetY(fromBottom)
	}
	pdf.SetY(pdf.GetY() + 6)
}

func drawLineItems(pdf *fpdf.Fpdf, inv domain.Invoice) {
	const (
		wDesc   = 88.0
		wSAC    = 22.0
		wPeriod = 40.0
	)
	wAmount := contentW - wDesc - wSAC - wPeriod

	// Header row
	pdf.SetFillColor(244, 246, 249)
	pdf.SetDrawColor(227, 227, 227)
	pdf.SetLineWidth(0.2)
	pdf.SetFont(fontFamily, "B", 7.5)
	pdf.SetTextColor(85, 85, 85)
	pdf.CellFormat(wDesc, 7, "DESCRIPTION", "TB", 0, "L", true, 0, "")
	pdf.CellFormat(wSAC, 7, "SAC", "TB", 0, "L", true, 0, "")
	pdf.CellFormat(wPeriod, 7, "PERIOD", "TB", 0, "L", true, 0, "")
	pdf.CellFormat(wAmount, 7, "AMOUNT ₹", "TB", 1, "R", true, 0, "")

	// Body row
	y := pdf.GetY()
	pdf.SetXY(marginLeft, y+2.5)
	pdf.SetFont(fontFamily, "B", 9)
	pdf.SetTextColor(26, 26, 26)
	pdf.MultiCell(wDesc, 4.4, inv.PlanName, "", "L", false)
	if inv.PlanDescription != "" {
		pdf.SetX(marginLeft)
		pdf.SetFont(fontFamily, "", 7.5)
		pdf.SetTextColor(119, 119, 119)
		pdf.MultiCell(wDesc, 3.8, inv.PlanDescription, "", "L", false)
	}
	descBottom := pdf.GetY()

	pdf.SetFont(fontFamily, "", 8.5)
	pdf.SetTextColor(26, 26, 26)
	pdf.SetXY(marginLeft+wDesc, y+2.5)
	pdf.CellFormat(wSAC, 4.4, orDash(inv.SellerSACCode), "", 0, "L", false, 0, "")

	period := ""
	if !inv.PeriodStart.IsZero() && !inv.PeriodEnd.IsZero() {
		period = inv.PeriodStart.Format("02 Jan 2006") + " –\n" + inv.PeriodEnd.Format("02 Jan 2006")
	}
	pdf.SetXY(marginLeft+wDesc+wSAC, y+2.5)
	pdf.MultiCell(wPeriod, 4.4, period, "", "L", false)

	pdf.SetXY(marginLeft+wDesc+wSAC+wPeriod, y+2.5)
	pdf.CellFormat(wAmount, 4.4, formatINR(inv.TaxableValue), "", 0, "R", false, 0, "")

	bottom := descBottom
	if pdf.GetY() > bottom {
		bottom = pdf.GetY()
	}
	bottom += 2.5
	pdf.SetDrawColor(238, 238, 238)
	pdf.Line(marginLeft, bottom, pageWidth-marginRight, bottom)
	pdf.SetY(bottom + 4)
}

func drawTotals(pdf *fpdf.Fpdf, inv domain.Invoice) {
	const w = 70.0
	x := pageWidth - marginRight - w

	totalLine(pdf, x, w, "Taxable value", formatINR(inv.TaxableValue), false)
	gstLabel := "GST"
	if inv.GSTRateBasisPoints > 0 {
		gstLabel = fmt.Sprintf("GST @ %s%%", formatBasisPoints(inv.GSTRateBasisPoints))
	}
	totalLine(pdf, x, w, gstLabel, formatINR(inv.GSTAmount), false)

	// Emphasised total with a rule above it.
	y := pdf.GetY() + 1
	pdf.SetDrawColor(26, 26, 26)
	pdf.SetLineWidth(0.5)
	pdf.Line(x, y, x+w, y)
	pdf.SetY(y + 1.5)
	totalLine(pdf, x, w, "Total", "₹"+formatINR(inv.Total), true)

	pdf.SetY(pdf.GetY() + 6)
}

func totalLine(pdf *fpdf.Fpdf, x, w float64, label, amount string, bold bool) {
	style, size := "", 9.0
	if bold {
		style, size = "B", 11.0
	}
	pdf.SetX(x)
	pdf.SetFont(fontFamily, style, size)
	if bold {
		pdf.SetTextColor(26, 26, 26)
	} else {
		pdf.SetTextColor(85, 85, 85)
	}
	pdf.CellFormat(w*0.55, 5.5, label, "", 0, "L", false, 0, "")
	pdf.SetTextColor(26, 26, 26)
	pdf.CellFormat(w*0.45, 5.5, amount, "", 1, "R", false, 0, "")
}

func drawPaymentBlock(pdf *fpdf.Fpdf, inv domain.Invoice) {
	y := pdf.GetY()
	h := 16.0

	pdf.SetFillColor(244, 246, 249)
	pdf.Rect(marginLeft, y, contentW, h, "F")

	col := contentW / 4
	cells := []struct{ label, value string }{
		{"STATUS", "Paid"},
		{"METHOD", titleCase(orDash(inv.PaymentMethod))},
		{"PAYMENT ID", orDash(inv.RazorpayPaymentID)},
		{"PAID ON", inv.PaidAt.Format("02 Jan 2006")},
	}

	for i, c := range cells {
		x := marginLeft + float64(i)*col + 3
		label(pdf, x, y+3, col-6, c.label)
		pdf.SetXY(x, y+7.5)
		pdf.SetFont(fontFamily, "", 8)
		if i == 0 {
			pdf.SetTextColor(21, 128, 61) // green for Paid
		} else {
			pdf.SetTextColor(26, 26, 26)
		}
		pdf.CellFormat(col-6, 4.5, c.value, "", 0, "L", false, 0, "")
	}

	pdf.SetY(y + h + 7)
}

func drawFooter(pdf *fpdf.Fpdf, inv domain.Invoice) {
	y := pdf.GetY()
	pdf.SetDrawColor(227, 227, 227)
	pdf.SetLineWidth(0.2)
	pdf.Line(marginLeft, y, pageWidth-marginRight, y)

	footer := inv.SellerFooterNote
	if footer == "" {
		footer = "Prices are inclusive of GST. Computer-generated invoice — no signature required."
	}
	if inv.SellerJurisdiction != "" {
		footer += " Subject to " + inv.SellerJurisdiction + " jurisdiction."
	}

	pdf.SetXY(marginLeft, y+3)
	pdf.SetFont(fontFamily, "", 7.5)
	pdf.SetTextColor(102, 102, 102)
	pdf.MultiCell(contentW-45, 3.6, footer, "", "L", false)

	// Signature block, right-aligned.
	sigX := pageWidth - marginRight - 42
	pdf.SetDrawColor(187, 187, 187)
	pdf.Line(sigX, y+16, pageWidth-marginRight, y+16)
	pdf.SetXY(sigX, y+16.5)
	pdf.SetFont(fontFamily, "", 7)
	pdf.CellFormat(42, 4, "Authorised Signatory", "", 0, "R", false, 0, "")
}

// --- small helpers ---------------------------------------------------------

func label(pdf *fpdf.Fpdf, x, y, w float64, text string) {
	pdf.SetXY(x, y)
	pdf.SetFont(fontFamily, "B", 6.5)
	pdf.SetTextColor(122, 122, 122)
	pdf.CellFormat(w, 3.5, text, "", 0, "L", false, 0, "")
}

func value(pdf *fpdf.Fpdf, x, y, w float64, text, align string) {
	pdf.SetXY(x, y)
	pdf.SetFont(fontFamily, "", 9)
	pdf.SetTextColor(26, 26, 26)
	pdf.CellFormat(w, 5, text, "", 0, align, false, 0, "")
}

func orDash(s string) string {
	if strings.TrimSpace(s) == "" {
		return "—"
	}
	return s
}

// titleCase upper-cases the first letter. strings.Title is deprecated in modern
// Go and golangci-lint (run by `make check`) rejects it.
func titleCase(s string) string {
	if s == "" {
		return s
	}
	return strings.ToUpper(s[:1]) + s[1:]
}

// formatBasisPoints renders 1800 as "18" and 1250 as "12.5".
func formatBasisPoints(bp int) string {
	if bp%100 == 0 {
		return strconv.Itoa(bp / 100)
	}
	return strings.TrimRight(strings.TrimRight(fmt.Sprintf("%.2f", float64(bp)/100), "0"), ".")
}

// formatINR renders minor units with Indian digit grouping: 12345678 paise
// becomes "1,23,456.78". Integer-only arithmetic — no float rounding.
func formatINR(m domain.Money) string {
	v := m.AmountMinor
	sign := ""
	if v < 0 {
		sign = "-"
		v = -v
	}

	rupees := strconv.FormatInt(v/100, 10)
	paise := fmt.Sprintf("%02d", v%100)

	if len(rupees) > 3 {
		head := rupees[:len(rupees)-3]
		tail := rupees[len(rupees)-3:]

		var parts []string
		for len(head) > 2 {
			parts = append([]string{head[len(head)-2:]}, parts...)
			head = head[:len(head)-2]
		}
		if head != "" {
			parts = append([]string{head}, parts...)
		}
		rupees = strings.Join(parts, ",") + "," + tail
	}

	return sign + rupees + "." + paise
}
