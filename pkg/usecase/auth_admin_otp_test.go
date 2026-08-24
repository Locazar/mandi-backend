package usecase

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/golang/mock/gomock"
	"github.com/rohit221990/mandi-backend/pkg/api/handler/request"
	"github.com/rohit221990/mandi-backend/pkg/domain"
	"github.com/rohit221990/mandi-backend/pkg/mock/mockrepo"
	"github.com/rohit221990/mandi-backend/pkg/repository/interfaces"
	"github.com/rohit221990/mandi-backend/pkg/service/otp"
	"github.com/rohit221990/mandi-backend/pkg/utils"
	"github.com/stretchr/testify/assert"
	"gorm.io/gorm"
)

// fakeOtpAuth stands in for the external SMS/email OTP provider. It records the
// phone it was asked about so tests can assert the country code is applied, and
// returns whatever verdict the test sets.
type fakeOtpAuth struct {
	validOTP     string
	verifyErr    error
	gotPhone     string
	verifyCalled bool
}

func (f *fakeOtpAuth) SentOtp(phoneNumber string) (string, error) { return "sent", nil }

func (f *fakeOtpAuth) VerifyOtp(phoneNumber, code string) (bool, error) {
	f.verifyCalled = true
	f.gotPhone = phoneNumber
	if f.verifyErr != nil {
		return false, f.verifyErr
	}
	return code == f.validOTP, nil
}

func (f *fakeOtpAuth) SentOtpEmail(email string) (string, error) { return "sent", nil }

func (f *fakeOtpAuth) VerifyOtpEmail(email, code string) (bool, error) {
	f.verifyCalled = true
	return code == f.validOTP, nil
}

// fakeAdminRepo implements only the two AdminRepository methods this flow
// touches; the embedded interface (nil) satisfies the rest, so any unexpected
// call nil-panics — a loud signal rather than a silent pass.
type fakeAdminRepo struct {
	interfaces.AdminRepository
	existing      domain.Admin
	findErr       error
	saved         domain.Admin
	findCalls     int
	saveAdminCall int
}

func (f *fakeAdminRepo) FindAdminWithShopVerificationByPhone(ctx context.Context, phone string) (domain.Admin, domain.ShopVerification, error) {
	f.findCalls++
	return f.existing, domain.ShopVerification{}, f.findErr
}

func (f *fakeAdminRepo) SaveAdmin(ctx context.Context, admin domain.Admin) (domain.Admin, error) {
	f.saveAdminCall++
	return f.saved, nil
}

func newAdminOtpUseCase(t *testing.T, ctrl *gomock.Controller, provider *fakeOtpAuth, skipValidation bool) (*authUseCase, *mockrepo.MockAuthRepository, *fakeAdminRepo) {
	t.Helper()
	authRepo := mockrepo.NewMockAuthRepository(ctrl)
	adminRepo := &fakeAdminRepo{}
	return &authUseCase{
		authRepo:          authRepo,
		adminRepo:         adminRepo,
		optAuth:           provider,
		otpService:        otp.NewMobileOTPService(0),
		skipOTPValidation: skipValidation,
	}, authRepo, adminRepo
}

func adminOtpSession(expireAt time.Time) domain.OtpSession {
	return domain.OtpSession{
		OtpID:    "otp_admin_1",
		UserID:   "adm_1",
		AdminID:  "adm_1",
		Phone:    "9876543210",
		UserType: domain.UserType("admin"),
		ExpireAt: expireAt,
		// No OtpHash: AdminSignUpOtpSend delegates to the SMS provider and
		// stores no local hash. This is why verification must use optAuth.
	}
}

// The bug: seller signup/login accepted ANY OTP. Worse than the customer case,
// because an unknown phone number was silently REGISTERED as a seller account.
func TestAdminSignUpOtpVerifyRejectsWrongOtp(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	provider := &fakeOtpAuth{validOTP: "123456"}
	uc, authRepo, adminRepo := newAdminOtpUseCase(t, ctrl, provider, false)
	session := adminOtpSession(time.Now().Add(5 * time.Minute))

	authRepo.EXPECT().FindOtpSession(gomock.Any(), session.OtpID).Times(1).Return(session, nil)
	// No login, no registration, and the OTP stays unspent.
	authRepo.EXPECT().DeleteOtpSession(gomock.Any(), gomock.Any()).Times(0)
	defer func() {
		assert.Zero(t, adminRepo.findCalls, "must not attempt login on a bad OTP")
		assert.Zero(t, adminRepo.saveAdminCall, "must NEVER register an account on a bad OTP")
	}()

	adminID, err := uc.AdminSignUpOtpVerify(context.Background(), request.OTPVerify{
		OtpID: session.OtpID,
		Otp:   "999999", // wrong
	})

	assert.ErrorIs(t, err, ErrInvalidOtp)
	assert.Empty(t, adminID)
	assert.True(t, provider.verifyCalled, "the provider must actually be consulted")
	// The provider was seeded with the country code, matching AdminSignUpOtpSend.
	assert.Equal(t, countryCode+session.Phone, provider.gotPhone)
}

// The correct OTP still logs an existing seller in, and spends the OTP.
func TestAdminSignUpOtpVerifyAcceptsCorrectOtpForExistingAdmin(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	provider := &fakeOtpAuth{validOTP: "123456"}
	uc, authRepo, adminRepo := newAdminOtpUseCase(t, ctrl, provider, false)
	session := adminOtpSession(time.Now().Add(5 * time.Minute))

	authRepo.EXPECT().FindOtpSession(gomock.Any(), session.OtpID).Times(1).Return(session, nil)
	authRepo.EXPECT().DeleteOtpSession(gomock.Any(), session.OtpID).Times(1).Return(nil)
	// A password-bearing account: AdminLogin rejects password-less ones with
	// ErrWrongPassword (see the note in TestAdminSignUpOtpVerify_PasswordlessAdmin).
	hashed, hashErr := utils.GetHashedPassword("")
	assert.NoError(t, hashErr)
	adminRepo.existing = domain.Admin{ID: "adm_1", Mobile: session.Phone, Password: hashed}

	adminID, err := uc.AdminSignUpOtpVerify(context.Background(), request.OTPVerify{
		OtpID: session.OtpID,
		Otp:   "123456",
	})

	assert.NoError(t, err)
	assert.Equal(t, "adm_1", adminID)
}

// A brand-new phone still registers a seller — but only after the OTP checks out.
func TestAdminSignUpOtpVerifyRegistersNewAdminAfterValidOtp(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	provider := &fakeOtpAuth{validOTP: "123456"}
	uc, authRepo, adminRepo := newAdminOtpUseCase(t, ctrl, provider, false)
	session := adminOtpSession(time.Now().Add(5 * time.Minute))

	authRepo.EXPECT().FindOtpSession(gomock.Any(), session.OtpID).Times(1).Return(session, nil)
	authRepo.EXPECT().DeleteOtpSession(gomock.Any(), session.OtpID).Times(1).Return(nil)
	adminRepo.findErr = gorm.ErrRecordNotFound
	adminRepo.saved = domain.Admin{ID: "adm_new"}

	adminID, err := uc.AdminSignUpOtpVerify(context.Background(), request.OTPVerify{
		OtpID: session.OtpID,
		Otp:   "123456",
	})

	assert.NoError(t, err)
	assert.Equal(t, "adm_new", adminID)
}

// An expired session is refused even with the right code.
func TestAdminSignUpOtpVerifyRejectsExpiredOtp(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	provider := &fakeOtpAuth{validOTP: "123456"}
	uc, authRepo, adminRepo := newAdminOtpUseCase(t, ctrl, provider, false)
	session := adminOtpSession(time.Now().Add(-1 * time.Minute))

	authRepo.EXPECT().FindOtpSession(gomock.Any(), session.OtpID).Times(1).Return(session, nil)

	adminID, err := uc.AdminSignUpOtpVerify(context.Background(), request.OTPVerify{
		OtpID: session.OtpID,
		Otp:   "123456",
	})

	assert.Zero(t, adminRepo.saveAdminCall)
	assert.ErrorIs(t, err, ErrOtpExpired)
	assert.Empty(t, adminID)
	assert.False(t, provider.verifyCalled, "expiry should short-circuit before the provider call")
}

// A provider outage must fail closed, not wave the login through.
func TestAdminSignUpOtpVerifyFailsClosedOnProviderError(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	provider := &fakeOtpAuth{validOTP: "123456", verifyErr: errors.New("provider unreachable")}
	uc, authRepo, adminRepo := newAdminOtpUseCase(t, ctrl, provider, false)
	session := adminOtpSession(time.Now().Add(5 * time.Minute))

	authRepo.EXPECT().FindOtpSession(gomock.Any(), session.OtpID).Times(1).Return(session, nil)

	adminID, err := uc.AdminSignUpOtpVerify(context.Background(), request.OTPVerify{
		OtpID: session.OtpID,
		Otp:   "123456",
	})

	assert.Zero(t, adminRepo.saveAdminCall, "a provider outage must not create an account")
	assert.Error(t, err)
	assert.Empty(t, adminID)
}

// The documented dev bypass still works for the seller path too.
func TestAdminSignUpOtpVerifySkipValidationBypassesWrongOtp(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	provider := &fakeOtpAuth{validOTP: "123456"}
	uc, authRepo, adminRepo := newAdminOtpUseCase(t, ctrl, provider, true)
	session := adminOtpSession(time.Now().Add(5 * time.Minute))

	authRepo.EXPECT().FindOtpSession(gomock.Any(), session.OtpID).Times(1).Return(session, nil)
	authRepo.EXPECT().DeleteOtpSession(gomock.Any(), session.OtpID).Times(1).Return(nil)
	hashed, hashErr := utils.GetHashedPassword("")
	assert.NoError(t, hashErr)
	adminRepo.existing = domain.Admin{ID: "adm_1", Password: hashed}

	adminID, err := uc.AdminSignUpOtpVerify(context.Background(), request.OTPVerify{
		OtpID: session.OtpID,
		Otp:   "000000",
	})

	assert.NoError(t, err)
	assert.Equal(t, "adm_1", adminID)
	assert.False(t, provider.verifyCalled, "bypass should not consult the provider")
}

// Documents a PRE-EXISTING quirk this fix does not change: AdminLogin rejects
// an account with no password set (ErrWrongPassword), and AdminSignUpOtpVerify
// treats anything other than ErrUserNotExist as fatal. So an OTP-only seller —
// exactly what this function creates via SaveAdmin — cannot log back in through
// it even with a correct OTP. Unreachable in production today (the route is
// dead), but pinned here so the behaviour is visible if it is ever revived.
func TestAdminSignUpOtpVerify_PasswordlessAdminCannotLogIn(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	provider := &fakeOtpAuth{validOTP: "123456"}
	uc, authRepo, adminRepo := newAdminOtpUseCase(t, ctrl, provider, false)
	session := adminOtpSession(time.Now().Add(5 * time.Minute))

	authRepo.EXPECT().FindOtpSession(gomock.Any(), session.OtpID).Times(1).Return(session, nil)
	authRepo.EXPECT().DeleteOtpSession(gomock.Any(), session.OtpID).Times(1).Return(nil)
	adminRepo.existing = domain.Admin{ID: "adm_1", Mobile: session.Phone} // no password

	_, err := uc.AdminSignUpOtpVerify(context.Background(), request.OTPVerify{
		OtpID: session.OtpID,
		Otp:   "123456", // correct OTP
	})

	// The OTP itself passed — the failure is the password check downstream.
	assert.Error(t, err)
	assert.True(t, provider.verifyCalled)
}
