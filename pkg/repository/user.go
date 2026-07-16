package repository

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/rohit221990/mandi-backend/pkg/api/handler/request"
	"github.com/rohit221990/mandi-backend/pkg/api/handler/response"
	"github.com/rohit221990/mandi-backend/pkg/domain"
	"github.com/rohit221990/mandi-backend/pkg/repository/interfaces"
	"gorm.io/gorm"
)

type userDatabase struct {
	DB *gorm.DB
}

func NewUserRepository(DB *gorm.DB) interfaces.UserRepository {
	return &userDatabase{DB: DB}
}

func (c *userDatabase) FindUserByUserID(ctx context.Context, userID string) (user domain.User, err error) {

	query := `SELECT * FROM users WHERE id = $1`
	err = c.DB.Raw(query, userID).Scan(&user).Error

	return user, err
}

func (c *userDatabase) FindUserByEmail(ctx context.Context, email string) (user domain.User, err error) {

	query := `SELECT * FROM users WHERE email = $1`

	err = c.DB.Raw(query, email).Scan(&user).Error

	return user, err
}

func (c *userDatabase) FindUserByPhoneNumber(ctx context.Context, phoneNumber string) (user domain.User, err error) {

	query := `SELECT * FROM users WHERE phone = $1`
	err = c.DB.Raw(query, phoneNumber).Scan(&user).Error

	return user, err
}
func (c *userDatabase) FindUserByUserName(ctx context.Context, userName string) (user domain.User, err error) {

	query := `SELECT * FROM users WHERE user_name = $1`
	err = c.DB.Raw(query, userName).Scan(&user).Error

	return user, err
}

func (c *userDatabase) FindUserByUserNameEmailOrPhoneNotID(ctx context.Context,
	userDetails domain.User) (user domain.User, err error) {
	// Debugging line

	query := `SELECT * FROM users WHERE (email = $1 OR phone = $2) AND id != $3`
	err = c.DB.Raw(query, userDetails.Email, userDetails.Phone, userDetails.ID).Scan(&user).Error

	return
}

func (c *userDatabase) SaveUser(ctx context.Context, user domain.User) (userID string, err error) {
	// Build dynamic column list and values
	columns := []string{}
	placeholders := []string{}
	values := []interface{}{}
	paramCount := 1

	// Always generate and include the ID (string PK, no DB-side autoincrement)
	if user.ID == "" {
		user.ID = domain.NewID(domain.PrefixUser)
	}
	columns = append(columns, "id")
	placeholders = append(placeholders, fmt.Sprintf("$%d", paramCount))
	values = append(values, user.ID)
	paramCount++

	if user.Email != "" {
		columns = append(columns, "email")
		placeholders = append(placeholders, fmt.Sprintf("$%d", paramCount))
		values = append(values, user.Email)
		paramCount++
	}
	if user.Phone != "" {
		columns = append(columns, "phone")
		placeholders = append(placeholders, fmt.Sprintf("$%d", paramCount))
		values = append(values, user.Phone)
		paramCount++
	}
	// password is NOT NULL in the DB; phone-only users have no password so we store
	// an empty string as a safe placeholder (auth is via OTP, not password, for these users)
	columns = append(columns, "password")
	placeholders = append(placeholders, fmt.Sprintf("$%d", paramCount))
	values = append(values, user.Password) // empty string is fine for phone-only signups
	paramCount++

	// Always add created_at
	columns = append(columns, "created_at")
	placeholders = append(placeholders, fmt.Sprintf("$%d", paramCount))
	values = append(values, time.Now())

	// Build dynamic query
	query := fmt.Sprintf("INSERT INTO users (%s) VALUES (%s) RETURNING id",
		strings.Join(columns, ", "),
		strings.Join(placeholders, ", "))

	err = c.DB.Raw(query, values...).Scan(&userID).Error
	return userID, err
}

func (c *userDatabase) UpdateVerified(ctx context.Context, userID string) error {

	query := `UPDATE users SET phone_verified = true WHERE id = $1`
	err := c.DB.Exec(query, userID).Error

	return err
}

func (c *userDatabase) UpdateAdminVerified(ctx context.Context, adminID string) error {

	query := `UPDATE admins SET verified_seller = 'T' WHERE id = $1`
	err := c.DB.Exec(query, adminID).Error

	return err
}

func (c *userDatabase) UpdateUser(ctx context.Context, user domain.User) (err error) {

	updatedAt := time.Now()
	if user.Password != "" && user.Email != "" {
		query := `UPDATE users SET first_name = $1, last_name = $2, date_of_birth = $3,
		email = $4, phone = $5, password = $6, updated_at = $7 WHERE id = $8`
		err = c.DB.Exec(query, user.FirstName, user.LastName, user.DateOfBirth, user.Email,
			user.Phone, user.Password, updatedAt, user.ID).Error
	} else if user.Password != "" {
		query := `UPDATE users SET first_name = $1, last_name = $2, date_of_birth = $3,
		phone = $4, password = $5, updated_at = $6 WHERE id = $7`
		err = c.DB.Exec(query, user.FirstName, user.LastName, user.DateOfBirth,
			user.Phone, user.Password, updatedAt, user.ID).Error
	} else if user.Email != "" {
		query := `UPDATE users SET first_name = $1, last_name = $2, date_of_birth = $3,
		email = $4, phone = $5, updated_at = $6 WHERE id = $7`
		err = c.DB.Exec(query, user.FirstName, user.LastName, user.DateOfBirth, user.Email,
			user.Phone, updatedAt, user.ID).Error
	} else {
		query := `UPDATE users SET first_name = $1, last_name = $2, date_of_birth = $3,
		phone = $4, updated_at = $5 WHERE id = $6`
		err = c.DB.Exec(query, user.FirstName, user.LastName, user.DateOfBirth,
			user.Phone, updatedAt, user.ID).Error
	}

	if err != nil {
		return fmt.Errorf("filed to update user detail of user with user_id %s", user.ID)
	}
	return nil
}

func (c *userDatabase) UpdateBlockStatus(ctx context.Context, userID string, blockStatus bool) error {

	query := `UPDATE users SET block_status = $1 WHERE id = $2`
	err := c.DB.Exec(query, blockStatus, userID).Error

	return err
}

func (c *userDatabase) IsAddressIDExist(ctx context.Context, addressID string) (exist bool, err error) {
	query := `SELECT EXISTS(SELECT 1 FROM addresses WHERE id = $1) AS exist FROM addresses`
	err = c.DB.Raw(query, addressID).Scan(&exist).Error

	return
}
func (c *userDatabase) FindAddressByID(ctx context.Context, addressID string) (address response.Address, err error) {

	query := `SELECT adrs.id, adrs.land_mark, adrs.area, adrs.city, adrs.pincode, adrs.country_id, c.country_name, adrs.latitude, adrs.longitude, adrs.phone_number, adrs.address_type, adrs.address_line1, adrs.address_line2, adrs.is_default
	FROM addresses adrs 
	INNER JOIN countries c ON c.id = adrs.country_id  
	WHERE adrs.id = $1 `
	err = c.DB.Raw(query, addressID).Scan(&address).Error

	return
}

func (c *userDatabase) IsAddressAlreadyExistForUser(ctx context.Context, address domain.Address, userID string) (exist bool, err error) {

	query := `SELECT DISTINCT CASE  WHEN adrs.id != 0 THEN 'T' ELSE 'F' END AS exist 
	FROM addresses adrs 
	INNER JOIN user_addresses urs ON adrs.id = urs.address_id 
	WHERE adrs.land_mark = $1 AND adrs.city = $2 AND adrs.pincode = $3 AND adrs.country_id = $4 AND urs.user_id = $5`
	err = c.DB.Raw(query, address.LandMark, address.City, address.Pincode, address.CountryID, userID).Scan(&exist).Error
	if err != nil {
		return exist, fmt.Errorf("filed to check address already exist for user with user_id %s", userID)
	}
	return
}

func (c *userDatabase) FindAllAddressByUserID(ctx context.Context, userID string) (addresses []response.Address, err error) {

	query := `SELECT a.id, a.land_mark, a.area, a.city, a.pincode, a.country_id, c.country_name, a.latitude, a.longitude, a.phone_number, a.address_type, a.address_line1, a.address_line2, a.is_default, ua.is_default
	FROM user_addresses ua JOIN addresses a ON ua.address_id=a.id 
	INNER JOIN countries c ON a.country_id=c.id WHERE ua.user_id = $1`

	err = c.DB.Raw(query, userID).Scan(&addresses).Error

	return addresses, err
}

func (c *userDatabase) FindCountryByID(ctx context.Context, countryID string) (domain.Country, error) {

	var country domain.Country

	if c.DB.Raw("SELECT * FROM countries WHERE id = ?", countryID).Scan(&country).Error != nil {
		return country, errors.New("filed to find the country")
	}

	return country, nil
}

// save address
func (c *userDatabase) SaveAddress(ctx context.Context, address domain.Address) (addressID string, err error) {
	query := `INSERT INTO addresses (user_id, land_mark, area, city, pincode, country_id, latitude, longitude, phone_number, address_type, address_line1, address_line2, is_default, created_at, updated_at) 
	VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15) RETURNING id`

	now := time.Now()

	if c.DB.Raw(query, address.UserID, address.LandMark, address.Area, address.City, address.Pincode, address.CountryID,
		address.Latitude, address.Longitude, address.PhoneNumber, address.AddressType, address.AddressLine1, address.AddressLine2, address.IsDefault, now, now,
	).Scan(&addressID).Error != nil {
		return addressID, errors.New("failed to insert address on database")
	}
	return addressID, nil
}

// update address
func (c *userDatabase) UpdateAddress(ctx context.Context, address domain.Address) error {
	query := `UPDATE addresses SET land_mark=$1, area=$2, city=$3, pincode=$4, country_id=$5, latitude=$6, longitude=$7, phone_number=$8, address_type=$9, address_line1=$10, address_line2=$11, is_default=$12, updated_at=$13 WHERE id=$14`

	updatedAt := time.Now()
	if c.DB.Raw(query, address.LandMark, address.Area, address.City, address.Pincode, address.CountryID,
		address.Latitude, address.Longitude, address.PhoneNumber, address.AddressType, address.AddressLine1, address.AddressLine2, address.IsDefault, updatedAt, address.ID).Scan(&address).Error != nil {
		return errors.New("failed to update the address for edit address")
	}
	return nil
}

func (c *userDatabase) SaveUserAddress(ctx context.Context, userAddress domain.UserAddress) error {

	// first check user's first address is this or not
	var userID uint
	query := `SELECT address_id FROM user_addresses WHERE user_id = $1`
	err := c.DB.Raw(query, userAddress.UserID).Scan(&userID).Error
	if err != nil {
		return fmt.Errorf("filed to check user have already address exit or not with user_id %v", userAddress.UserID)
	}

	// if the given address is need to set default  then remove all other from default
	if userID == 0 { // it means user have no other addresses
		userAddress.IsDefault = true
	} else if userAddress.IsDefault {
		query := `UPDATE user_addresses SET is_default = 'f' WHERE user_id = ?`
		if c.DB.Raw(query, userAddress.UserID).Scan(&userAddress).Error != nil {
			return errors.New("filed to remove default status of address")
		}
	}

	query = `INSERT INTO user_addresses (user_id,address_id,is_default) VALUES ($1, $2, $3)`
	err = c.DB.Exec(query, userAddress.UserID, userAddress.AddressID, userAddress.IsDefault).Error
	if err != nil {
		return errors.New("filed to insert userAddress on database")
	}
	return nil
}

func (c *userDatabase) UpdateUserAddress(ctx context.Context, userAddress domain.UserAddress) error {
	// if it need to set default the change the old default
	if userAddress.IsDefault {

		query := `UPDATE user_addresses SET is_default = 'f' WHERE user_id = ?`
		if c.DB.Raw(query, userAddress.UserID).Scan(&userAddress).Error != nil {
			return errors.New("filed to remove default status of address")
		}
	}

	// update the user address
	query := `UPDATE user_addresses SET is_default = ? WHERE address_id=? AND user_id=?`
	if c.DB.Raw(query, userAddress.IsDefault, userAddress.AddressID, userAddress.UserID).Scan(&userAddress).Error != nil {
		return errors.New("filed to update user address")
	}
	return nil
}

// wish list

func (c *userDatabase) FindWishListItem(ctx context.Context, productID, userID string) (domain.WishList, error) {

	var wishList domain.WishList
	query := `SELECT * FROM wish_lists WHERE user_id=? AND product_item_id=?`
	if c.DB.Raw(query, userID, productID).Scan(&wishList).Error != nil {
		return wishList, errors.New("filed to find wishlist item")
	}
	return wishList, nil
}

func (c *userDatabase) FindAllWishListItemsByUserID(ctx context.Context, userID string) (productItems []response.WishListItem, err error) {

	query := `SELECT p.name, wl.id, pi.id AS product_item_id, pi.product_id, FROM wish_lists wl 
	INNER JOIN product_items pi ON wl.product_item_id = pi.id 
	INNER JOIN products p ON pi.product_id = p.id 
	AND wl.user_id = $1`
	err = c.DB.Raw(query, userID).Scan(&productItems).Error

	return
}

func (c *userDatabase) SaveWishListItem(ctx context.Context, wishList domain.WishList) error {

	query := `INSERT INTO wish_lists (user_id,product_item_id,shop_id,admin_id) VALUES ($1,$2,$3,$4) RETURNING *`

	if c.DB.Raw(query, wishList.UserID, wishList.ProductItemID, wishList.ShopID, wishList.AdminID).Scan(&wishList).Error != nil {
		return errors.New("filed to insert new wishlist on database")
	}
	return nil
}

func (c *userDatabase) RemoveWishListItem(ctx context.Context, userID, productItemID string) error {

	query := `DELETE FROM wish_lists WHERE product_item_id = $1 AND user_id = $2`
	err := c.DB.Exec(query, productItemID, userID).Error

	return err
}

func (c *userDatabase) FindSellersByRadius(ctx context.Context, reqData request.SellerRadiusRequest) (sellers []response.Shop, err error) {
	query := `
		SELECT * FROM (
	 SELECT a.id, a.shop_name, a.email, a.phone, a.latitude, a.longitude,
		a.owner_name, a.shop_image_url, a.address_line1, a.address_line2, a.city, a.country, a.state, a.pincode,
		a.shop_verification_status, a.created_at, a.updated_at,
			(6371 * acos(
					cos(radians($1)) * cos(radians(a.latitude)) *
					cos(radians(a.longitude) - radians($2)) +
					sin(radians($1)) * sin(radians(a.latitude))
			)) AS distance_km
		FROM shop_details a
		WHERE a.latitude IS NOT NULL AND a.longitude IS NOT NULL AND a.shop_status = 'active'
	) AS subquery
	WHERE distance_km <= $3
	LIMIT $4 OFFSET $5
	`

	err = c.DB.Raw(query, reqData.Latitude, reqData.Longitude, reqData.RadiusKm, reqData.Limit, reqData.Offset).Scan(&sellers).Error

	return sellers, err
}

func (c *userDatabase) FindSellersByPincode(ctx context.Context, reqData request.SellerPincodeRequest) (sellers []response.Shop, err error) {
	query := `
		SELECT id, shop_name, email, phone, latitude, longitude,
		owner_name, shop_image_url, address_line1, address_line2, city, country, state, pincode,
		shop_verification_status, created_at, updated_at
		FROM shop_details
		WHERE pincode = $1 AND shop_status = 'active'
		LIMIT $2 OFFSET $3
	`

	err = c.DB.Raw(query, reqData.Pincode, reqData.Limit, reqData.Offset).Scan(&sellers).Error

	return sellers, err
}

func (c *userDatabase) SearchShopList(ctx context.Context, reqData request.SearchShopListRequest) (shops []response.Shop, err error) {
	paramIndex := 1
	args := []interface{}{}
	whereClause := " WHERE 1=1"
	orderBy := " ORDER BY sd.id, sd.created_at DESC"
	distanceExpr := "NULL::double precision AS distance_km"

	// Only approved, live shops are visible to customers — shop_status is
	// the single source of truth for the shop approval lifecycle.
	whereClause += " AND sd.shop_status = 'active'"

	// Add search query condition if provided.
	if reqData.Query != "" {
		whereClause += fmt.Sprintf(` AND (sd.shop_name ILIKE $%d OR sd.owner_name ILIKE $%d)`, paramIndex, paramIndex)
		args = append(args, "%"+reqData.Query+"%")
		paramIndex++
	}

	// Filter by geolocation (lat + long + radius) OR pincode, but not both.
	if reqData.Latitude != 0 && reqData.Longitude != 0 && reqData.Radius > 0 {
		latIdx := paramIndex
		lonIdx := paramIndex + 1
		radiusKmIdx := paramIndex + 2

		distanceExpr = fmt.Sprintf(
			`(6371 * acos(
				LEAST(1, GREATEST(-1,
					cos(radians($%d)) * cos(radians(sd.latitude::double precision)) *
					cos(radians(sd.longitude::double precision) - radians($%d)) +
					sin(radians($%d)) * sin(radians(sd.latitude::double precision))
				))
			)) AS distance_km`,
			latIdx,
			lonIdx,
			latIdx,
		)
		whereClause += fmt.Sprintf(
			` AND sd.latitude IS NOT NULL AND sd.longitude IS NOT NULL AND
			  (6371 * acos(
				LEAST(1, GREATEST(-1,
					cos(radians($%d)) * cos(radians(sd.latitude::double precision)) *
					cos(radians(sd.longitude::double precision) - radians($%d)) +
					sin(radians($%d)) * sin(radians(sd.latitude::double precision))
				))
			  )) <= $%d`,
			latIdx,
			lonIdx,
			latIdx,
			radiusKmIdx,
		)
		orderBy = " ORDER BY sd.id, distance_km ASC NULLS LAST, sd.created_at DESC"

		args = append(args, reqData.Latitude, reqData.Longitude, reqData.Radius)
		paramIndex += 3
	} else if reqData.Pincode != nil {
		whereClause += fmt.Sprintf(` AND sd.pincode = $%d`, paramIndex)
		args = append(args, fmt.Sprintf("%d", *reqData.Pincode))
		paramIndex++
	}

	// Filter by department_id and/or category_id if provided (with AND condition)
	// department_id and category_id are varchar(32) strings (e.g. "dept_00001"), not integers.
	if reqData.DepartmentID != nil || reqData.CategoryID != nil {
		deptProvided := reqData.DepartmentID != nil && *reqData.DepartmentID != ""
		catProvided := reqData.CategoryID != nil && *reqData.CategoryID != ""

		if deptProvided && catProvided {
			// Both department_id and category_id provided - check for products matching BOTH
			whereClause += fmt.Sprintf(` AND sd.id IN (
				SELECT DISTINCT pi.shop_id FROM product_items pi
				WHERE pi.department_id = $%d AND pi.category_id = $%d
			)`, paramIndex, paramIndex+1)
			args = append(args, *reqData.DepartmentID, *reqData.CategoryID)
			paramIndex += 2
		} else if deptProvided {
			// Only department_id provided
			whereClause += fmt.Sprintf(` AND sd.id IN (
				SELECT DISTINCT pi.shop_id FROM product_items pi
				WHERE pi.department_id = $%d
			)`, paramIndex)
			args = append(args, *reqData.DepartmentID)
			paramIndex++
		} else if catProvided {
			// Only category_id provided
			whereClause += fmt.Sprintf(` AND sd.id IN (
				SELECT DISTINCT pi.shop_id FROM product_items pi
				WHERE pi.category_id = $%d
			)`, paramIndex)
			args = append(args, *reqData.CategoryID)
			paramIndex++
		}
	}

	query := fmt.Sprintf(`
		SELECT DISTINCT ON (sd.id)
			sd.id,
			sd.shop_name,
			sd.email,
			sd.phone,
			sd.latitude,
			sd.longitude,
			sd.owner_name,
			sd.shop_image_url,
			sd.address_line1,
			sd.address_line2,
			sd.city,
			sd.country,
			sd.state,
			sd.pincode,
			sd.shop_verification_status,
			`+IsSubscribedColumn+`,
			%s,
			sd.created_at,
			sd.updated_at
		FROM shop_details sd
		%s
		%s
		LIMIT $%d OFFSET $%d
	`, distanceExpr, whereClause, orderBy, paramIndex, paramIndex+1)

	args = append(args, reqData.Limit, reqData.Offset)
	err = c.DB.Raw(query, args...).Scan(&shops).Error
	if err != nil {
		return shops, err
	}

	// Batch-fetch reviews AND shop_times for all returned shops.
	if len(shops) > 0 {
		shopIDs := make([]string, len(shops))
		for i, s := range shops {
			shopIDs[i] = s.ID
		}

		// --- Reviews ---
		type reviewRow struct {
			ShopID    string    `gorm:"column:shop_id"`
			UserID    string    `gorm:"column:user_id"`
			Rating    uint      `gorm:"column:rating"`
			Review    string    `gorm:"column:review"`
			CreatedAt time.Time `gorm:"column:created_at"`
			UpdatedAt time.Time `gorm:"column:updated_at"`
		}
		var reviewRows []reviewRow
		c.DB.Raw(`
			SELECT shop_id, user_id, rating, review, created_at, updated_at
			FROM shop_socials
			WHERE shop_id IN (?) AND review IS NOT NULL AND TRIM(review) <> ''
			ORDER BY updated_at DESC
		`, shopIDs).Scan(&reviewRows)

		reviewMap := make(map[string][]response.ShopReview, len(shops))
		for _, r := range reviewRows {
			reviewMap[r.ShopID] = append(reviewMap[r.ShopID], response.ShopReview{
				UserID:    r.UserID,
				Rating:    r.Rating,
				Review:    r.Review,
				CreatedAt: r.CreatedAt,
				UpdatedAt: r.UpdatedAt,
			})
		}
		for i := range shops {
			if reviews, ok := reviewMap[shops[i].ID]; ok {
				shops[i].Reviews = reviews
			} else {
				shops[i].Reviews = []response.ShopReview{}
			}
		}

		// --- is_open: computed in Go using IST ---
		type shopTimeRow struct {
			ShopID    string   `gorm:"column:shop_id"`
			Status    string `gorm:"column:status"`
			OpenTime  string `gorm:"column:open_time"`
			CloseTime string `gorm:"column:close_time"`
		}
		var shopTimes []shopTimeRow
		c.DB.Raw(`
			SELECT DISTINCT ON (shop_id) shop_id, status, open_time, close_time
			FROM shop_times
			WHERE shop_id IN (?)
			ORDER BY shop_id, id DESC
		`, shopIDs).Scan(&shopTimes)

		istLoc, _ := time.LoadLocation("Asia/Kolkata")
		nowIST := time.Now().In(istLoc)
		nowStr := nowIST.Format("15:04") // HH:MM

		shopTimeMap := make(map[string]shopTimeRow, len(shopTimes))
		for _, st := range shopTimes {
			shopTimeMap[st.ShopID] = st
		}
		for i := range shops {
			st, ok := shopTimeMap[shops[i].ID]
			if !ok || st.Status != "open" {
				continue
			}
			// Normalise stored times to HH:MM
			openStr := st.OpenTime
			if len(openStr) > 5 {
				openStr = openStr[:5]
			}
			closeStr := st.CloseTime
			if len(closeStr) > 5 {
				closeStr = closeStr[:5]
			}
			shops[i].IsOpen = nowStr >= openStr && nowStr <= closeStr
		}
	}

	return shops, err
}

func (c *userDatabase) DeleteRefreshSessionByUserID(ctx context.Context, userId string, userType string) error {
	if userType == "admin" {
		query := `DELETE FROM admin_refresh_sessions WHERE user_id = $1 AND user_type = $2`
		err := c.DB.Exec(query, userId, userType).Error
		return err
	} else {
		query := `DELETE FROM user_refresh_sessions WHERE user_id = $1 AND user_type = $2`
		err := c.DB.Exec(query, userId, userType).Error
		return err
	}
}
func (c *userDatabase) FindShopByID(ctx context.Context, shopID string) (response.Shop, error) {

	var shop response.Shop
	query := `
		SELECT
			sd.id, sd.shop_name, sd.email, sd.phone,
			sd.address_line1, sd.address_line2, sd.city, sd.state, sd.country, sd.pincode,
			sd.shop_type, sd.shop_verification_status, sd.shop_image_url,
			sd.latitude, sd.longitude,
			sd.created_at, sd.updated_at,
			` + IsSubscribedColumn + `,
			CASE
				WHEN st.status = 'open' AND (CURRENT_TIMESTAMP AT TIME ZONE 'Asia/Kolkata')::time BETWEEN st.open_time::time AND st.close_time::time THEN true
				ELSE false
			END AS is_open
		FROM shop_details sd
		LEFT JOIN (
			SELECT DISTINCT ON (shop_id) *
			FROM shop_times
			ORDER BY shop_id, id DESC
		) st ON st.shop_id = sd.id
		WHERE sd.id = $1 AND sd.shop_status = 'active'`
	if c.DB.Raw(query, shopID).Scan(&shop).Error != nil {
		return shop, errors.New("failed to find shop by ID")
	}

	return shop, nil
}

func (c *userDatabase) UpdateTrialUsed(ctx context.Context, userID string) error {
	return c.DB.Exec("UPDATE users SET trial_used = true WHERE id = $1", userID).Error
}

func (c *userDatabase) IsTrialUsed(ctx context.Context, userID string) (bool, error) {
	var trialUsed bool
	err := c.DB.Raw("SELECT trial_used FROM users WHERE id = $1", userID).Scan(&trialUsed).Error
	return trialUsed, err
}
