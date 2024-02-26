package services

import (
	"context"
	"go-grpc/cmd/helpers"
	paggingPb "go-grpc/pb/pagination"
	productPb "go-grpc/pb/product"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"gorm.io/gorm"

)

type ProductService struct {
	productPb.UnimplementedProductServiceServer
	DB *gorm.DB
}

func (p *ProductService) GetProducts(
	ctx context.Context,
	pageParam *productPb.Page,
) (*productPb.Products, error) {
	var (
		page       int64
		limit      int64
		pagination paggingPb.Pagination
		products   []*productPb.Product
	)

	if pageParam.GetPage() != 0 && pageParam.GetLimit() != 0 {
		page = pageParam.GetPage()
		limit = pageParam.GetLimit()
	}

	sql := p.DB.Table("products AS p").
		Joins("LEFT JOIN categories AS c ON c.id = p.category_id").
		Select("p.id", "p.name", "p.price", "p.stock", "c.id AS category_id", "c.name AS category_name")

	offset, limit := helpers.Pagination(sql, page, limit, &pagination)

	rows, err := sql.Offset(int(offset)).Limit(int(limit)).Rows()
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}
	defer rows.Close()

	for rows.Next() {
		var p productPb.Product
		var c productPb.Category

		if err := rows.Scan(
			&p.Id,
			&p.Name,
			&p.Price,
			&p.Stock,
			&c.Id,
			&c.Name,
		); err != nil {
			return nil, status.Error(codes.Internal, "Failed get products")
		}

		p.Category = &c
		products = append(products, &p)
	}

	res := &productPb.Products{
		Pagination: &pagination,
		Data:       products,
	}

	return res, nil
}

func (p *ProductService) GetProduct(
	ctx context.Context,
	id *productPb.Id,
) (*productPb.Product, error) {
	row := p.DB.Table("products AS p").
		Joins("LEFT JOIN categories AS c ON c.id = p.category_id").
		Select("p.id", "p.name", "p.price", "p.stock", "c.id AS category_id", "c.name AS category_name").
		Where("p.id = ?", id.GetId()).
		Row()

	var (
		product productPb.Product
		c       productPb.Category
	)

	if err := row.Scan(
		&product.Id,
		&product.Name,
		&product.Price,
		&product.Stock,
		&c.Id,
		&c.Name,
	); err != nil {
		return nil, status.Error(codes.NotFound, "Product not found")
	}

	product.Category = &c

	return &product, nil
}

func (p *ProductService) CreateProduct(
	ctx context.Context,
	productData *productPb.Product,
) (*productPb.Id, error) {
	var res productPb.Id

	err := p.DB.Transaction(func(tx *gorm.DB) error {
		category := productPb.Category{
			Id:   productData.GetCategory().GetId(),
			Name: productData.GetCategory().GetName(),
		}

		if err := tx.Table("categories").
			Where("LOWER(name) = LOWER(?)", category.GetName()).
			FirstOrCreate(&category).Error; err != nil {
			return err
		}

		product := struct {
			Id          uint64
			Name        string
			Price       float64
			Stock       uint32
			Category_id uint32
		}{
			Id:          productData.GetId(),
			Name:        productData.GetName(),
			Price:       productData.GetPrice(),
			Stock:       productData.GetStock(),
			Category_id: category.GetId(),
		}

		if err := tx.Table("products").Create(&product).Error; err != nil {
			return err
		}

		res.Id = product.Id
		return nil
	})
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	return &res, nil
}

func (p *ProductService) UpdateProduct(
	ctx context.Context,
	productData *productPb.Product,
) (*productPb.Status, error) {
	var res productPb.Status

	err := p.DB.Transaction(func(tx *gorm.DB) error {
		category := productPb.Category{
			Id:   productData.GetCategory().GetId(),
			Name: productData.GetCategory().GetName(),
		}

		if err := tx.Table("categories").
			Where("LOWER(name) = LOWER(?)", category.GetName()).
			FirstOrCreate(&category).Error; err != nil {
			return err
		}

		product := struct {
			Id          uint64
			Name        string
			Price       float64
			Stock       uint32
			Category_id uint32
		}{
			Id:          productData.GetId(),
			Name:        productData.GetName(),
			Price:       productData.GetPrice(),
			Stock:       productData.GetStock(),
			Category_id: category.GetId(),
		}

		if err := tx.Table("products").Where("id = ?", product.Id).Updates(&product).Error; err != nil {
			return err
		}

		res.Status = 1
		return nil
	})
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	return &res, nil
}

func (p *ProductService) DeleteProduct(
	ctx context.Context,
	id *productPb.Id,
) (*productPb.Status, error) {
	var res productPb.Status

	if err := p.DB.Table("products").
		Where("id = ?", id.GetId()).
		Delete(nil).Error; err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	res.Status = 1

	return &res, nil
}
