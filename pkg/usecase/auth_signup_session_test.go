package usecase

import (
	"context"
	"strings"
	"testing"

	"github.com/golang/mock/gomock"
	"github.com/rohit221990/mandi-backend/pkg/api/handler/request"
	"github.com/rohit221990/mandi-backend/pkg/domain"
	"github.com/rohit221990/mandi-backend/pkg/service/otp"
	"github.com/stretchr/testify/assert"
)

// UserSignUp and SingUpOtpVerify are two halves of one contract: the session
// signup saves must carry the bcrypt hash that verification compares against.
// They were allowed to drift apart once — verification started checking
// OtpHash while signup still saved a session without one, and because bcrypt
// rejects every candidate against an empty hash, no new user could ever
// complete signup. These tests pin the two halves together.
func TestUserSignUpStoresVerifiableOtpSession(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	useCase, authRepo, userRepo := newSignUpOtpUseCase(t, ctrl, true /* skip SMS, not verification */)

	signUp := domain.User{Phone: "9876543210"}

	userRepo.EXPECT().FindUserByUserNameEmailOrPhoneNotID(gomock.Any(), gomock.Any()).
		Times(1).Return(domain.User{}, nil)
	userRepo.EXPECT().SaveUser(gomock.Any(), gomock.Any()).
		Times(1).Return("usr_new", nil)

	var saved domain.OtpSession
	authRepo.EXPECT().SaveOtpSession(gomock.Any(), gomock.Any()).
		Times(1).
		DoAndReturn(func(_ context.Context, s domain.OtpSession) error {
			saved = s
			return nil
		})

	otpID, err := useCase.UserSignUp(context.Background(), signUp)
	assert.NoError(t, err)
	assert.NotEmpty(t, otpID, "signup must return an otp_id for the client to verify with")

	// The invariant that broke: a session with no hash is unverifiable.
	assert.NotEmpty(t, saved.OtpHash,
		"signup must store an OTP hash — SingUpOtpVerify compares against it")
	assert.True(t, strings.HasPrefix(saved.OtpHash, "$2"),
		"stored value must be a bcrypt hash, got %q", saved.OtpHash)
	assert.Equal(t, otpID, saved.OtpID, "returned otp_id must identify the saved session")
	assert.Equal(t, "usr_new", saved.UserID)
	assert.Equal(t, signUp.Phone, saved.Phone)
}

// The session signup produces must be usable by the real verification path:
// a wrong code is rejected as an invalid OTP, not as a malformed hash.
func TestSignUpSessionIsVerifiableByRealVerifier(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	signUpCase, authRepo, userRepo := newSignUpOtpUseCase(t, ctrl, true)

	userRepo.EXPECT().FindUserByUserNameEmailOrPhoneNotID(gomock.Any(), gomock.Any()).
		Times(1).Return(domain.User{}, nil)
	userRepo.EXPECT().SaveUser(gomock.Any(), gomock.Any()).Times(1).Return("usr_new", nil)

	var saved domain.OtpSession
	authRepo.EXPECT().SaveOtpSession(gomock.Any(), gomock.Any()).Times(1).
		DoAndReturn(func(_ context.Context, s domain.OtpSession) error {
			saved = s
			return nil
		})

	otpID, err := signUpCase.UserSignUp(context.Background(), domain.User{Phone: "9876543210"})
	assert.NoError(t, err)

	// Now verify with validation ON, against the session signup just produced.
	verifyCase, verifyAuthRepo, _ := newSignUpOtpUseCase(t, ctrl, false)
	verifyCase.otpService = otp.NewMobileOTPService(0)
	verifyAuthRepo.EXPECT().FindOtpSession(gomock.Any(), otpID).Times(1).Return(saved, nil)

	_, err = verifyCase.SingUpOtpVerify(context.Background(), request.OTPVerify{
		OtpID: otpID,
		Otp:   "000000", // deliberately wrong
	})
	assert.ErrorIs(t, err, ErrInvalidOtp,
		"a wrong code must fail as an invalid OTP, not as a malformed/empty hash")
}
