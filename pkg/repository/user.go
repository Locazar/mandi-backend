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

func (c *userDatabase) FindUserByUserID(ctx context.Context, userID uint) (user domain.User, err error) {

	query := `SELECT * FROM users WHERE id = $1`
	err = c.DB.Raw(query, userID).Scan(&user).Error

	return user, err
}

func (c *userDatabase) FindUserByEmail(ctx context.Context, email string) (user domain.User, err error) {

	query := `SELECT * FROM users WHERE email = $1`

	fmt.Printf("Executing query: %s with email: %s\n", query, email) // Debugging line
	err = c.DB.Raw(query, email).Scan(&user).Error
	fmt.Printf("Query result: %+v, error: %v\n", user, err) // Debugging line

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
	fmt.Printf("Checking for existing user with Email: %s, Phone: %s excluding ID: %d\n",
		userDetails.Email, userDetails.Phone, userDetails.ID) // Debugging line

	query := `SELECT * FROM users WHERE (email = $1 OR phone = $2) AND id != $3`
	err = c.DB.Raw(query, userDetails.Email, userDetails.Phone, userDetails.ID).Scan(&user).Error

	fmt.Printf("Found user: %+v, error: %v\n", user, err) // Debugging line

	return
}

func (c *userDatabase) SaveUser(ctx context.Context, user domain.User) (userID uint, err error) {
	// Build dynamic column list and values
	columns := []string{}
	placeholders := []string{}
	values := []interface{}{}
	paramCount := 1

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

func (c *userDatabase) UpdateVerified(ctx context.Context, userID uint) error {

	query := `UPDATE users SET verified = 'T' WHERE id = $1`
	err := c.DB.Exec(query, userID).Error

	return err
}

func (c *userDatabase) UpdateAdminVerified(ctx context.Context, adminID uint) error {

	query := `UPDATE admins SET verified_seller = 'T' WHERE id = $1`
	err := c.DB.Exec(query, adminID).Error

	return err
}

func (c *userDatabase) UpdateUser(ctx context.Context, user domain.User) (err error) {

	updatedAt := time.Now()
	// check password need to update or not
	if user.Password != "" {
		query := `UPDATE users SET first_name = $1, last_name = $2,age = $3, 
		email = $4, phone = $5, password = $6, updated_at = $7 WHERE id = $8`
		err = c.DB.Exec(query, user.FirstName, user.LastName, user.Age, user.Email,
			user.Phone, user.Password, updatedAt, user.ID).Error
	} else {
		query := `UPDATE users SET first_name = $1, last_name = $2,age = $3, 
		email = $4, phone = $5,  updated_at = $6 WHERE id = $7`
		err = c.DB.Exec(query, user.FirstName, user.LastName, user.Age, user.Email,
			user.Phone, updatedAt, user.ID).Error
	}

	if err != nil {
		return fmt.Errorf("filed to update user detail of user with user_id %d", user.ID)
	}
	return nil
}

func (c *userDatabase) UpdateBlockStatus(ctx context.Context, userID uint, blockStatus bool) error {

	query := `UPDATE users SET block_status = $1 WHERE id = $2`
	err := c.DB.Exec(query, blockStatus, userID).Error

	return err
}

func (c *userDatabase) IsAddressIDExist(ctx context.Context, addressID uint) (exist bool, err error) {
	query := `SELECT EXISTS(SELECT 1 FROM addresses WHERE id = $1) AS exist FROM addresses`
	err = c.DB.Raw(query, addressID).Scan(&exist).Error

	return
}
func (c *userDatabase) FindAddressByID(ctx context.Context, addressID uint) (address response.Address, err error) {

	query := `SELECT adrs.id, adrs.address_line1, adrs.address_line2, adrs.phone_number, adrs.area, adrs.land_mark, 
	adrs.city, adrs.pincode, country_id, country_name, adrs.latitude, adrs.longitude FROM addresses adrs 
	INNER JOIN countries c ON c.id = adrs.country_id  
	INNER JOIN user_addresses uadrs ON uadrs.address_id = adrs.id 
	WHERE adrs.id = $1 `
	err = c.DB.Raw(query, addressID).Scan(&address).Error

	return
}

func (c *userDatabase) IsAddressAlreadyExistForUser(ctx context.Context, address domain.Address, userID uint) (exist bool, err error) {
	address.CountryID = 1 // hardcoded !!!! should change

	query := `SELECT DISTINCT CASE  WHEN adrs.id != 0 THEN 'T' ELSE 'F' END AS exist 
	FROM addresses adrs 
	INNER JOIN user_addresses urs ON adrs.id = urs.address_id 
	WHERE adrs.name = $1 AND adrs.house = $2 AND adrs.land_mark = $3 
	AND adrs.pincode = $4 AND adrs.country_id = $5  AND urs.user_id = $6`
	err = c.DB.Raw(query, address.AddressLine1, address.AddressLine2, address.LandMark, address.Pincode, address.CountryID, userID).Scan(&exist).Error
	if err != nil {
		return exist, fmt.Errorf("filed to check address already exist for user with user_id %d", userID)
	}
	return
}

func (c *userDatabase) FindAllAddressByUserID(ctx context.Context, userID uint) (addresses []response.Address, err error) {

	query := `SELECT a.id, a.house,a.name, a.phone_number, a.area, a.land_mark,a.city, 
	a.pincode, a.country_id, c.country_name, ua.is_default
	FROM user_addresses ua JOIN addresses a ON ua.address_id=a.id 
	INNER JOIN countries c ON a.country_id=c.id AND ua.user_id = $1`

	err = c.DB.Raw(query, userID).Scan(&addresses).Error

	return addresses, err
}

func (c *userDatabase) FindCountryByID(ctx context.Context, countryID uint) (domain.Country, error) {

	var country domain.Country

	fmt.Printf("Finding country with ID: %d\n", countryID) // Debugging line

	if c.DB.Raw("SELECT * FROM countries WHERE id = ?", countryID).Scan(&country).Error != nil {
		return country, errors.New("filed to find the country")
	}

	return country, nil
}

// save address
func (c *userDatabase) SaveAddress(ctx context.Context, address domain.Address) (addressID uint, err error) {
	address.CountryID = 1 // hardcoded !!!! should change
	query := `INSERT INTO addresses (user_id, area, land_mark, city, pincode, country_id, latitude, longitude, created_at, name, phone_number, house, address_line1, address_line2) 
	VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14) RETURNING id`

	createdAt := time.Now()

	if c.DB.Raw(query, address.UserID, address.Area, address.LandMark, address.City, address.Pincode, address.CountryID,
		address.Latitude, address.Longitude, createdAt, address.Name, address.PhoneNumber, address.House, address.AddressLine1, address.AddressLine2,
	).Scan(&addressID).Error != nil {
		return addressID, errors.New("failed to insert address on database")
	}
	return addressID, nil
}

// update address
func (c *userDatabase) UpdateAddress(ctx context.Context, address domain.Address) error {

	// address.CountryID = 1 // hardcoded !!!! should change
	query := `UPDATE addresses SET area=$1, land_mark=$2, city=$3, pincode=$4, country_id=$5, latitude=$6, longitude=$7, updated_at=$8, name=$9, phone_number=$10, house=$11, address_line1=$12, address_line2=$13 WHERE id=$14`

	updatedAt := time.Now()
	if c.DB.Raw(query, address.Area, address.LandMark, address.City, address.Pincode, address.CountryID,
		address.Latitude, address.Longitude, updatedAt, address.Name, address.PhoneNumber, address.House, address.AddressLine1, address.AddressLine2, address.ID).Scan(&address).Error != nil {
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

func (c *userDatabase) FindWishListItem(ctx context.Context, productID, userID uint) (domain.WishList, error) {

	var wishList domain.WishList
	query := `SELECT * FROM wish_lists WHERE user_id=? AND product_item_id=?`
	if c.DB.Raw(query, userID, productID).Scan(&wishList).Error != nil {
		return wishList, errors.New("filed to find wishlist item")
	}
	return wishList, nil
}

func (c *userDatabase) FindAllWishListItemsByUserID(ctx context.Context, userID uint) (productItems []response.WishListItem, err error) {

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

func (c *userDatabase) RemoveWishListItem(ctx context.Context, userID, productItemID uint) error {

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
		WHERE a.latitude IS NOT NULL AND a.longitude IS NOT NULL
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
		WHERE pincode = $1
		LIMIT $2 OFFSET $3
	`

	err = c.DB.Raw(query, reqData.Pincode, reqData.Limit, reqData.Offset).Scan(&sellers).Error

	return sellers, err
}

func (c *userDatabase) SearchShopList(ctx context.Context, reqData request.SearchShopListRequest) (shops []response.Shop, err error) {
	query := `
		SELECT id, shop_name, email, phone, latitude, longitude,
		owner_name, shop_image_url, address_line1, address_line2, city, country, state, pincode,
		shop_verification_status, created_at, updated_at
		FROM shop_details
		WHERE 1=1
	`

	paramIndex := 1
	args := []interface{}{}

	// Add search query condition if provided
	if reqData.Query != "" {
		query += fmt.Sprintf(` AND (shop_name ILIKE $%d OR owner_name ILIKE $%d)`, paramIndex, paramIndex)
		args = append(args, "%"+reqData.Query+"%")
		paramIndex++
	}

	// Filter by geolocation (lat + long + radius) OR pincode, but not both
	if reqData.Latitude != 0 && reqData.Longitude != 0 && reqData.Radius > 0 {
		// Using Haversine formula for distance calculation (in km)
		query += fmt.Sprintf(` AND latitude IS NOT NULL AND longitude IS NOT NULL AND (6371 * acos(cos(radians($%d)) * cos(radians(latitude)) * 
			cos(radians(longitude) - radians($%d)) + sin(radians($%d)) * 
			sin(radians(latitude)))) <= $%d`, paramIndex, paramIndex+1, paramIndex, paramIndex+2)
		args = append(args, reqData.Latitude, reqData.Longitude, reqData.Radius)
		paramIndex += 3
	} else if reqData.Pincode != nil {
		// Use pincode filter only if geolocation is not provided
		query += fmt.Sprintf(` AND pincode = $%d`, paramIndex)
		args = append(args, fmt.Sprintf("%d", *reqData.Pincode))
		paramIndex++
	}

	query += fmt.Sprintf(` LIMIT $%d OFFSET $%d`, paramIndex, paramIndex+1)
	args = append(args, reqData.Limit, reqData.Offset)

	err = c.DB.Raw(query, args...).Scan(&shops).Error

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
func (c *userDatabase) FindShopByID(ctx context.Context, shopID uint) (response.Shop, error) {

	var shop response.Shop
	query := `SELECT id, shop_name, email, phone, address_line1, address_line2, city, state, country, pincode,
	shop_type, shop_verification_status, shop_image_url, latitude, longitude, created_at, updated_at
	FROM shop_details WHERE id = $1`
	if c.DB.Raw(query, shopID).Scan(&shop).Error != nil {
		return shop, errors.New("failed to find shop by ID")
	}

	return shop, nil
}
