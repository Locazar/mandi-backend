package invoice

import (
	"bytes"
	"compress/zlib"
	"context"
	"image"
	"image/color"
	"image/png"
	"io"
	"os"
	"regexp"
	"testing"
	"time"

	"github.com/rohit221990/mandi-backend/pkg/domain"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func sampleInvoice() domain.Invoice {
	paidAt := time.Date(2026, 7, 30, 16, 18, 0, 0, time.UTC)
	return domain.Invoice{
		ID:                  "inv_test",
		SubscriptionOrderID: "subo_abc",
		UserID:              "adm_seller1",
		InvoiceNumber:       "LZ/2026-27/000042",
		FinancialYear:       "2026-27",
		SequenceNumber:      42,
		InvoiceDate:         paidAt,

		SellerLegalName: "Locazar Technologies Pvt. Ltd.",
		SellerGSTIN:     "29AABCL1234M1Z7",
		SellerPAN:       "AABCL1234M",
		SellerAddress:   "No. 42, 3rd Floor, 12th Main\nIndiranagar\nBengaluru 560038\nKarnataka, India",
		SellerSACCode:   "998599",
		PlaceOfSupply:   "Karnataka (29)",

		BuyerShopID:      "shp_8f2ac91b",
		BuyerName:        "Sharma Kirana Store",
		BuyerContactName: "Rakesh Sharma",
		BuyerAddress:     "Shop 7, Bannerghatta Road\nBengaluru 560076\nKarnataka, India",
		BuyerGSTIN:       "29ABCDE1234F2Z5",

		PlanName:        "Seller Subscription — 3 Months",
		PlanDescription: "Online marketplace listing & support services",
		PeriodStart:     paidAt,
		PeriodEnd:       paidAt.AddDate(0, 0, 90),

		Total:              domain.INR(149900),
		TaxableValue:       domain.INR(127034),
		GSTAmount:          domain.INR(22866),
		GSTRateBasisPoints: 1800,

		RazorpayPaymentID: "pay_QxK9mR2VbL7ta1",
		RazorpayOrderID:   "order_QxK8pJ4YdN2sc9",
		PaymentMethod:     "razorpay",
		PaidAt:            paidAt,
	}
}

func TestRenderProducesValidPDF(t *testing.T) {
	r := NewPDFRenderer()
	out, err := r.Render(context.Background(), sampleInvoice(), nil)

	require.NoError(t, err)
	assert.NotEmpty(t, out)
	// Every PDF starts with the %PDF magic bytes.
	assert.True(t, bytes.HasPrefix(out, []byte("%PDF")), "output is not a PDF")
	assert.True(t, bytes.Contains(out, []byte("%%EOF")), "PDF is not terminated")
	// A one-page invoice with a logo should still be small.
	assert.Less(t, len(out), 2_000_000)

	// Write it out for eyeballing; harmless in CI.
	if dir := os.Getenv("INVOICE_PDF_OUT"); dir != "" {
		require.NoError(t, os.WriteFile(dir+"/invoice-sample.pdf", out, 0o644))
	}
}

func TestRenderIncludesBuyerGSTINWhenPresent(t *testing.T) {
	// Positive control for TestRenderOmitsBuyerGSTINWhenAbsent below: proves
	// pdfContainsText can actually find the GSTIN when it IS drawn, so the
	// negative case isn't passing just because the check never matches anything.
	r := NewPDFRenderer()
	out, err := r.Render(context.Background(), sampleInvoice(), nil)
	require.NoError(t, err)

	assert.True(t, pdfContainsText(t, out, "29ABCDE1234F2Z5"),
		"rendered PDF does not contain the buyer's GSTIN, but BuyerGSTIN was set")
}

func TestRenderOmitsBuyerGSTINWhenAbsent(t *testing.T) {
	inv := sampleInvoice()
	inv.BuyerGSTIN = ""

	r := NewPDFRenderer()
	out, err := r.Render(context.Background(), inv, nil)

	require.NoError(t, err)
	assert.True(t, bytes.HasPrefix(out, []byte("%PDF")))
	// The whole point of this test: the seller's own GSTIN is set in
	// sampleInvoice and must still print (Task 5's rule only omits the BUYER's
	// GSTIN for non-GST sellers), but the buyer's must not appear anywhere.
	assert.True(t, pdfContainsText(t, out, "29AABCL1234M1Z7"),
		"seller GSTIN must still print when only the buyer's GSTIN is absent")
	assert.False(t, pdfContainsText(t, out, "29ABCDE1234F2Z5"),
		"buyer GSTIN must not appear anywhere in the PDF when BuyerGSTIN is empty")

	// A blank BuyerGSTIN produces the same VALUE as no-value-at-all, so the two
	// checks above cannot catch a regression that still reads inv.BuyerGSTIN
	// but drops the `!= ""` guard — that only draws an empty "GSTIN" label,
	// not the (already-blank) number. Count occurrences of the "GSTIN" label
	// itself: it must appear exactly once (the seller's), not twice.
	gstinLabelCount := 0
	for _, stream := range decompressedStreams(out) {
		gstinLabelCount += bytes.Count(stream, utf16BE("GSTIN"))
	}
	assert.Equal(t, 1, gstinLabelCount,
		"exactly one \"GSTIN\" label (the seller's) must appear when the buyer's is absent — "+
			"a second occurrence means an empty buyer GSTIN line rendered anyway")
}

func TestRenderToleratesMissingLogo(t *testing.T) {
	r := NewPDFRenderer()

	nilLogoOut, err := r.Render(context.Background(), sampleInvoice(), nil)
	require.NoError(t, err)

	// Corrupt bytes must be skipped, not fatal — a broken logo must never stop
	// a seller getting their invoice.
	corruptLogoOut, err := r.Render(context.Background(), sampleInvoice(), []byte("not-a-png"))
	require.NoError(t, err)
	assert.True(t, bytes.HasPrefix(corruptLogoOut, []byte("%PDF")))

	// Stronger than "still produces a PDF": corrupt bytes must render EXACTLY
	// as if no logo were supplied at all, not merely "not crash". The page
	// content stream — the actual drawing operators — must be byte-identical
	// between the two, and neither may embed an image XObject.
	nilContent := pdfPageContentStream(t, nilLogoOut)
	corruptContent := pdfPageContentStream(t, corruptLogoOut)
	require.NotEmpty(t, nilContent, "could not extract a page content stream to compare")
	assert.True(t, bytes.Equal(nilContent, corruptContent),
		"corrupt logo bytes changed the page content stream — it should render identically to no logo at all")

	assert.False(t, bytes.Contains(corruptLogoOut, []byte("/Subtype/Image")) ||
		bytes.Contains(corruptLogoOut, []byte("/Subtype /Image")),
		"corrupt logo bytes must not be embedded as an image XObject")
}

func TestRenderEmbedsValidLogo(t *testing.T) {
	// Positive control for TestRenderToleratesMissingLogo: proves a real PNG
	// DOES change the output, so "identical to no-logo" above is meaningful
	// evidence of graceful degradation rather than the renderer ignoring the
	// logo parameter entirely.
	r := NewPDFRenderer()
	nilLogoOut, err := r.Render(context.Background(), sampleInvoice(), nil)
	require.NoError(t, err)

	validOut, err := r.Render(context.Background(), sampleInvoice(), validPNGFixture(t))
	require.NoError(t, err)

	assert.True(t, bytes.Contains(validOut, []byte("/Subtype/Image")) ||
		bytes.Contains(validOut, []byte("/Subtype /Image")),
		"a valid logo should be embedded as an image XObject")

	nilContent := pdfPageContentStream(t, nilLogoOut)
	validContent := pdfPageContentStream(t, validOut)
	assert.False(t, bytes.Equal(nilContent, validContent),
		"a valid logo should change the page content stream (an image-drawing operator is added)")
}

// --- PDF content-inspection helpers -----------------------------------------
//
// fpdf compresses page content streams with zlib and, when a UTF8 font is
// active (AddUTF8FontFromBytes — used throughout this renderer for the ₹
// glyph), encodes drawn text as UTF-16BE string literals rather than raw
// ASCII. A plain bytes.Contains(pdf, []byte("some text")) therefore never
// matches real content and silently proves nothing. These helpers decompress
// every content stream and search using the actual encoding, so assertions
// about what text appears (or doesn't) in the rendered PDF are real.

var pdfStreamPattern = regexp.MustCompile(`(?s)stream\r?\n(.*?)\r?\nendstream`)

// pdfContainsText reports whether s (encoded the way fpdf's UTF8 fonts encode
// drawn text — UTF-16BE) appears in any decompressed content stream of pdf.
func pdfContainsText(t *testing.T, pdf []byte, s string) bool {
	t.Helper()
	want := utf16BE(s)
	for _, stream := range decompressedStreams(pdf) {
		if bytes.Contains(stream, want) {
			return true
		}
	}
	return false
}

// pdfPageContentStream returns the decompressed page content stream — the one
// containing the page's drawing operators (identifiable by a "Tj" text-show
// operator) — for comparing two renders' actual drawn output.
func pdfPageContentStream(t *testing.T, pdf []byte) []byte {
	t.Helper()
	for _, stream := range decompressedStreams(pdf) {
		if bytes.Contains(stream, []byte("Tj")) {
			return stream
		}
	}
	return nil
}

func decompressedStreams(pdf []byte) [][]byte {
	var out [][]byte
	for _, m := range pdfStreamPattern.FindAllSubmatch(pdf, -1) {
		zr, err := zlib.NewReader(bytes.NewReader(m[1]))
		if err != nil {
			continue // not zlib-compressed (e.g. raw font/image binary data)
		}
		decoded, err := io.ReadAll(zr)
		if err != nil {
			continue
		}
		out = append(out, decoded)
	}
	return out
}

// utf16BE encodes s as fpdf encodes text drawn with a UTF8 font: two bytes per
// rune, big-endian. Every string used in these tests is ASCII, so this never
// needs to handle surrogate pairs.
func utf16BE(s string) []byte {
	out := make([]byte, 0, len(s)*2)
	for _, r := range s {
		out = append(out, byte(r>>8), byte(r))
	}
	return out
}

// validPNGFixture returns a minimal well-formed 1x1 red PNG, built in-process
// so the test has no external file dependency.
func validPNGFixture(t *testing.T) []byte {
	t.Helper()
	var buf bytes.Buffer
	img := image.NewRGBA(image.Rect(0, 0, 1, 1))
	img.Set(0, 0, color.RGBA{R: 255, A: 255})
	require.NoError(t, png.Encode(&buf, img))
	return buf.Bytes()
}

func TestFormatINR(t *testing.T) {
	assert.Equal(t, "1,499.00", formatINR(domain.INR(149900)))
	assert.Equal(t, "1,270.34", formatINR(domain.INR(127034)))
	assert.Equal(t, "228.66", formatINR(domain.INR(22866)))
	assert.Equal(t, "0.00", formatINR(domain.INR(0)))
	assert.Equal(t, "12,34,567.89", formatINR(domain.INR(123456789))) // Indian grouping
}

// TestRenderUsesInvoiceFooterNoteWhenSet guards a real bug: the footer was
// previously a hardcoded string that ignored inv.SellerFooterNote entirely,
// so an admin editing the "footer note" field in the portal had zero effect
// on any rendered invoice — the field looked live but silently did nothing.
func TestRenderUsesInvoiceFooterNoteWhenSet(t *testing.T) {
	inv := sampleInvoice()
	inv.SellerFooterNote = "This is a distinctive custom footer set by an admin."
	inv.SellerJurisdiction = "Chennai"

	r := NewPDFRenderer()
	out, err := r.Render(context.Background(), inv, nil)
	require.NoError(t, err)

	assert.True(t, pdfContainsText(t, out, "This is a distinctive custom footer set by an admin."),
		"custom SellerFooterNote did not appear in the rendered PDF")
	assert.True(t, pdfContainsText(t, out, "Chennai"),
		"SellerJurisdiction did not appear in the rendered PDF")
}

// TestRenderFallsBackToDefaultFooterWhenUnset covers invoices issued before
// the footer/jurisdiction fields existed (or with a blank profile field) —
// the renderer must still produce sensible boilerplate, not an empty footer.
func TestRenderFallsBackToDefaultFooterWhenUnset(t *testing.T) {
	inv := sampleInvoice()
	inv.SellerFooterNote = ""
	inv.SellerJurisdiction = ""

	r := NewPDFRenderer()
	out, err := r.Render(context.Background(), inv, nil)
	require.NoError(t, err)

	assert.True(t, pdfContainsText(t, out, "Computer-generated invoice"),
		"default footer text should still appear when SellerFooterNote is unset")
}
