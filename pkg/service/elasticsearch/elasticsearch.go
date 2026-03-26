package elasticsearch

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"log"
	"strconv"

	"github.com/elastic/go-elasticsearch/v8"
	"github.com/rohit221990/mandi-backend/pkg/domain"
)

type ElasticService struct {
	Client *elasticsearch.Client
}

func (es *ElasticService) UpdateProductItem(ctx context.Context, domainItem domain.ProductItem) error {
	// Re-index with the same document ID to update
	body := map[string]interface{}{
		"id":                  domainItem.ID,
		"sub_category_name":   domainItem.SubCategoryName,
		"category_id":         domainItem.CategoryID,
		"department_id":       domainItem.DepartmentID,
		"sub_category_id":     domainItem.SubCategoryID,
		"admin_id":            domainItem.AdminID,
		"dynamic_fields":      domainItem.DynamicFields,
		"product_item_images": domainItem.ProductItemImages,
	}
	data, err := json.Marshal(body)
	if err != nil {
		log.Printf("Error marshaling product item for update: %v", err)
		return err
	}

	res, err := es.Client.Index(
		"product_items",
		bytes.NewReader(data),
		es.Client.Index.WithDocumentID(fmt.Sprintf("%d", domainItem.ID)),
		es.Client.Index.WithContext(ctx),
	)
	if err != nil {
		log.Printf("Error updating product item in Elasticsearch: %v", err)
		return err
	}
	defer res.Body.Close()

	if res.IsError() {
		log.Printf("Error response when updating product item: %s", res.String())
		return fmt.Errorf("error updating product item: %s", res.String())
	}

	log.Printf("Product item %d updated successfully in Elasticsearch", domainItem.ID)
	return nil
}

func NewElasticService(url string) (*ElasticService, error) {
	cfg := elasticsearch.Config{
		Addresses: []string{url},
	}
	client, err := elasticsearch.NewClient(cfg)
	if err != nil {
		return nil, err
	}
	return &ElasticService{Client: client}, nil
}

// IndexProduct indexes a product in Elasticsearch
func (es *ElasticService) IndexProduct(ctx context.Context, product domain.Product) error {
	body := map[string]interface{}{
		"id":          product.ID,
		"name":        product.Name,
		"description": product.Description,
		"category_id": product.CategoryID,
	}
	data, err := json.Marshal(body)
	if err != nil {
		return err
	}

	res, err := es.Client.Index(
		"products",
		bytes.NewReader(data),
		es.Client.Index.WithDocumentID(fmt.Sprintf("%d", product.ID)),
		es.Client.Index.WithContext(ctx),
	)
	if err != nil {
		return err
	}
	defer res.Body.Close()

	if res.IsError() {
		return fmt.Errorf("error indexing product: %s", res.String())
	}

	log.Printf("Product %d indexed successfully", product.ID)
	return nil
}

// IndexProductItem indexes a product item in Elasticsearch
func (es *ElasticService) IndexProductItem(ctx context.Context, productItem domain.ProductItem) error {
	body := map[string]interface{}{
		"id":                  productItem.ID,
		"sub_category_name":   productItem.SubCategoryName,
		"category_id":         productItem.CategoryID,
		"department_id":       productItem.DepartmentID,
		"sub_category_id":     productItem.SubCategoryID,
		"admin_id":            productItem.AdminID,
		"dynamic_fields":      productItem.DynamicFields,
		"product_item_images": productItem.ProductItemImages,
	}
	data, err := json.Marshal(body)
	if err != nil {
		return err
	}

	res, err := es.Client.Index(
		"product_items",
		bytes.NewReader(data),
		es.Client.Index.WithDocumentID(fmt.Sprintf("%d", productItem.ID)),
		es.Client.Index.WithContext(ctx),
	)
	if err != nil {
		return err
	}
	defer res.Body.Close()

	if res.IsError() {
		return fmt.Errorf("error indexing product item: %s", res.String())
	}

	log.Printf("Product item %d indexed successfully", productItem.ID)
	return nil
}

// SearchProducts searches for products in Elasticsearch with filters
func (es *ElasticService) SearchProducts(ctx context.Context, keyword string, categoryID *string, brandID *string, minPrice, maxPrice *float64, limit, offset int) ([]domain.Product, error) {
	query := map[string]interface{}{
		"query": map[string]interface{}{
			"bool": map[string]interface{}{
				"must": []map[string]interface{}{
					{
						"multi_match": map[string]interface{}{
							"query":  keyword,
							"fields": []string{"name^2", "description"}, // Boost name
						},
					},
				},
				"filter": []map[string]interface{}{},
			},
		},
		"sort": []map[string]interface{}{
			{"_score": map[string]string{"order": "desc"}},
			{"price": map[string]string{"order": "asc"}},
		},
		"size": limit,
		"from": offset,
	}

	filters := query["query"].(map[string]interface{})["bool"].(map[string]interface{})["filter"].([]map[string]interface{})

	if categoryID != nil {
		filters = append(filters, map[string]interface{}{
			"term": map[string]interface{}{
				"category_id": *categoryID,
			},
		})
	}
	if brandID != nil {
		filters = append(filters, map[string]interface{}{
			"term": map[string]interface{}{
				"brand_id": *brandID,
			},
		})
	}
	if minPrice != nil || maxPrice != nil {
		rangeFilter := map[string]interface{}{
			"range": map[string]interface{}{
				"price": map[string]interface{}{},
			},
		}
		if minPrice != nil {
			rangeFilter["range"].(map[string]interface{})["price"].(map[string]interface{})["gte"] = *minPrice
		}
		if maxPrice != nil {
			rangeFilter["range"].(map[string]interface{})["price"].(map[string]interface{})["lte"] = *maxPrice
		}
		filters = append(filters, rangeFilter)
	}

	data, err := json.Marshal(query)
	if err != nil {
		return nil, err
	}

	res, err := es.Client.Search(
		es.Client.Search.WithContext(ctx),
		es.Client.Search.WithIndex("products"),
		es.Client.Search.WithBody(bytes.NewReader(data)),
	)
	if err != nil {
		return nil, err
	}
	defer res.Body.Close()

	if res.IsError() {
		return nil, fmt.Errorf("error searching products: %s", res.String())
	}

	var result map[string]interface{}
	if err := json.NewDecoder(res.Body).Decode(&result); err != nil {
		return nil, err
	}

	hits := result["hits"].(map[string]interface{})["hits"].([]interface{})
	products := make([]domain.Product, 0, len(hits))
	for _, hit := range hits {
		source := hit.(map[string]interface{})["_source"].(map[string]interface{})
		product := domain.Product{
			ID:          uint(source["id"].(float64)),
			Name:        source["name"].(string),
			Description: source["description"].(string),
			CategoryID:  uint(source["category_id"].(float64))}
		products = append(products, product)
	}

	return products, nil
}

// SearchProductItems searches for product items in Elasticsearch and returns IDs
func (es *ElasticService) SearchProductItems(ctx context.Context, keyword string, categoryID *string, shopID *string, limit, offset int) ([]uint, error) {
	query := map[string]interface{}{
		"query": map[string]interface{}{
			"bool": map[string]interface{}{
				"must": []map[string]interface{}{
					{
						"multi_match": map[string]interface{}{
							"query":  keyword,
							"fields": []string{"sub_category_name", "dynamic_fields"},
						},
					},
				},
				"filter": []map[string]interface{}{},
			},
		},
		"sort": []map[string]interface{}{
			{"_score": map[string]string{"order": "desc"}},
		},
		"size": limit,
		"from": offset,
	}

	filters := query["query"].(map[string]interface{})["bool"].(map[string]interface{})["filter"].([]map[string]interface{})

	if categoryID != nil {
		if cid, err := strconv.ParseUint(*categoryID, 10, 64); err == nil {
			filters = append(filters, map[string]interface{}{
				"term": map[string]interface{}{
					"category_id": cid,
				},
			})
		}
	}

	if shopID != nil {
		if sid, err := strconv.ParseUint(*shopID, 10, 64); err == nil {
			filters = append(filters, map[string]interface{}{
				"term": map[string]interface{}{
					"shop_id": sid,
				},
			})
		}
	}

	data, err := json.Marshal(query)
	if err != nil {
		return nil, err
	}

	res, err := es.Client.Search(
		es.Client.Search.WithContext(ctx),
		es.Client.Search.WithIndex("product_items"),
		es.Client.Search.WithBody(bytes.NewReader(data)),
	)
	if err != nil {
		return nil, err
	}
	defer res.Body.Close()

	if res.IsError() {
		return nil, fmt.Errorf("error searching product items: %s", res.String())
	}

	var result map[string]interface{}
	if err := json.NewDecoder(res.Body).Decode(&result); err != nil {
		return nil, err
	}

	hits := result["hits"].(map[string]interface{})["hits"].([]interface{})
	ids := make([]uint, 0, len(hits))
	for _, hit := range hits {
		source := hit.(map[string]interface{})["_source"].(map[string]interface{})
		id := uint(source["id"].(float64))
		ids = append(ids, id)
	}

	return ids, nil
}

// IndexDepartment indexes a department in Elasticsearch
func (es *ElasticService) IndexDepartment(ctx context.Context, dept domain.Department) error {
	body := map[string]interface{}{
		"id":   dept.ID,
		"name": dept.Name,
	}
	data, err := json.Marshal(body)
	if err != nil {
		return err
	}

	res, err := es.Client.Index(
		"departments",
		bytes.NewReader(data),
		es.Client.Index.WithDocumentID(fmt.Sprintf("%d", dept.ID)),
		es.Client.Index.WithContext(ctx),
	)
	if err != nil {
		return err
	}
	defer res.Body.Close()

	if res.IsError() {
		return fmt.Errorf("error indexing department: %s", res.String())
	}

	return nil
}

// SearchDepartments searches for departments in Elasticsearch
func (es *ElasticService) SearchDepartments(ctx context.Context, query string, limit, offset int) ([]domain.Department, error) {
	searchQuery := map[string]interface{}{
		"query": map[string]interface{}{
			"multi_match": map[string]interface{}{
				"query":  query,
				"fields": []string{"name"},
			},
		},
		"size": limit,
		"from": offset,
	}
	data, err := json.Marshal(searchQuery)
	if err != nil {
		return nil, err
	}

	res, err := es.Client.Search(
		es.Client.Search.WithContext(ctx),
		es.Client.Search.WithIndex("departments"),
		es.Client.Search.WithBody(bytes.NewReader(data)),
	)
	if err != nil {
		return nil, err
	}
	defer res.Body.Close()

	if res.IsError() {
		return nil, fmt.Errorf("error searching departments: %s", res.String())
	}

	var result map[string]interface{}
	if err := json.NewDecoder(res.Body).Decode(&result); err != nil {
		return nil, err
	}

	hits := result["hits"].(map[string]interface{})["hits"].([]interface{})
	departments := make([]domain.Department, 0, len(hits))
	for _, hit := range hits {
		source := hit.(map[string]interface{})["_source"].(map[string]interface{})
		dept := domain.Department{
			ID:   uint(source["id"].(float64)),
			Name: source["name"].(string),
		}
		departments = append(departments, dept)
	}

	return departments, nil
}

// IndexCategory indexes a category in Elasticsearch
func (es *ElasticService) IndexCategory(ctx context.Context, cat domain.Category) error {
	body := map[string]interface{}{
		"id":            cat.ID,
		"name":          cat.Name,
		"department_id": cat.DepartmentID,
	}
	data, err := json.Marshal(body)
	if err != nil {
		return err
	}

	res, err := es.Client.Index(
		"categories",
		bytes.NewReader(data),
		es.Client.Index.WithDocumentID(fmt.Sprintf("%d", cat.ID)),
		es.Client.Index.WithContext(ctx),
	)
	if err != nil {
		return err
	}
	defer res.Body.Close()

	if res.IsError() {
		return fmt.Errorf("error indexing category: %s", res.String())
	}

	return nil
}

// SearchCategories searches for categories in Elasticsearch
func (es *ElasticService) SearchCategories(ctx context.Context, query string, departmentID *uint, limit, offset int) ([]domain.Category, error) {
	queryMap := map[string]interface{}{
		"query": map[string]interface{}{
			"bool": map[string]interface{}{
				"must": map[string]interface{}{
					"multi_match": map[string]interface{}{
						"query":  query,
						"fields": []string{"name"},
					},
				},
			},
		},
		"size": limit,
		"from": offset,
	}

	if departmentID != nil {
		queryMap["query"].(map[string]interface{})["bool"].(map[string]interface{})["filter"] = []map[string]interface{}{
			{
				"term": map[string]interface{}{
					"department_id": *departmentID,
				},
			},
		}
	}

	data, err := json.Marshal(queryMap)
	if err != nil {
		return nil, err
	}

	res, err := es.Client.Search(
		es.Client.Search.WithContext(ctx),
		es.Client.Search.WithIndex("categories"),
		es.Client.Search.WithBody(bytes.NewReader(data)),
	)
	if err != nil {
		return nil, err
	}
	defer res.Body.Close()

	if res.IsError() {
		return nil, fmt.Errorf("error searching categories: %s", res.String())
	}

	var result map[string]interface{}
	if err := json.NewDecoder(res.Body).Decode(&result); err != nil {
		return nil, err
	}

	hits := result["hits"].(map[string]interface{})["hits"].([]interface{})
	categories := make([]domain.Category, 0, len(hits))
	for _, hit := range hits {
		source := hit.(map[string]interface{})["_source"].(map[string]interface{})
		cat := domain.Category{
			ID:           uint(source["id"].(float64)),
			Name:         source["name"].(string),
			DepartmentID: uint(source["department_id"].(float64)),
		}
		categories = append(categories, cat)
	}

	return categories, nil
}

// IndexBrand indexes a brand in Elasticsearch
func (es *ElasticService) IndexBrand(ctx context.Context, brand domain.Brand) error {
	body := map[string]interface{}{
		"id":   brand.ID,
		"name": brand.Name,
	}
	data, err := json.Marshal(body)
	if err != nil {
		return err
	}

	res, err := es.Client.Index(
		"brands",
		bytes.NewReader(data),
		es.Client.Index.WithDocumentID(fmt.Sprintf("%d", brand.ID)),
		es.Client.Index.WithContext(ctx),
	)
	if err != nil {
		return err
	}
	defer res.Body.Close()

	if res.IsError() {
		return fmt.Errorf("error indexing brand: %s", res.String())
	}

	return nil
}

// SearchBrands searches for brands in Elasticsearch
func (es *ElasticService) SearchBrands(ctx context.Context, query string, limit, offset int) ([]domain.Brand, error) {
	searchQuery := map[string]interface{}{
		"query": map[string]interface{}{
			"multi_match": map[string]interface{}{
				"query":  query,
				"fields": []string{"name"},
			},
		},
		"size": limit,
		"from": offset,
	}
	data, err := json.Marshal(searchQuery)
	if err != nil {
		return nil, err
	}

	res, err := es.Client.Search(
		es.Client.Search.WithContext(ctx),
		es.Client.Search.WithIndex("brands"),
		es.Client.Search.WithBody(bytes.NewReader(data)),
	)
	if err != nil {
		return nil, err
	}
	defer res.Body.Close()

	if res.IsError() {
		return nil, fmt.Errorf("error searching brands: %s", res.String())
	}

	var result map[string]interface{}
	if err := json.NewDecoder(res.Body).Decode(&result); err != nil {
		return nil, err
	}

	hits := result["hits"].(map[string]interface{})["hits"].([]interface{})
	brands := make([]domain.Brand, 0, len(hits))
	for _, hit := range hits {
		source := hit.(map[string]interface{})["_source"].(map[string]interface{})
		brand := domain.Brand{
			ID:   uint(source["id"].(float64)),
			Name: source["name"].(string),
		}
		brands = append(brands, brand)
	}

	return brands, nil
}

// IndexOffer indexes an offer in Elasticsearch
func (es *ElasticService) IndexOffer(ctx context.Context, offer domain.Offer) error {
	body := map[string]interface{}{
		"id":          offer.ID,
		"name":        offer.Name,
		"description": offer.Description,
		"price":       offer.DiscountRate, // or relevant field
	}
	data, err := json.Marshal(body)
	if err != nil {
		return err
	}

	res, err := es.Client.Index(
		"offers",
		bytes.NewReader(data),
		es.Client.Index.WithDocumentID(fmt.Sprintf("%d", offer.ID)),
		es.Client.Index.WithContext(ctx),
	)
	if err != nil {
		return err
	}
	defer res.Body.Close()

	if res.IsError() {
		return fmt.Errorf("error indexing offer: %s", res.String())
	}

	return nil
}

// SearchOffers searches for offers in Elasticsearch
func (es *ElasticService) SearchOffers(ctx context.Context, query string, limit, offset int) ([]domain.Offer, error) {
	searchQuery := map[string]interface{}{
		"query": map[string]interface{}{
			"multi_match": map[string]interface{}{
				"query":  query,
				"fields": []string{"name", "description"},
			},
		},
		"size": limit,
		"from": offset,
	}
	data, err := json.Marshal(searchQuery)
	if err != nil {
		return nil, err
	}

	res, err := es.Client.Search(
		es.Client.Search.WithContext(ctx),
		es.Client.Search.WithIndex("offers"),
		es.Client.Search.WithBody(bytes.NewReader(data)),
	)
	if err != nil {
		return nil, err
	}
	defer res.Body.Close()

	if res.IsError() {
		return nil, fmt.Errorf("error searching offers: %s", res.String())
	}

	var result map[string]interface{}
	if err := json.NewDecoder(res.Body).Decode(&result); err != nil {
		return nil, err
	}

	hits := result["hits"].(map[string]interface{})["hits"].([]interface{})
	offers := make([]domain.Offer, 0, len(hits))
	for _, hit := range hits {
		source := hit.(map[string]interface{})["_source"].(map[string]interface{})
		offer := domain.Offer{
			ID:          uint(source["id"].(float64)),
			Name:        source["name"].(string),
			Description: source["description"].(string),
		}
		offers = append(offers, offer)
	}

	return offers, nil
}

// IndexUser indexes a user in Elasticsearch
func (es *ElasticService) IndexUser(ctx context.Context, user domain.User) error {
	body := map[string]interface{}{
		"id":         user.ID,
		"first_name": user.FirstName,
		"last_name":  user.LastName,
		"email":      user.Email,
		"phone":      user.Phone,
	}
	data, err := json.Marshal(body)
	if err != nil {
		return err
	}

	res, err := es.Client.Index(
		"users",
		bytes.NewReader(data),
		es.Client.Index.WithDocumentID(fmt.Sprintf("%d", user.ID)),
		es.Client.Index.WithContext(ctx),
	)
	if err != nil {
		return err
	}
	defer res.Body.Close()

	if res.IsError() {
		return fmt.Errorf("error indexing user: %s", res.String())
	}

	return nil
}

// SearchUsers searches for users in Elasticsearch
func (es *ElasticService) SearchUsers(ctx context.Context, query string, limit, offset int) ([]domain.User, error) {
	searchQuery := map[string]interface{}{
		"query": map[string]interface{}{
			"multi_match": map[string]interface{}{
				"query":  query,
				"fields": []string{"first_name", "last_name", "email", "phone"},
			},
		},
		"size": limit,
		"from": offset,
	}
	data, err := json.Marshal(searchQuery)
	if err != nil {
		return nil, err
	}

	res, err := es.Client.Search(
		es.Client.Search.WithContext(ctx),
		es.Client.Search.WithIndex("users"),
		es.Client.Search.WithBody(bytes.NewReader(data)),
	)
	if err != nil {
		return nil, err
	}
	defer res.Body.Close()

	if res.IsError() {
		return nil, fmt.Errorf("error searching users: %s", res.String())
	}

	var result map[string]interface{}
	if err := json.NewDecoder(res.Body).Decode(&result); err != nil {
		return nil, err
	}

	hits := result["hits"].(map[string]interface{})["hits"].([]interface{})
	users := make([]domain.User, 0, len(hits))
	for _, hit := range hits {
		source := hit.(map[string]interface{})["_source"].(map[string]interface{})
		user := domain.User{
			ID:        uint(source["id"].(float64)),
			FirstName: source["first_name"].(string),
			LastName:  source["last_name"].(string),
			Email:     source["email"].(string),
			Phone:     source["phone"].(string),
		}
		users = append(users, user)
	}

	return users, nil
}

// IndexAdmin indexes an admin in Elasticsearch
func (es *ElasticService) IndexAdmin(ctx context.Context, admin domain.Admin) error {
	body := map[string]interface{}{
		"id":        admin.ID,
		"full_name": admin.FullName,
		"email":     admin.Email,
		"mobile":    admin.Mobile,
	}
	data, err := json.Marshal(body)
	if err != nil {
		return err
	}

	res, err := es.Client.Index(
		"admins",
		bytes.NewReader(data),
		es.Client.Index.WithDocumentID(fmt.Sprintf("%d", admin.ID)),
		es.Client.Index.WithContext(ctx),
	)
	if err != nil {
		return err
	}
	defer res.Body.Close()

	if res.IsError() {
		return fmt.Errorf("error indexing admin: %s", res.String())
	}

	return nil
}

// SearchAdmins searches for admins in Elasticsearch
func (es *ElasticService) SearchAdmins(ctx context.Context, query string, limit, offset int) ([]domain.Admin, error) {
	searchQuery := map[string]interface{}{
		"query": map[string]interface{}{
			"multi_match": map[string]interface{}{
				"query":  query,
				"fields": []string{"full_name", "email", "mobile"},
			},
		},
		"size": limit,
		"from": offset,
	}
	data, err := json.Marshal(searchQuery)
	if err != nil {
		return nil, err
	}

	res, err := es.Client.Search(
		es.Client.Search.WithContext(ctx),
		es.Client.Search.WithIndex("admins"),
		es.Client.Search.WithBody(bytes.NewReader(data)),
	)
	if err != nil {
		return nil, err
	}
	defer res.Body.Close()

	if res.IsError() {
		return nil, fmt.Errorf("error searching admins: %s", res.String())
	}

	var result map[string]interface{}
	if err := json.NewDecoder(res.Body).Decode(&result); err != nil {
		return nil, err
	}

	hits := result["hits"].(map[string]interface{})["hits"].([]interface{})
	admins := make([]domain.Admin, 0, len(hits))
	for _, hit := range hits {
		source := hit.(map[string]interface{})["_source"].(map[string]interface{})
		admin := domain.Admin{
			ID:       uint(source["id"].(float64)),
			FullName: source["full_name"].(string),
			Email:    source["email"].(string),
			Mobile:   source["mobile"].(string),
		}
		admins = append(admins, admin)
	}

	return admins, nil
}

// BulkIndexProducts indexes multiple products
func (es *ElasticService) BulkIndexProducts(ctx context.Context, products []domain.Product) error {
	var body bytes.Buffer
	for _, product := range products {
		meta := map[string]interface{}{
			"index": map[string]interface{}{
				"_index": "products",
				"_id":    fmt.Sprintf("%d", product.ID),
			},
		}
		metaData, _ := json.Marshal(meta)
		body.WriteString(string(metaData) + "\n")

		doc := map[string]interface{}{
			"id":            product.ID,
			"name":          product.Name,
			"description":   product.Description,
			"category_id":   product.CategoryID,
			"department_id": product.DepartmentID,
			"shop_id":       product.ShopID,
			"image":         product.Image,
		}
		docData, _ := json.Marshal(doc)
		body.WriteString(string(docData) + "\n")
	}

	res, err := es.Client.Bulk(bytes.NewReader(body.Bytes()), es.Client.Bulk.WithIndex("products"))
	if err != nil {
		return err
	}
	defer res.Body.Close()

	if res.IsError() {
		return fmt.Errorf("error bulk indexing products: %s", res.String())
	}

	log.Printf("Bulk indexed %d products", len(products))
	return nil
}

// Similarly for others, but abbreviated for brevity
// BulkIndexCategories, BulkIndexBrands, BulkIndexOffers, BulkIndexUsers, BulkIndexAdmins can be added similarly
