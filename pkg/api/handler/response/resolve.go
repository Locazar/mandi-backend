package response

import "github.com/rohit221990/mandi-backend/pkg/service/cloud"

// ResolveImages methods walk each image-bearing field and route it through
// cloud.ResolveURL so that:
//   - legacy "uploads/<dir>/<file>" rows stay served by local StaticFS, and
//   - bare object keys (post-backfill) resolve to Utho public URLs.
//
// Call once at the handler boundary, just before ctx.JSON(...). Promotion icon
// paths (IconPath fields on PromotionCategory / PromotionsType / Promotion)
// are intentionally NOT resolved here — promotion migration is deferred to a
// later release and those paths must continue to flow through unchanged.

// --- Product / ProductItems -------------------------------------------------

func (p *Product) ResolveImages(cs cloud.CloudService) {
	if p == nil {
		return
	}
	p.Image = cloud.ResolveURL(cs, p.Image)
	p.CategoryImageURL = cloud.ResolveURL(cs, p.CategoryImageURL)
	for i := range p.ProductItems {
		p.ProductItems[i].ResolveImages(cs)
	}
}

func (pi *ProductItems) ResolveImages(cs cloud.CloudService) {
	if pi == nil {
		return
	}
	pi.SubCategoryImageURL = cloud.ResolveURL(cs, pi.SubCategoryImageURL)
	for i, v := range pi.ProductItemImages {
		pi.ProductItemImages[i] = cloud.ResolveURL(cs, v)
	}
	for i := range pi.OfferProducts {
		pi.OfferProducts[i].ResolveImages(cs)
	}
}

func (op *OfferProduct) ResolveImages(cs cloud.CloudService) {
	if op == nil {
		return
	}
	op.Image = cloud.ResolveURL(cs, op.Image)
	op.Thumbnail = cloud.ResolveURL(cs, op.Thumbnail)
	// PromotionCategory + PromotionType IconPath intentionally left as-is.
}

// --- Category / SubCategory / Department ------------------------------------

func (c *Category) ResolveImages(cs cloud.CloudService) {
	if c == nil {
		return
	}
	c.ImageUrl = cloud.ResolveURL(cs, c.ImageUrl)
	// Icon left as-is (bundled-asset path; promotions deferred).
}

func (sc *SubCategory) ResolveImages(cs cloud.CloudService) {
	if sc == nil {
		return
	}
	sc.ImageUrl = cloud.ResolveURL(cs, sc.ImageUrl)
}

func (d *Department) ResolveImages(cs cloud.CloudService) {
	if d == nil {
		return
	}
	d.ImageUrl = cloud.ResolveURL(cs, d.ImageUrl)
	// Icon left as-is.
}

// --- Offer ------------------------------------------------------------------

func (o *Offer) ResolveImages(cs cloud.CloudService) {
	if o == nil {
		return
	}
	o.Image = cloud.ResolveURL(cs, o.Image)
	o.Thumbnail = cloud.ResolveURL(cs, o.Thumbnail)
}

// --- Shop / Banner ----------------------------------------------------------

func (s *Shop) ResolveImages(cs cloud.CloudService) {
	if s == nil {
		return
	}
	s.ShopImageURL = cloud.ResolveURL(cs, s.ShopImageURL)
}

func (b *Banner) ResolveImages(cs cloud.CloudService) {
	if b == nil {
		return
	}
	b.ImageURL = cloud.ResolveURL(cs, b.ImageURL)
}

// --- Slice helpers (handlers usually return slices) -------------------------

func ResolveProductItemsImages(cs cloud.CloudService, items []ProductItems) {
	for i := range items {
		items[i].ResolveImages(cs)
	}
}

func ResolveProductsImages(cs cloud.CloudService, products []Product) {
	for i := range products {
		products[i].ResolveImages(cs)
	}
}

func ResolveCategoriesImages(cs cloud.CloudService, categories []Category) {
	for i := range categories {
		categories[i].ResolveImages(cs)
	}
}

func ResolveSubCategoriesImages(cs cloud.CloudService, subs []SubCategory) {
	for i := range subs {
		subs[i].ResolveImages(cs)
	}
}

func ResolveDepartmentsImages(cs cloud.CloudService, depts []Department) {
	for i := range depts {
		depts[i].ResolveImages(cs)
	}
}

func ResolveOffersImages(cs cloud.CloudService, offers []Offer) {
	for i := range offers {
		offers[i].ResolveImages(cs)
	}
}

func ResolveShopsImages(cs cloud.CloudService, shops []Shop) {
	for i := range shops {
		shops[i].ResolveImages(cs)
	}
}

func ResolveBannersImages(cs cloud.CloudService, banners []Banner) {
	for i := range banners {
		banners[i].ResolveImages(cs)
	}
}

// --- Search results (used by global search + autocomplete) -----------------

func (p *ProductSearchItem) ResolveImages(cs cloud.CloudService) {
	if p == nil {
		return
	}
	p.ImageURL = cloud.ResolveURL(cs, p.ImageURL)
	for i, v := range p.Images {
		p.Images[i] = cloud.ResolveURL(cs, v)
	}
}

func (c *CategorySearchItem) ResolveImages(cs cloud.CloudService) {
	if c == nil {
		return
	}
	c.ImageURL = cloud.ResolveURL(cs, c.ImageURL)
}

func (s *ShopSearchItem) ResolveImages(cs cloud.CloudService) {
	if s == nil {
		return
	}
	s.ShopImageURL = cloud.ResolveURL(cs, s.ShopImageURL)
}

func (d *DepartmentSearchItem) ResolveImages(cs cloud.CloudService) {
	if d == nil {
		return
	}
	d.ImageURL = cloud.ResolveURL(cs, d.ImageURL)
}

func (o *OfferSearchItem) ResolveImages(cs cloud.CloudService) {
	if o == nil {
		return
	}
	o.Image = cloud.ResolveURL(cs, o.Image)
}

func (a *AutocompleteSuggestion) ResolveImages(cs cloud.CloudService) {
	if a == nil {
		return
	}
	a.ImageURL = cloud.ResolveURL(cs, a.ImageURL)
}

func (r *GlobalSearchResult) ResolveImages(cs cloud.CloudService) {
	if r == nil {
		return
	}
	for i := range r.Products {
		r.Products[i].ResolveImages(cs)
	}
	for i := range r.Categories {
		r.Categories[i].ResolveImages(cs)
	}
	for i := range r.Shops {
		r.Shops[i].ResolveImages(cs)
	}
	for i := range r.Departments {
		r.Departments[i].ResolveImages(cs)
	}
	for i := range r.Offers {
		r.Offers[i].ResolveImages(cs)
	}
}

func ResolveAutocompleteImages(cs cloud.CloudService, suggestions []AutocompleteSuggestion) {
	for i := range suggestions {
		suggestions[i].ResolveImages(cs)
	}
}

// --- Order items ------------------------------------------------------------

func (oi *OrderItem) ResolveImages(cs cloud.CloudService) {
	if oi == nil {
		return
	}
	oi.Image = cloud.ResolveURL(cs, oi.Image)
}

func ResolveOrderItemsImages(cs cloud.CloudService, items []OrderItem) {
	for i := range items {
		items[i].ResolveImages(cs)
	}
}
