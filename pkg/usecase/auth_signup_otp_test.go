package usecase

import (
	"context"
	"testing"
	"time"

	"github.com/golang/mock/gomock"
	"github.com/rohit221990/mandi-backend/pkg/api/handler/request"
	"github.com/rohit221990/mandi-backend/pkg/domain"
	"github.com/rohit221990/mandi-backend/pkg/mock/mockrepo"
	"github.com/rohit221990/mandi-backend/pkg/service/otp"
	"github.com/stretchr/testify/assert"
)

// newSignUpOtpUseCase wires an authUseCase with a real MobileOTPService (it is
// pure crypto, no I/O) and mocked repositories, so the tests exercise the
// actual hash comparison rather than a stub of it.
func newSignUpOtpUseCase(t *testing.T, ctrl *gomock.Controller, skipValidation bool) (*authUseCase, *mockrepo.MockAuthRepository, *mockrepo.MockUserRepository) {
	t.Helper()
	authRepo := mockrepo.NewMockAuthRepository(ctrl)
	userRepo := mockrepo.NewMockUserRepository(ctrl)
	return &authUseCase{
		authRepo:          authRepo,
		userRepo:          &userRepoAdapter{MockUserRepository: userRepo},
		otpService:        otp.NewMobileOTPService(0),
		skipOTPValidation: skipValidation,
	}, authRepo, userRepo
}

// otpSessionFor builds a stored session holding the bcrypt hash of realOTP.
func otpSessionFor(t *testing.T, realOTP string, expireAt time.Time) domain.OtpSession {
	t.Helper()
	hash, err := otp.NewMobileOTPService(0).HashOTP(realOTP)
	assert.NoError(t, err)
	return domain.OtpSession{
		OtpID:    "otp_session_1",
		OtpHash:  hash,
		UserID:   "usr_1",
		Phone:    "9876543210",
		UserType: domain.UserTypeUser,
		ExpireAt: expireAt,
	}
}

// The bug: signup verification accepted ANY OTP value, so anyone holding an
// otp_id could obtain a token without knowing the code that was texted.
func TestSingUpOtpVerifyRejectsWrongOtp(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	uc, authRepo, userRepo := newSignUpOtpUseCase(t, ctrl, false)
	session := otpSessionFor(t, "123456", time.Now().Add(5*time.Minute))

	authRepo.EXPECT().FindOtpSession(gomock.Any(), session.OtpID).Times(1).Return(session, nil)
	// The account must NOT be marked verified and the session must NOT be consumed.
	userRepo.EXPECT().UpdateVerified(gomock.Any(), gomock.Any()).Times(0)
	authRepo.EXPECT().DeleteOtpSession(gomock.Any(), gomock.Any()).Times(0)

	userID, err := uc.SingUpOtpVerify(context.Background(), request.OTPVerify{
		OtpID: session.OtpID,
		Otp:   "999999", // wrong
	})

	assert.ErrorIs(t, err, ErrInvalidOtp)
	assert.Empty(t, userID)
}

// The correct OTP still works: user marked verified, session consumed.
func TestSingUpOtpVerifyAcceptsCorrectOtp(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	uc, authRepo, userRepo := newSignUpOtpUseCase(t, ctrl, false)
	session := otpSessionFor(t, "123456", time.Now().Add(5*time.Minute))

	authRepo.EXPECT().FindOtpSession(gomock.Any(), session.OtpID).Times(1).Return(session, nil)
	userRepo.EXPECT().UpdateVerified(gomock.Any(), session.UserID).Times(1).Return(nil)
	authRepo.EXPECT().DeleteOtpSession(gomock.Any(), session.OtpID).Times(1).Return(nil)

	userID, err := uc.SingUpOtpVerify(context.Background(), request.OTPVerify{
		OtpID: session.OtpID,
		Otp:   "123456",
	})

	assert.NoError(t, err)
	assert.Equal(t, session.UserID, userID)
}

// An expired session must be refused even when the code itself is correct.
func TestSingUpOtpVerifyRejectsExpiredOtp(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	uc, authRepo, userRepo := newSignUpOtpUseCase(t, ctrl, false)
	session := otpSessionFor(t, "123456", time.Now().Add(-1*time.Minute)) // already expired

	authRepo.EXPECT().FindOtpSession(gomock.Any(), session.OtpID).Times(1).Return(session, nil)
	userRepo.EXPECT().UpdateVerified(gomock.Any(), gomock.Any()).Times(0)

	userID, err := uc.SingUpOtpVerify(context.Background(), request.OTPVerify{
		OtpID: session.OtpID,
		Otp:   "123456",
	})

	assert.ErrorIs(t, err, ErrOtpExpired)
	assert.Empty(t, userID)
}

// The session is deleted on success so the same OTP cannot be replayed for a
// second token.
func TestSingUpOtpVerifyConsumesSession(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	uc, authRepo, userRepo := newSignUpOtpUseCase(t, ctrl, false)
	session := otpSessionFor(t, "654321", time.Now().Add(5*time.Minute))

	authRepo.EXPECT().FindOtpSession(gomock.Any(), session.OtpID).Times(1).Return(session, nil)
	userRepo.EXPECT().UpdateVerified(gomock.Any(), session.UserID).Times(1).Return(nil)
	// The assertion that matters: exactly one delete of this session.
	authRepo.EXPECT().DeleteOtpSession(gomock.Any(), session.OtpID).Times(1).Return(nil)

	_, err := uc.SingUpOtpVerify(context.Background(), request.OTPVerify{
		OtpID: session.OtpID,
		Otp:   "654321",
	})
	assert.NoError(t, err)
}

// The documented dev bypass still works, so local/test environments that set
// SKIP_OTP_VALIDATION=true are not broken by this fix.
func TestSingUpOtpVerifySkipValidationBypassesWrongOtp(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	uc, authRepo, userRepo := newSignUpOtpUseCase(t, ctrl, true)
	session := otpSessionFor(t, "123456", time.Now().Add(5*time.Minute))

	authRepo.EXPECT().FindOtpSession(gomock.Any(), session.OtpID).Times(1).Return(session, nil)
	userRepo.EXPECT().UpdateVerified(gomock.Any(), session.UserID).Times(1).Return(nil)
	authRepo.EXPECT().DeleteOtpSession(gomock.Any(), session.OtpID).Times(1).Return(nil)

	userID, err := uc.SingUpOtpVerify(context.Background(), request.OTPVerify{
		OtpID: session.OtpID,
		Otp:   "000000",
	})

	assert.NoError(t, err)
	assert.Equal(t, session.UserID, userID)
}
