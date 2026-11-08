package repository

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/stretchr/testify/assert"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

// findShopByIDMatcher matches the FindShopByID query and fails if it gates on
// shop_status: a direct lookup by ID must return the shop in any state.
var findShopByIDMatcher = sqlmock.QueryMatcherFunc(func(expectedSQL, actualSQL string) error {
	if strings.Contains(actualSQL, ShopActivePredicate) {
		return fmt.Errorf("FindShopByID must not filter on %q, got query: %s", ShopActivePredicate, actualSQL)
	}
	return sqlmock.QueryMatcherRegexp.Match(expectedSQL, actualSQL)
})

func newFindShopByIDRepo(t *testing.T) (*userDatabase, sqlmock.Sqlmock, func()) {
	t.Helper()
	db, mock, err := sqlmock.New(sqlmock.QueryMatcherOption(findShopByIDMatcher))
	assert.NoError(t, err)
	gormDB, err := gorm.Open(postgres.New(postgres.Config{Conn: db}), &gorm.Config{})
	assert.NoError(t, err)
	return &userDatabase{DB: gormDB}, mock, func() { db.Close() }
}

func TestFindShopByID(t *testing.T) {
	const shopID = "shp_test"

	for _, status := range []string{"active", "under_review", "rejected", "suspended", "inactive", ""} {
		status := status
		t.Run(fmt.Sprintf("ShopIsReturnedWithStatus_%q", status), func(t *testing.T) {
			repo, mock, closeDB := newFindShopByIDRepo(t)
			defer closeDB()

			rows := sqlmock.NewRows([]string{"id", "shop_name", "city", "shop_status"}).
				AddRow(shopID, "Test Shop", "Jaipur", status)
			mock.ExpectQuery(`FROM shop_details sd`).WithArgs(shopID).WillReturnRows(rows)

			shop, err := repo.FindShopByID(context.Background(), shopID)
			assert.NoError(t, err)
			assert.Equal(t, shopID, shop.ID)
			assert.Equal(t, "Test Shop", shop.ShopName)
			assert.Equal(t, status, shop.ShopStatus)
			assert.NoError(t, mock.ExpectationsWereMet())
		})
	}

	t.Run("UnknownShopReturnsRecordNotFound", func(t *testing.T) {
		repo, mock, closeDB := newFindShopByIDRepo(t)
		defer closeDB()

		mock.ExpectQuery(`FROM shop_details sd`).WithArgs(shopID).
			WillReturnRows(sqlmock.NewRows([]string{"id", "shop_name"}))

		shop, err := repo.FindShopByID(context.Background(), shopID)
		assert.True(t, errors.Is(err, gorm.ErrRecordNotFound), "expected gorm.ErrRecordNotFound, got %v", err)
		assert.Empty(t, shop.ID)
		assert.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("DatabaseErrorIsWrapped", func(t *testing.T) {
		repo, mock, closeDB := newFindShopByIDRepo(t)
		defer closeDB()

		dbErr := errors.New("connection refused")
		mock.ExpectQuery(`FROM shop_details sd`).WithArgs(shopID).WillReturnError(dbErr)

		_, err := repo.FindShopByID(context.Background(), shopID)
		assert.Error(t, err)
		assert.True(t, errors.Is(err, dbErr))
		assert.False(t, errors.Is(err, gorm.ErrRecordNotFound))
		assert.NoError(t, mock.ExpectationsWereMet())
	})
}
