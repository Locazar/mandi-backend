package repository

import (
	"context"
	"errors"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/stretchr/testify/assert"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

func newFindShopByIDRepo(t *testing.T) (*userDatabase, sqlmock.Sqlmock, func()) {
	t.Helper()
	db, mock, err := sqlmock.New()
	assert.NoError(t, err)
	gormDB, err := gorm.Open(postgres.New(postgres.Config{Conn: db}), &gorm.Config{})
	assert.NoError(t, err)
	return &userDatabase{DB: gormDB}, mock, func() { db.Close() }
}

func TestFindShopByID(t *testing.T) {
	const shopID = "shp_test"

	t.Run("ActiveShopIsReturned", func(t *testing.T) {
		repo, mock, closeDB := newFindShopByIDRepo(t)
		defer closeDB()

		rows := sqlmock.NewRows([]string{"id", "shop_name", "city"}).
			AddRow(shopID, "Test Shop", "Jaipur")
		mock.ExpectQuery(`FROM shop_details sd`).WithArgs(shopID).WillReturnRows(rows)

		shop, err := repo.FindShopByID(context.Background(), shopID)
		assert.NoError(t, err)
		assert.Equal(t, shopID, shop.ID)
		assert.Equal(t, "Test Shop", shop.ShopName)
		assert.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("MissingOrInactiveShopReturnsRecordNotFound", func(t *testing.T) {
		repo, mock, closeDB := newFindShopByIDRepo(t)
		defer closeDB()

		// The query filters on shop_status = 'active', so a shop that exists but
		// is under_review/rejected/suspended yields zero rows, same as an unknown ID.
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
