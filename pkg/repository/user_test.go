package repository

import (
	"context"
	"regexp"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/rohit221990/mandi-backend/pkg/api/handler/request"
	"github.com/rohit221990/mandi-backend/pkg/domain"
	"github.com/stretchr/testify/assert"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

// func TestFindUserByColumnNameAndValue(t *testing.T) {
// 	type columnNameAndValue struct {
// 		columnName string
// 		value      any
// 	}

// 	tests := []struct {
// 		testName       string
// 		input          columnNameAndValue
// 		expectedOutput domain.User
// 		buildStub      func(mock sqlmock.Sqlmock, columnName string)
// 		expectedError  error
// 	}{
// 		{
// 			testName: "invalidColumnNameWillReturnError",
// 			input: columnNameAndValue{
// 				columnName: "nonExistingColumn",
// 			},
// 			expectedOutput: domain.User{},
// 			buildStub: func(mock sqlmock.Sqlmock, columnName string) {
// 				expectedQuery := fmt.Sprintf(`SELECT \* FROM users WHERE %s = \$1`, columnName)
// 				mock.ExpectQuery(expectedQuery).
// 					WillReturnError(fmt.Errorf("%s column not existst", columnName))
// 			},
// 			expectedError: fmt.Errorf("%s column not existst", "nonExistingColumn"),
// 		},
// 		{
// 			testName: "validColumnAndNonExistringValueWillReturnEmptyUser",
// 			input: columnNameAndValue{
// 				columnName: "email",
// 				value:      "nonExistringUser@gmail.com",
// 			},
// 			expectedOutput: domain.User{
// 				ID: 0,
// 			},
// 			buildStub: func(mock sqlmock.Sqlmock, columnName string) {
// 				expectedQuery := fmt.Sprintf(`SELECT \* FROM users WHERE %s = \$1`, columnName)
// 				mock.ExpectQuery(expectedQuery).
// 					WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(0))
// 			},
// 		},
// 		// {
// 		// 	testName: "validColumnNameAndExistingValueWillReturnUser",
// 		// 	input: columnNameAndValue{
// 		// 		columnName: "user_name",

// 		// 	},
// 		// }
// 	}

// 	for _, test := range tests {
// 		t.Run(test.testName, func(t *testing.T) {
// 			db, mock, err := sqlmock.New()
// 			assert.Nil(t, err, "an error '%s' not expected when opening mock database", err)

// 			gormDB, err := gorm.Open(postgres.New(postgres.Config{
// 				Conn: db,
// 			}), &gorm.Config{})
// 			assert.Nil(t, err, "an error '%s' not expected when opening gorm database", err)

// 			test.buildStub(mock, test.input.columnName)

// 			userRepo := NewUserRepository(gormDB)

// 			user, err := userRepo.FindUserByColumnNameAndValue(context.Background(), test.input.columnName, test.input.value)

// 			if test.expectedError == nil {
// 				assert.Nil(t, err)
// 			} else {
// 				assert.Equal(t, test.expectedError, err)
// 			}

// 			assert.Equal(t, test.expectedOutput, user)
// 		})
// 	}
// }

func TestFindUserByEmail(t *testing.T) {
	tests := []struct {
		testName       string
		inputEmail     string
		expectedOutput domain.User
		buildStub      func(mock sqlmock.Sqlmock)
		expectedError  error
	}{
		{
			testName:       "nonExistingEmailReturnEmptyUser",
			inputEmail:     "nonExistingUser@gmail.com",
			expectedOutput: domain.User{},
			buildStub: func(mock sqlmock.Sqlmock) {
				mock.ExpectQuery(`SELECT  \* FROM users WHERE email \= \$1`).
					WithArgs("nonExistingUser@gmail.com").
					WillReturnRows(sqlmock.NewRows([]string{"id", "email"}).
						AddRow(0, ""))
			},
			expectedError: nil,
		},
		{
			testName:   "exsitingEmailReturnUser",
			inputEmail: "existingUser@gmail.com",
			expectedOutput: domain.User{
				ID:        1,
				Email:     "existingUser@gmail.com",
				FirstName: "existingUserUserName",
				Password:  "existingUserHashedPassword",
			},
			buildStub: func(mock sqlmock.Sqlmock) {
				mock.ExpectQuery(`SELECT  \* FROM users WHERE email \= \$1`).
					WithArgs("existingUser@gmail.com").
					WillReturnRows(sqlmock.NewRows([]string{"id", "email", "first_name", "password"}).
						AddRow(1, "existingUser@gmail.com", "existingUserUserName", "existingUserHashedPassword"))
			},
			expectedError: nil,
		},
	}

	for _, test := range tests {
		t.Run(test.testName, func(t *testing.T) {
			db, mock, err := sqlmock.New()
			assert.Nil(t, err, "an error '%s' not expected when opening mock database", err)

			gormDB, err := gorm.Open(postgres.New(postgres.Config{
				Conn: db,
			}), &gorm.Config{})
			assert.Nil(t, err, "an error '%s' not expected when opening gorm database", err)

			test.buildStub(mock)

			userRepo := NewUserRepository(gormDB)

			user, _ := userRepo.FindUserByEmail(context.Background(), test.inputEmail)

			if test.expectedError == nil {
				assert.Nil(t, err)
			} else {
				assert.Equal(t, err, test.expectedError)
			}

			assert.Equal(t, test.expectedOutput, user)
		})
	}
}

func TestSearchShopList_WithGeoDistance(t *testing.T) {
	db, mock, err := sqlmock.New()
	assert.NoError(t, err)

	gormDB, err := gorm.Open(postgres.New(postgres.Config{Conn: db}), &gorm.Config{})
	assert.NoError(t, err)

	queryMatcher := `(?s)SELECT\s+.*distance_km.*FROM shop_details sd\s+LEFT JOIN shop_times st ON st\.shop_id = sd\.id\s+.*sd\.latitude IS NOT NULL AND sd\.longitude IS NOT NULL.*ORDER BY distance_km ASC NULLS LAST, sd\.created_at DESC\s+LIMIT \$5 OFFSET \$6`

	rows := sqlmock.NewRows([]string{
		"id", "shop_name", "email", "phone", "latitude", "longitude",
		"owner_name", "shop_image_url", "address_line1", "address_line2", "city", "country", "state", "pincode",
		"shop_verification_status", "distance_km", "created_at", "updated_at",
	}).AddRow(
		1, "Nearby Shop", "shop@example.com", "9999999999", 12.97, 77.59,
		"Owner", "/uploads/shops/1.png", "Addr1", "Addr2", "City", "India", "State", "560001",
		true, 2.35, time.Now(), time.Now(),
	)

	mock.ExpectQuery(queryMatcher).
		WithArgs("%mart%", 12.9716, 77.5946, 5.0, 25, 0).
		WillReturnRows(rows)

	repo := NewUserRepository(gormDB)
	shops, err := repo.SearchShopList(context.Background(), request.SearchShopListRequest{
		Query:     "mart",
		Latitude:  12.9716,
		Longitude: 77.5946,
		Radius:    5,
		Pagination: request.Pagination{
			Limit:  25,
			Offset: 0,
		},
	})

	assert.NoError(t, err)
	assert.Len(t, shops, 1)
	if assert.NotNil(t, shops[0].DistanceKm) {
		assert.InDelta(t, 2.35, *shops[0].DistanceKm, 0.0001)
	}
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestSearchShopList_WithPincode(t *testing.T) {
	db, mock, err := sqlmock.New()
	assert.NoError(t, err)

	gormDB, err := gorm.Open(postgres.New(postgres.Config{Conn: db}), &gorm.Config{})
	assert.NoError(t, err)

	queryMatcher := `(?s)SELECT\s+.*NULL::double precision AS distance_km.*FROM shop_details sd\s+LEFT JOIN shop_times st ON st\.shop_id = sd\.id\s+.*sd\.pincode = \$1\s+.*ORDER BY sd\.created_at DESC\s+LIMIT \$2 OFFSET \$3`

	rows := sqlmock.NewRows([]string{
		"id", "shop_name", "email", "phone", "latitude", "longitude",
		"owner_name", "shop_image_url", "address_line1", "address_line2", "city", "country", "state", "pincode",
		"shop_verification_status", "distance_km", "created_at", "updated_at",
	}).AddRow(
		2, "Pincode Shop", "pin@example.com", "8888888888", 12.90, 77.60,
		"Owner2", "/uploads/shops/2.png", "Addr1", "Addr2", "City", "India", "State", "560001",
		true, nil, time.Now(), time.Now(),
	)

	pincode := uint(560001)
	mock.ExpectQuery(queryMatcher).
		WithArgs("560001", 10, 5).
		WillReturnRows(rows)

	repo := NewUserRepository(gormDB)
	shops, err := repo.SearchShopList(context.Background(), request.SearchShopListRequest{
		Pincode: &pincode,
		Pagination: request.Pagination{
			Limit:  10,
			Offset: 5,
		},
	})

	assert.NoError(t, err)
	assert.Len(t, shops, 1)
	assert.Nil(t, shops[0].DistanceKm)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestSearchShopList_WithTextOnly(t *testing.T) {
	db, mock, err := sqlmock.New()
	assert.NoError(t, err)

	gormDB, err := gorm.Open(postgres.New(postgres.Config{Conn: db}), &gorm.Config{})
	assert.NoError(t, err)

	queryMatcher := `(?s)SELECT\s+.*NULL::double precision AS distance_km.*FROM shop_details sd\s+LEFT JOIN shop_times st ON st\.shop_id = sd\.id\s+.*\(sd\.shop_name ILIKE \$1 OR sd\.owner_name ILIKE \$1\).*ORDER BY sd\.created_at DESC\s+LIMIT \$2 OFFSET \$3`

	rows := sqlmock.NewRows([]string{
		"id", "shop_name", "email", "phone", "latitude", "longitude",
		"owner_name", "shop_image_url", "address_line1", "address_line2", "city", "country", "state", "pincode",
		"shop_verification_status", "distance_km", "created_at", "updated_at",
	}).AddRow(
		3, "Text Shop", "text@example.com", "7777777777", 0.0, 0.0,
		"Owner3", "", "", "", "City", "India", "State", "500001",
		false, nil, time.Now(), time.Now(),
	)

	mock.ExpectQuery(queryMatcher).
		WithArgs("%fresh%", 20, 0).
		WillReturnRows(rows)

	repo := NewUserRepository(gormDB)
	shops, err := repo.SearchShopList(context.Background(), request.SearchShopListRequest{
		Query: "fresh",
		Pagination: request.Pagination{
			Limit:  20,
			Offset: 0,
		},
	})

	assert.NoError(t, err)
	assert.Len(t, shops, 1)
	assert.Equal(t, "Text Shop", shops[0].ShopName)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestSearchShopList_QueryError(t *testing.T) {
	db, mock, err := sqlmock.New()
	assert.NoError(t, err)

	gormDB, err := gorm.Open(postgres.New(postgres.Config{Conn: db}), &gorm.Config{})
	assert.NoError(t, err)

	mock.ExpectQuery(regexp.QuoteMeta("FROM shop_details sd")).WillReturnError(assert.AnError)

	repo := NewUserRepository(gormDB)
	_, err = repo.SearchShopList(context.Background(), request.SearchShopListRequest{
		Pagination: request.Pagination{Limit: 10, Offset: 0},
	})

	assert.Error(t, err)
	assert.NoError(t, mock.ExpectationsWereMet())
}
