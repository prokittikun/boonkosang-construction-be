package postgres

import (
	"boonkosang/internal/domain/models"
	"boonkosang/internal/repositories"
	"boonkosang/internal/requests"
	"context"
	"database/sql"
	"errors"
	"fmt"
	"strings"

	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
)

type supplierRepository struct {
	db *sqlx.DB
}

func NewSupplierRepository(db *sqlx.DB) repositories.SupplierRepository {
	return &supplierRepository{
		db: db,
	}
}

func (r *supplierRepository) Create(ctx context.Context, req requests.CreateSupplierRequest) (*models.Supplier, error) {
	supplier := &models.Supplier{
		SupplierID: uuid.New(),
		Name:       req.Name,
		Email:      req.Email,
		Tel:        req.Tel,
		Address:    req.Address,
	}

	query := `
	INSERT INTO Supplier (
	supplier_id, name, email,tel, address
	) VALUES (
	 :supplier_id, :name , :email, :tel, :address
	 ) RETURNING *
	`

	rows, err := r.db.NamedQueryContext(ctx, query, supplier)

	if err != nil {
		if strings.Contains(err.Error(), "unique constraint") {
			return nil, errors.New("supplier with this email already exists")
		}
		return nil, fmt.Errorf("filed to create supplier: %w", err)
	}
	defer rows.Close()

	if rows.Next() {
		err = rows.StructScan(supplier)
		if err != nil {
			return nil, fmt.Errorf("failed to scan supplier: %w", err)
		}
		return supplier, nil
	}
	return nil, errors.New("failed to create  supplier: no rows returned")
}

func (r *supplierRepository) Update(ctx context.Context, id uuid.UUID, req requests.UpdateSupplierRequest) error {
	query := `
        UPDATE Supplier SET 
            name = :name,
            email = :email,
            tel = :tel,
            address = :address,
        WHERE supplier_id = :supplier_id`

	params := map[string]interface{}{
		"supplier_id": id,
		"name":        req.Name,
		"email":       req.Email,
		"tel":         req.Tel,
		"address":     req.Address,
	}

	result, err := r.db.NamedExecContext(ctx, query, params)
	if err != nil {
		if strings.Contains(err.Error(), "unique constraint") {
			return errors.New("supplier with this email already exists")
		}
		return fmt.Errorf("failed to update supplier: %w", err)
	}

	rows, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get affected rows: %w", err)
	}

	if rows == 0 {
		return errors.New("supplier not found")
	}

	return nil
}
func (r *supplierRepository) Delete(ctx context.Context, id uuid.UUID) error {
	query := `DELETE FROM Supplier WHERE supplier_id = $1`

	result, err := r.db.ExecContext(ctx, query, id)
	if err != nil {
		return fmt.Errorf("failed to delete supplier: %w", err)
	}

	rows, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get affected rows: %w", err)
	}

	if rows == 0 {
		return errors.New("supplier not found")
	}

	return nil
}
func (r *supplierRepository) GetByID(ctx context.Context, id uuid.UUID) (*models.Supplier, error) {
	supplier := &models.Supplier{}
	query := `SELECT * FROM Supplier WHERE supplier_id = $1`

	err := r.db.GetContext(ctx, supplier, query, id)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, errors.New("supplier not found")
		}
		return nil, fmt.Errorf("failed to get supplier: %w", err)
	}

	return supplier, nil
}

func (r *supplierRepository) GetByEmail(ctx context.Context, email string) (*models.Supplier, error) {
	supplier := &models.Supplier{}
	query := `SELECT * FROM Supplier WHERE email = $1`

	err := r.db.GetContext(ctx, supplier, query, email)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, fmt.Errorf("failed to get supplier: %w", err)
	}

	return supplier, nil
}

func (r *supplierRepository) List(ctx context.Context, limit, offset int) ([]models.Supplier, int64, error) {
	var suppliers []models.Supplier
	var total int64

	countQuery := `SELECT COUNT(*) FROM Supplier`
	err := r.db.GetContext(ctx, &total, countQuery)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to get total count: %w", err)
	}

	query := `
        SELECT * FROM Supplier 
        LIMIT $1 OFFSET $2`

	err = r.db.SelectContext(ctx, &suppliers, query, limit, offset)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to list suppliers: %w", err)
	}

	return suppliers, total, nil
}
