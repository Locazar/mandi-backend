package db

import (
	"errors"
	"log"

	"gorm.io/gorm"
)

func SetUpDBTriggers(db *gorm.DB) error {

	// Setup shop_departments table constraints
	if err := setupShopDepartmentsTable(db); err != nil {
		return errors.New("failed to setup shop_departments table: " + err.Error())
	}

	// Execute PostGIS setup commands first
	// for _, cmd := range postgisSetupCommands {
	// 	if db.Exec(cmd).Error != nil {
	// 		return errors.New("failed to execute postgis setup command: " + cmd)
	// 	}
	// }

	// first execute the trigger funtion
	if db.Exec(cartTotalPriceUpdateSqlFunc).Error != nil {
		return errors.New("failed to execute cart_total_price update trigger fun()")
	}

	// create trigger for calling the update funciton
	if db.Exec(cartTotalPriceTriggerExec).Error != nil {
		return errors.New("failed to create trigger for update total price on cart")
	}

	// update product_item qty on order time
	if db.Exec(orderProductUpdateOnPlaceOrder).Error != nil {
		return errors.New("failed to execute orderProductUpdate() trigger function")
	}

	if db.Exec(orderProductUpdateOnPlaceOrderTriggerExec).Error != nil {
		return errors.New("failed to create orderProductUpdateTriggerExec trigger")
	}

	//update product_item qty on order returned
	if db.Exec(orderReturnProductUpdate).Error != nil {
		return errors.New("failed to create orderReturnProductUpdate() trigger function")
	}

	if db.Exec(orderReturnProductUpdateExec).Error != nil {
		return errors.New("failed to create orderReturnProductUpdateExec trigger")
	}
	log.Printf("successfully triggers updated for database")
	return nil
}

// setupShopDepartmentsTable creates shop_departments table with proper constraints if it doesn't exist
func setupShopDepartmentsTable(db *gorm.DB) error {
	// Create table with foreign key constraint
	createTableSQL := `
	CREATE TABLE IF NOT EXISTS shop_departments (
		id SERIAL PRIMARY KEY,
		admin_id INTEGER NOT NULL,
		shop_id INTEGER NOT NULL,
		department_id INTEGER NOT NULL,
		category_id INTEGER NOT NULL,
		sub_category_id INTEGER NOT NULL,
		created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
		updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
		
		CONSTRAINT fk_shop_departments_shop_id 
			FOREIGN KEY (shop_id) 
			REFERENCES shop_details(id) 
			ON DELETE CASCADE,
		
		CONSTRAINT unique_shop_department 
			UNIQUE (shop_id, department_id)
	);
	`

	if err := db.Exec(createTableSQL).Error; err != nil {
		log.Printf("Note: shop_departments table SQL creation returned: %v (AutoMigrate may have already created it)", err)
	} else {
		log.Printf("✓ shop_departments table ensured with all constraints")
	}

	// Create indexes for foreign key
	createIndexSQL := `
	CREATE INDEX IF NOT EXISTS idx_shop_department_shop_id ON shop_departments(shop_id);
	CREATE INDEX IF NOT EXISTS idx_shop_department_admin_id ON shop_departments(admin_id);
	CREATE INDEX IF NOT EXISTS idx_shop_department_category_id ON shop_departments(category_id);
	CREATE INDEX IF NOT EXISTS idx_shop_department_sub_category_id ON shop_departments(sub_category_id);
	`

	if err := db.Exec(createIndexSQL).Error; err != nil {
		log.Printf("Warning: Could not create indexes on shop_departments: %v", err)
	}

	return nil
}

// var postgisSetupCommands = []string{
// 	"CREATE EXTENSION IF NOT EXISTS postgis;",
// 	"ALTER TABLE locations ADD COLUMN IF NOT EXISTS geog geography(Point, 4326);",
// 	"UPDATE locations SET geog = ST_MakePoint(longitude, latitude)::geography WHERE geog IS NULL;",
// 	"CREATE INDEX IF NOT EXISTS idx_locations_geog ON locations USING GIST(geog);",
// }

var (
	// function which return total price calculation on cart when product_item added or remove delete cart
	// in here checking  first it delete any row from cart_item then take its cart_id an find all cart_items with this id and calculate total price and update it
	// cart with this cart_id
	// any other like update or inset then take the cart_id and calculate the total price update on cart
	cartTotalPriceUpdateSqlFunc = `CREATE OR REPLACE FUNCTION update_cart_total_price() 
	RETURNS TRIGGER AS $$ 
	BEGIN 
	IF (TG_OP = 'DELETE') THEN
		UPDATE carts c
			SET total_price_amount_minor = (
				SELECT COALESCE ( SUM ( CASE WHEN pi.discount_price > 0 THEN pi.discount_price * ci.qty ELSE pi.price * ci.qty END), 0)::bigint
				FROM cart_items ci INNER JOIN product_items pi ON ci.product_item_id = pi.id
				WHERE ci.cart_id = OLD.cart_id
			), applied_coupon_id = 0, discount_amount_amount_minor = 0
		WHERE c.id = OLD.cart_id;
		RETURN NEW;
	ELSE
		UPDATE carts c
			SET total_price_amount_minor = (
				SELECT SUM (CASE WHEN pi.discount_price > 0 THEN pi.discount_price * ci.qty ELSE pi.price * ci.qty END)
				FROM cart_items ci INNER JOIN product_items pi ON ci.product_item_id = pi.id
				WHERE ci.cart_id = NEW.cart_id
			), applied_coupon_id = 0, discount_amount_amount_minor = 0
			WHERE c.id = NEW.cart_id;
	
	END IF; 
	RETURN NEW; 
	END; 
	$$ LANGUAGE plpgsql;`

	// for calling the trigger function above when an event of insert or update or delte happen on cart_items
	cartTotalPriceTriggerExec = `CREATE OR REPLACE TRIGGER  update_cart_total_price
	AFTER INSERT OR UPDATE OR DELETE ON cart_items
	FOR EACH ROW EXECUTE FUNCTION update_cart_total_price();`

	//for updating product_item quantity when order place
	orderProductUpdateOnPlaceOrder = `CREATE OR REPLACE FUNCTION update_product_quantity() 
	RETURNS TRIGGER AS $$ 
	BEGIN 
		IF (TG_OP = 'INSERT') THEN 
			UPDATE product_items pi 
			SET qty_in_stock = pi.qty_in_stock - NEW.qty 
			WHERE pi.id = NEW.product_item_id; 
	
		END IF; 
		RETURN NEW; 
	END; 
	$$ LANGUAGE plpgsql;`

	orderProductUpdateOnPlaceOrderTriggerExec = `CREATE OR REPLACE TRIGGER update_product_quantity 
	AFTER INSERT ON order_lines 
	FOR EACH ROW EXECUTE FUNCTION update_product_quantity();`

	//for order reuturn time product_item quantity update
	orderReturnProductUpdate = `CREATE OR REPLACE FUNCTION update_product_quantity_on_return()
	RETURNS TRIGGER AS $$
	BEGIN
	  IF (TG_OP = 'UPDATE') THEN 
		EXECUTE format('UPDATE product_items pi
						SET qty_in_stock = qty_in_stock + ol.qty
						FROM %I ol
						WHERE pi.id = ol.product_item_id
						AND ol.shop_order_id = $1.id',
						'order_lines')
		USING NEW;
	  
		RETURN NEW;
	  ELSE
		RETURN NULL;
	  END IF;
	END;
	$$ LANGUAGE plpgsql;`

	orderReturnProductUpdateExec = `CREATE OR REPLACE TRIGGER update_product_qty_on_order_return
	AFTER UPDATE OF status ON shop_orders
	FOR EACH ROW
	WHEN (NEW.status = 'order returned')
	EXECUTE FUNCTION update_product_quantity_on_return();`
)
