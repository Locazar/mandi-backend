package repositories

import (
    "context"
    "time"

    "github.com/rohit221990/mandi-backend/internal/models"
    "gorm.io/gorm"
)

type UserRepository struct {
    DB *gorm.DB
}

func NewUserRepository(db *gorm.DB) *UserRepository { return &UserRepository{DB: db} }

func (r *UserRepository) Create(ctx context.Context, u *models.User) error {
    return r.DB.WithContext(ctx).Create(u).Error
}

func (r *UserRepository) FindByEmail(ctx context.Context, email string) (*models.User, error) {
    var u models.User
    if err := r.DB.WithContext(ctx).Where("email = ?", email).First(&u).Error; err != nil {
        return nil, err
    }
    return &u, nil
}

func (r *UserRepository) FindByPhone(ctx context.Context, phone string) (*models.User, error) {
    var u models.User
    if err := r.DB.WithContext(ctx).Where("phone = ?", phone).First(&u).Error; err != nil {
        return nil, err
    }
    return &u, nil
}

func (r *UserRepository) Update(ctx context.Context, u *models.User) error {
    u.UpdatedAt = time.Now()
    return r.DB.WithContext(ctx).Save(u).Error
}

// OTP repository
type OTPRepository struct {
    DB *gorm.DB
}

func NewOTPRepository(db *gorm.DB) *OTPRepository { return &OTPRepository{DB: db} }

func (r *OTPRepository) Create(ctx context.Context, o *models.OTPRequest) error {
    return r.DB.WithContext(ctx).Create(o).Error
}

func (r *OTPRepository) FindLatest(ctx context.Context, target string) (*models.OTPRequest, error) {
    var o models.OTPRequest
    if err := r.DB.WithContext(ctx).Where("target = ?", target).Order("created_at desc").First(&o).Error; err != nil {
        return nil, err
    }
    return &o, nil
}

func (r *OTPRepository) Update(ctx context.Context, o *models.OTPRequest) error {
    return r.DB.WithContext(ctx).Save(o).Error
}

// Audit logs
type AuditRepo struct{
    DB *gorm.DB
}

func NewAuditRepo(db *gorm.DB) *AuditRepo { return &AuditRepo{DB: db} }

func (r *AuditRepo) Create(ctx context.Context, l *models.LoginAuditLog) error {
    return r.DB.WithContext(ctx).Create(l).Error
}
