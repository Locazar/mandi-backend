package usecase

import (
	"context"
	"testing"
	"time"

	"github.com/golang/mock/gomock"
	"github.com/rohit221990/mandi-backend/pkg/domain"
	"github.com/rohit221990/mandi-backend/pkg/mock/mockrepo"
	"github.com/rohit221990/mandi-backend/pkg/service/otp"
	"github.com/stretchr/testify/assert"
)

// VerifyShopDocument backs /api/admin/shops/business-document/verify-otp, which
// the seller onboarding wizard calls to mark a business document verified. It
// was a bare `return nil` — every code passed, so a logged-in seller could
// self-attest KYC without proving phone ownership. These cases pin the gate
// shut and confirm SKIP_OTP_VALIDATION still opens it deliberately.
func TestVerifyShopDocumentOtp(t *testing.T) {
	const (
		correctOTP = "735182"
		wrongOTP   = "111111"
		otpID      = "otp_doc_1"
		adminID    = "adm_test1"
	)

	validHash, err := otp.NewMobileOTPService(0).HashOTP(correctOTP)
	assert.NoError(t, err)

	activeSession := func() domain.OtpSession {
		return domain.OtpSession{
			OtpID:    otpID,
			OtpHash:  validHash,
			AdminID:  adminID,
			Phone:    "9876543210",
			UserType: domain.UserTypeAdmin,
			ExpireAt: time.Now().Add(5 * time.Minute),
		}
	}

	tests := []struct {
		testName       string
		skipValidation bool
		callerID       string
		otp            string
		buildStub      func(authRepo *mockrepo.MockAuthRepository)
		expectError    bool
	}{
		{
			testName: "WrongOtpIsRejectedAndSessionSurvives",
			callerID: adminID, otp: wrongOTP,
			buildStub: func(a *mockrepo.MockAuthRepository) {
				a.EXPECT().FindLatestOtpSessionByAdminID(gomock.Any(), adminID).Times(1).Return(activeSession(), nil)
				a.EXPECT().DeleteOtpSession(gomock.Any(), gomock.Any()).Times(0)
			},
			expectError: true,
		},
		{
			testName: "NoOutstandingOtpIsRejected",
			callerID: adminID, otp: correctOTP,
			buildStub: func(a *mockrepo.MockAuthRepository) {
				a.EXPECT().FindLatestOtpSessionByAdminID(gomock.Any(), adminID).Times(1).Return(domain.OtpSession{}, nil)
			},
			expectError: true,
		},
		{
			testName: "ExpiredOtpIsRejected",
			callerID: adminID, otp: correctOTP,
			buildStub: func(a *mockrepo.MockAuthRepository) {
				sess := activeSession()
				sess.ExpireAt = time.Now().Add(-time.Minute)
				a.EXPECT().FindLatestOtpSessionByAdminID(gomock.Any(), adminID).Times(1).Return(sess, nil)
			},
			expectError: true,
		},
		{
			testName: "SessionIssuedToAnotherAdminIsRejected",
			callerID: adminID, otp: correctOTP,
			buildStub: func(a *mockrepo.MockAuthRepository) {
				sess := activeSession()
				sess.AdminID = "adm_someone_else"
				a.EXPECT().FindLatestOtpSessionByAdminID(gomock.Any(), adminID).Times(1).Return(sess, nil)
			},
			expectError: true,
		},
		{
			testName:    "MissingCallerIdentityIsRejected",
			callerID:    "",
			otp:         correctOTP,
			buildStub:   func(a *mockrepo.MockAuthRepository) {},
			expectError: true,
		},
		{
			testName: "CorrectOtpPassesAndConsumesSession",
			callerID: adminID, otp: correctOTP,
			buildStub: func(a *mockrepo.MockAuthRepository) {
				a.EXPECT().FindLatestOtpSessionByAdminID(gomock.Any(), adminID).Times(1).Return(activeSession(), nil)
				a.EXPECT().DeleteOtpSession(gomock.Any(), otpID).Times(1).Return(nil)
			},
			expectError: false,
		},
		{
			testName:       "SkipValidationAcceptsAnyOtp",
			skipValidation: true,
			callerID:       adminID, otp: wrongOTP,
			buildStub: func(a *mockrepo.MockAuthRepository) {
				a.EXPECT().FindLatestOtpSessionByAdminID(gomock.Any(), adminID).Times(1).Return(activeSession(), nil)
				a.EXPECT().DeleteOtpSession(gomock.Any(), otpID).Times(1).Return(nil)
			},
			expectError: false,
		},
	}

	for _, test := range tests {
		test := test
		t.Run(test.testName, func(t *testing.T) {
			t.Parallel()
			ctl := gomock.NewController(t)
			defer ctl.Finish()

			authMockRepo := mockrepo.NewMockAuthRepository(ctl)
			test.buildStub(authMockRepo)

			useCase := &adminUseCase{
				authRepo:          authMockRepo,
				otpService:        otp.NewMobileOTPService(0),
				skipOTPValidation: test.skipValidation,
			}

			err := useCase.VerifyShopDocument(context.Background(), test.callerID, test.otp)
			if test.expectError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}
