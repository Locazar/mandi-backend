package repository

import (
	"context"
	"sync"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// testDB lives in invoice_test.go (same package) — reuse it, do not redeclare.

// Invoice numbers must be gapless and unique even when several payments are
// verified at the same instant — the exact situation a naive
// "SELECT max+1 then INSERT" would corrupt.
func TestAllocateInvoiceSequenceIsConcurrencySafe(t *testing.T) {
	db := testDB(t)
	repo := NewInvoiceRepository(db)

	const fy = "2099-00" // isolated from real data
	require.NoError(t, db.Exec("DELETE FROM invoice_number_sequences WHERE financial_year = ?", fy).Error)
	t.Cleanup(func() {
		db.Exec("DELETE FROM invoice_number_sequences WHERE financial_year = ?", fy)
	})

	const workers = 50
	var (
		wg   sync.WaitGroup
		mu   sync.Mutex
		seen = make(map[int]bool, workers)
	)

	for i := 0; i < workers; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			seq, err := repo.AllocateInvoiceSequence(context.Background(), fy)
			assert.NoError(t, err)
			mu.Lock()
			seen[seq] = true
			mu.Unlock()
		}()
	}
	wg.Wait()

	// Exactly the numbers 1..workers, each once — no duplicates, no gaps.
	assert.Len(t, seen, workers)
	for i := 1; i <= workers; i++ {
		assert.True(t, seen[i], "sequence %d missing — allocation is not gapless", i)
	}
}
