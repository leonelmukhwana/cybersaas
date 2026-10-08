package mpesa

import (
	"context"
	"errors"
	"fmt"
	"strconv"
	"strings"
	"time"

	"cybersaas/backend/payment"
	"cybersaas/backend/receipt"
	"cybersaas/backend/subscription"

	"github.com/google/uuid"
)

var (
	ErrInvalidProvider    = errors.New("invalid mpesa provider")
	ErrInvalidEnvironment = errors.New("invalid mpesa environment")
	ErrInvalidBranchID    = errors.New("invalid branch id")
)

type Service struct {
	repository       *Repository
	encryptor        *Encryptor
	subscriptionRepo *subscription.Repository
	paymentRepo      *payment.Repository
	receiptRepo      *receipt.Repository
}

func NewService(
	repository *Repository,
	encryptor *Encryptor,
	subscriptionRepo *subscription.Repository,
	paymentRepo *payment.Repository,
	receiptRepo *receipt.Repository,
) *Service {
	return &Service{
		repository:       repository,
		encryptor:        encryptor,
		subscriptionRepo: subscriptionRepo,
		paymentRepo:      paymentRepo,
		receiptRepo:      receiptRepo,
	}
}

func (s *Service) SaveBranchConfiguration(
	ctx context.Context,
	tenantID string,
	request SaveConfigurationRequest,
) (*Configuration, error) {
	if _, err := uuid.Parse(tenantID); err != nil {
		return nil, errors.New("invalid tenant id")
	}

	if _, err := uuid.Parse(request.BranchID); err != nil {
		return nil, ErrInvalidBranchID
	}

	request.Provider = strings.ToLower(strings.TrimSpace(request.Provider))

	if request.Provider == "" {
		request.Provider = "daraja"
	}

	if request.Provider != "daraja" {
		return nil, ErrInvalidProvider
	}

	request.Environment = strings.ToLower(strings.TrimSpace(request.Environment))

	if request.Environment != "sandbox" &&
		request.Environment != "production" {
		return nil, ErrInvalidEnvironment
	}

	if err := s.repository.VerifyBranchBelongsToTenant(
		ctx,
		tenantID,
		request.BranchID,
	); err != nil {
		return nil, err
	}

	consumerKey, err := s.encryptOptional(request.ConsumerKey)
	if err != nil {
		return nil, err
	}

	consumerSecret, err := s.encryptOptional(request.ConsumerSecret)
	if err != nil {
		return nil, err
	}

	passkey, err := s.encryptOptional(request.Passkey)
	if err != nil {
		return nil, err
	}

	active := true

	if request.Active != nil {
		active = *request.Active
	}

	err = s.repository.SaveBranchConfiguration(
		ctx,
		SaveBranchConfigurationParams{
			ID:                      uuid.New().String(),
			TenantID:                tenantID,
			BranchID:                request.BranchID,
			Provider:                request.Provider,
			Environment:             request.Environment,
			BusinessShortCode:       normalizeOptional(request.BusinessShortCode),
			TillNumber:              normalizeOptional(request.TillNumber),
			PaybillNumber:           normalizeOptional(request.PaybillNumber),
			ConsumerKeyEncrypted:    consumerKey,
			ConsumerSecretEncrypted: consumerSecret,
			PasskeyEncrypted:        passkey,
			AccountReference:        normalizeOptional(request.AccountReference),
			CallbackURL:             normalizeOptional(request.CallbackURL),
			Active:                  active,
		},
	)
	if err != nil {
		return nil, err
	}

	return s.repository.GetBranchConfiguration(
		ctx,
		tenantID,
		request.BranchID,
	)
}

func (s *Service) GetBranchConfiguration(
	ctx context.Context,
	tenantID string,
	branchID string,
) (*Configuration, error) {
	if _, err := uuid.Parse(tenantID); err != nil {
		return nil, errors.New("invalid tenant id")
	}

	if _, err := uuid.Parse(branchID); err != nil {
		return nil, ErrInvalidBranchID
	}

	return s.repository.GetBranchConfiguration(
		ctx,
		tenantID,
		branchID,
	)
}

func (s *Service) DeleteBranchConfiguration(
	ctx context.Context,
	tenantID string,
	branchID string,
) error {
	if _, err := uuid.Parse(tenantID); err != nil {
		return errors.New("invalid tenant id")
	}

	if _, err := uuid.Parse(branchID); err != nil {
		return ErrInvalidBranchID
	}

	return s.repository.DeleteBranchConfiguration(
		ctx,
		tenantID,
		branchID,
	)
}

func (s *Service) SavePlatformConfiguration(
	ctx context.Context,
	request PlatformSaveConfigurationRequest,
) (*Configuration, error) {
	request.Provider = strings.ToLower(strings.TrimSpace(request.Provider))

	if request.Provider == "" {
		request.Provider = "daraja"
	}

	if request.Provider != "daraja" {
		return nil, ErrInvalidProvider
	}

	request.Environment = strings.ToLower(strings.TrimSpace(request.Environment))

	if request.Environment != "sandbox" &&
		request.Environment != "production" {
		return nil, ErrInvalidEnvironment
	}

	consumerKey, err := s.encryptOptional(request.ConsumerKey)
	if err != nil {
		return nil, err
	}

	consumerSecret, err := s.encryptOptional(request.ConsumerSecret)
	if err != nil {
		return nil, err
	}

	passkey, err := s.encryptOptional(request.Passkey)
	if err != nil {
		return nil, err
	}

	active := true

	if request.Active != nil {
		active = *request.Active
	}

	err = s.repository.SavePlatformConfiguration(
		ctx,
		SavePlatformConfigurationParams{
			ID:                      uuid.New().String(),
			Provider:                request.Provider,
			Environment:             request.Environment,
			BusinessShortCode:       normalizeOptional(request.BusinessShortCode),
			TillNumber:              normalizeOptional(request.TillNumber),
			PaybillNumber:           normalizeOptional(request.PaybillNumber),
			ConsumerKeyEncrypted:    consumerKey,
			ConsumerSecretEncrypted: consumerSecret,
			PasskeyEncrypted:        passkey,
			AccountReference:        normalizeOptional(request.AccountReference),
			CallbackURL:             normalizeOptional(request.CallbackURL),
			Active:                  active,
		},
	)
	if err != nil {
		return nil, err
	}

	return s.repository.GetPlatformConfiguration(ctx)
}

func (s *Service) GetPlatformConfiguration(
	ctx context.Context,
) (*Configuration, error) {
	return s.repository.GetPlatformConfiguration(ctx)
}

// InitiateSTKPush starts an M-Pesa STK Push.
func (s *Service) InitiateSTKPush(
	ctx context.Context,
	tenantID string,
	userID string,
	request InitiateSTKRequest,
) (*STKRequest, error) {
	if _, err := uuid.Parse(tenantID); err != nil {
		return nil, errors.New("invalid tenant id")
	}

	if _, err := uuid.Parse(userID); err != nil {
		return nil, errors.New("invalid user id")
	}

	if _, err := uuid.Parse(request.BranchID); err != nil {
		return nil, ErrInvalidBranchID
	}

	if _, err := uuid.Parse(request.SaleID); err != nil {
		return nil, errors.New("invalid sale id")
	}

	if request.PaymentID != nil {
		if _, err := uuid.Parse(*request.PaymentID); err != nil {
			return nil, errors.New("invalid payment id")
		}
	}

	request.PhoneNumber = strings.TrimSpace(request.PhoneNumber)
	request.Amount = strings.TrimSpace(request.Amount)
	request.AccountReference = strings.TrimSpace(request.AccountReference)
	request.TransactionDescription = strings.TrimSpace(
		request.TransactionDescription,
	)

	if request.PhoneNumber == "" {
		return nil, errors.New("phone number is required")
	}

	if request.Amount == "" {
		return nil, errors.New("amount is required")
	}

	if request.AccountReference == "" {
		return nil, errors.New("account reference is required")
	}

	if request.TransactionDescription == "" {
		return nil, errors.New("transaction description is required")
	}

	formattedPhone, err := normalizeMpesaPhone(
		request.PhoneNumber,
	)
	if err != nil {
		return nil, err
	}

	numericAmount, err := parseSTKAmount(request.Amount)
	if err != nil {
		return nil, err
	}

	if err := s.repository.VerifyBranchBelongsToTenant(
		ctx,
		tenantID,
		request.BranchID,
	); err != nil {
		return nil, err
	}

	config, err := s.repository.GetBranchConfiguration(
		ctx,
		tenantID,
		request.BranchID,
	)
	if err != nil {
		return nil, err
	}

	if !config.Active {
		return nil, errors.New("mpesa configuration is inactive")
	}

	if config.Provider != "daraja" {
		return nil, ErrInvalidProvider
	}

	if !config.ConsumerKeyConfigured ||
		!config.ConsumerSecretConfigured ||
		!config.PasskeyConfigured {
		return nil, errors.New("mpesa credentials are incomplete")
	}

	credentials, err := s.repository.GetBranchCredentials(
		ctx,
		tenantID,
		request.BranchID,
	)
	if err != nil {
		return nil, err
	}

	consumerKey, err := s.encryptor.Decrypt(
		credentials.ConsumerKeyEncrypted,
	)
	if err != nil {
		return nil, fmt.Errorf(
			"failed to decrypt mpesa consumer key: %w",
			err,
		)
	}

	consumerSecret, err := s.encryptor.Decrypt(
		credentials.ConsumerSecretEncrypted,
	)
	if err != nil {
		return nil, fmt.Errorf(
			"failed to decrypt mpesa consumer secret: %w",
			err,
		)
	}

	passkey, err := s.encryptor.Decrypt(
		credentials.PasskeyEncrypted,
	)
	if err != nil {
		return nil, fmt.Errorf(
			"failed to decrypt mpesa passkey: %w",
			err,
		)
	}

	businessShortCode := getShortCode(config)

	if businessShortCode == "" {
		return nil, errors.New(
			"mpesa business short code is required",
		)
	}

	tillNumber := ""
	if config.TillNumber != nil {
		tillNumber = strings.TrimSpace(*config.TillNumber)
	}

	paybillNumber := ""
	if config.PaybillNumber != nil {
		paybillNumber = strings.TrimSpace(*config.PaybillNumber)
	}

	callbackURL := ""
	if config.CallbackURL != nil {
		callbackURL = strings.TrimSpace(*config.CallbackURL)
	}

	if callbackURL == "" {
		return nil, errors.New(
			"mpesa callback url is required",
		)
	}

	if !IsValidCallbackURL(callbackURL) {
		return nil, errors.New(
			"mpesa callback url must be a valid HTTPS URL",
		)
	}

	accountReference := request.AccountReference

	darajaClient := NewDarajaClient(
		DarajaConfig{
			Environment:       DarajaEnvironment(config.Environment),
			ConsumerKey:       consumerKey,
			ConsumerSecret:    consumerSecret,
			Passkey:           passkey,
			BusinessShortCode: businessShortCode,
			TillNumber:        tillNumber,
			PaybillNumber:     paybillNumber,
			CallbackURL:       callbackURL,
			AccountReference:  accountReference,
		},
	)

	stkID := uuid.New().String()

	err = s.repository.CreateSTKRequest(
		ctx,
		CreateSTKRequestParams{
			ID:                     stkID,
			TenantID:               tenantID,
			BranchID:               request.BranchID,
			SaleID:                 request.SaleID,
			PaymentID:              request.PaymentID,
			PhoneNumber:            formattedPhone,
			Amount:                 request.Amount,
			AccountReference:       accountReference,
			TransactionDescription: request.TransactionDescription,
		},
	)
	if err != nil {
		return nil, err
	}

	response, err := darajaClient.InitiateSTKPush(
		ctx,
		STKPushRequest{
			Amount:           numericAmount,
			PhoneNumber:      formattedPhone,
			AccountReference: accountReference,
			TransactionDesc:  request.TransactionDescription,
		},
	)
	if err != nil {
		_ = s.repository.UpdateSTKFailed(
			ctx,
			stkID,
			"",
			err.Error(),
		)

		return nil, err
	}

	err = s.repository.UpdateSTKAccepted(
		ctx,
		stkID,
		response.MerchantRequestID,
		response.CheckoutRequestID,
		response.ResponseCode,
		response.ResponseDescription,
		response.CustomerMessage,
	)
	if err != nil {
		return nil, err
	}

	return s.repository.GetSTKRequest(
		ctx,
		stkID,
	)
}

// InitiateTerminalSTKPush starts an M-Pesa STK Push for a terminal payment.
//
// The terminal has already created the pending payment and sale.
// This method only starts the Daraja STK request and links it to
// the existing payment through PaymentID.
func (s *Service) InitiateTerminalSTKPush(
	ctx context.Context,
	tenantID string,
	terminalID string,
	branchID string,
	request InitiateSTKRequest,
) (*STKRequest, error) {
	if _, err := uuid.Parse(tenantID); err != nil {
		return nil, errors.New("invalid tenant id")
	}

	if _, err := uuid.Parse(terminalID); err != nil {
		return nil, errors.New("invalid terminal id")
	}

	if _, err := uuid.Parse(branchID); err != nil {
		return nil, errors.New("invalid branch id")
	}

	request.BranchID = strings.TrimSpace(branchID)

	if _, err := uuid.Parse(request.SaleID); err != nil {
		return nil, errors.New("invalid sale id")
	}

	if request.PaymentID == nil ||
		strings.TrimSpace(*request.PaymentID) == "" {
		return nil, errors.New("payment id is required")
	}

	if _, err := uuid.Parse(*request.PaymentID); err != nil {
		return nil, errors.New("invalid payment id")
	}

	request.PhoneNumber = strings.TrimSpace(request.PhoneNumber)
	request.Amount = strings.TrimSpace(request.Amount)
	request.AccountReference = strings.TrimSpace(request.AccountReference)
	request.TransactionDescription = strings.TrimSpace(
		request.TransactionDescription,
	)

	if request.PhoneNumber == "" {
		return nil, errors.New("phone number is required")
	}

	if request.Amount == "" {
		return nil, errors.New("amount is required")
	}

	if request.AccountReference == "" {
		return nil, errors.New("account reference is required")
	}

	if request.TransactionDescription == "" {
		return nil, errors.New("transaction description is required")
	}

	formattedPhone, err := normalizeMpesaPhone(request.PhoneNumber)
	if err != nil {
		return nil, err
	}

	numericAmount, err := parseSTKAmount(request.Amount)
	if err != nil {
		return nil, err
	}

	if err := s.repository.VerifyBranchBelongsToTenant(
		ctx,
		tenantID,
		branchID,
	); err != nil {
		return nil, err
	}

	config, err := s.repository.GetBranchConfiguration(
		ctx,
		tenantID,
		branchID,
	)
	if err != nil {
		return nil, err
	}

	if !config.Active {
		return nil, errors.New("mpesa configuration is inactive")
	}

	if config.Provider != "daraja" {
		return nil, ErrInvalidProvider
	}

	if !config.ConsumerKeyConfigured ||
		!config.ConsumerSecretConfigured ||
		!config.PasskeyConfigured {
		return nil, errors.New("mpesa credentials are incomplete")
	}

	credentials, err := s.repository.GetBranchCredentials(
		ctx,
		tenantID,
		branchID,
	)
	if err != nil {
		return nil, err
	}

	consumerKey, err := s.encryptor.Decrypt(
		credentials.ConsumerKeyEncrypted,
	)
	if err != nil {
		return nil, fmt.Errorf(
			"failed to decrypt mpesa consumer key: %w",
			err,
		)
	}

	consumerSecret, err := s.encryptor.Decrypt(
		credentials.ConsumerSecretEncrypted,
	)
	if err != nil {
		return nil, fmt.Errorf(
			"failed to decrypt mpesa consumer secret: %w",
			err,
		)
	}

	passkey, err := s.encryptor.Decrypt(
		credentials.PasskeyEncrypted,
	)
	if err != nil {
		return nil, fmt.Errorf(
			"failed to decrypt mpesa passkey: %w",
			err,
		)
	}

	businessShortCode := getShortCode(config)

	if businessShortCode == "" {
		return nil, errors.New(
			"mpesa business short code is required",
		)
	}

	tillNumber := ""
	if config.TillNumber != nil {
		tillNumber = strings.TrimSpace(*config.TillNumber)
	}

	paybillNumber := ""
	if config.PaybillNumber != nil {
		paybillNumber = strings.TrimSpace(*config.PaybillNumber)
	}

	callbackURL := ""
	if config.CallbackURL != nil {
		callbackURL = strings.TrimSpace(*config.CallbackURL)
	}

	if callbackURL == "" {
		return nil, errors.New(
			"mpesa callback url is required",
		)
	}

	if !IsValidCallbackURL(callbackURL) {
		return nil, errors.New(
			"mpesa callback url must be a valid HTTPS URL",
		)
	}

	darajaClient := NewDarajaClient(
		DarajaConfig{
			Environment:       DarajaEnvironment(config.Environment),
			ConsumerKey:       consumerKey,
			ConsumerSecret:    consumerSecret,
			Passkey:           passkey,
			BusinessShortCode: businessShortCode,
			TillNumber:        tillNumber,
			PaybillNumber:     paybillNumber,
			CallbackURL:       callbackURL,
			AccountReference:  request.AccountReference,
		},
	)

	stkID := uuid.New().String()

	err = s.repository.CreateSTKRequest(
		ctx,
		CreateSTKRequestParams{
			ID:                     stkID,
			TenantID:               tenantID,
			BranchID:               branchID,
			SaleID:                 request.SaleID,
			PaymentID:              request.PaymentID,
			PhoneNumber:            formattedPhone,
			Amount:                 request.Amount,
			AccountReference:       request.AccountReference,
			TransactionDescription: request.TransactionDescription,
		},
	)
	if err != nil {
		return nil, err
	}

	response, err := darajaClient.InitiateSTKPush(
		ctx,
		STKPushRequest{
			Amount:           numericAmount,
			PhoneNumber:      formattedPhone,
			AccountReference: request.AccountReference,
			TransactionDesc:  request.TransactionDescription,
		},
	)
	if err != nil {
		_ = s.repository.UpdateSTKFailed(
			ctx,
			stkID,
			"",
			err.Error(),
		)

		return nil, err
	}

	err = s.repository.UpdateSTKAccepted(
		ctx,
		stkID,
		response.MerchantRequestID,
		response.CheckoutRequestID,
		response.ResponseCode,
		response.ResponseDescription,
		response.CustomerMessage,
	)
	if err != nil {
		return nil, err
	}

	return s.repository.GetSTKRequest(
		ctx,
		stkID,
	)
}

func (s *Service) encryptOptional(
	value *string,
) (*string, error) {
	if value == nil {
		return nil, nil
	}

	valueCopy := strings.TrimSpace(*value)

	if valueCopy == "" {
		return nil, nil
	}

	encrypted, err := s.encryptor.Encrypt(valueCopy)
	if err != nil {
		return nil, err
	}

	return &encrypted, nil
}

func normalizeOptional(value *string) *string {
	if value == nil {
		return nil
	}

	valueCopy := strings.TrimSpace(*value)

	if valueCopy == "" {
		return nil
	}

	return &valueCopy
}

func normalizeMpesaPhone(phone string) (string, error) {
	p := strings.TrimSpace(phone)

	p = strings.ReplaceAll(p, " ", "")
	p = strings.ReplaceAll(p, "-", "")

	if strings.HasPrefix(p, "+") {
		p = strings.TrimPrefix(p, "+")
	}

	if strings.HasPrefix(p, "07") ||
		strings.HasPrefix(p, "01") {
		if len(p) != 10 {
			return "", errors.New(
				"invalid Kenyan phone number",
			)
		}

		p = "254" + p[1:]
	}

	if strings.HasPrefix(p, "254") {
		if len(p) != 12 {
			return "", errors.New(
				"invalid Kenyan phone number",
			)
		}
	} else {
		return "", errors.New(
			"invalid Kenyan phone number",
		)
	}

	for _, character := range p {
		if character < '0' || character > '9' {
			return "", errors.New(
				"invalid Kenyan phone number",
			)
		}
	}

	return p, nil
}

func getShortCode(config *Configuration) string {
	if config.BusinessShortCode != nil {
		if value := strings.TrimSpace(
			*config.BusinessShortCode,
		); value != "" {
			return value
		}
	}

	if config.PaybillNumber != nil {
		if value := strings.TrimSpace(
			*config.PaybillNumber,
		); value != "" {
			return value
		}
	}

	if config.TillNumber != nil {
		if value := strings.TrimSpace(
			*config.TillNumber,
		); value != "" {
			return value
		}
	}

	return ""
}

func parseSTKAmount(amount string) (int64, error) {
	amount = strings.TrimSpace(amount)

	if amount == "" {
		return 0, errors.New("amount is required")
	}

	value, err := strconv.ParseFloat(amount, 64)
	if err != nil {
		return 0, errors.New("amount must be a valid number")
	}

	if value <= 0 {
		return 0, errors.New("amount must be greater than zero")
	}

	whole := int64(value)

	if float64(whole) != value {
		return 0, errors.New("amount must be a whole number")
	}

	return whole, nil
}

func (s *Service) ProcessSTKCallback(
	ctx context.Context,
	callback STKCallback,
) (*STKRequest, error) {
	checkoutRequestID := strings.TrimSpace(callback.CheckoutRequestID)

	if checkoutRequestID == "" {
		return nil, errors.New("checkout request id is required")
	}

	stkRequest, normalErr :=
		s.repository.GetSTKRequestByCheckoutRequestID(
			ctx,
			checkoutRequestID,
		)

	if normalErr != nil {
		subscriptionSTKRequest, subscriptionErr :=
			s.repository.GetSubscriptionSTKRequestByCheckoutRequestID(
				ctx,
				checkoutRequestID,
			)

		if subscriptionErr != nil {
			return nil, ErrSTKRequestNotFound
		}

		return s.processSubscriptionSTKCallback(
			ctx,
			subscriptionSTKRequest,
			callback,
		)
	}

	return s.processNormalSTKCallback(
		ctx,
		stkRequest,
		callback,
	)
}

func (s *Service) processNormalSTKCallback(
	ctx context.Context,
	stkRequest *STKRequest,
	callback STKCallback,
) (*STKRequest, error) {
	if stkRequest == nil {
		return nil, ErrSTKRequestNotFound
	}

	if stkRequest.Status == "completed" ||
		stkRequest.Status == "failed" ||
		stkRequest.Status == "cancelled" ||
		stkRequest.Status == "timeout" {
		return stkRequest, nil
	}

	resultCode := strconv.FormatInt(callback.ResultCode, 10)
	resultDescription := strings.TrimSpace(callback.ResultDesc)

	if callback.ResultCode != 0 {
		if err := s.repository.UpdateSTKFailed(
			ctx,
			stkRequest.ID,
			resultCode,
			resultDescription,
		); err != nil {
			return nil, err
		}

		return s.repository.GetSTKRequest(
			ctx,
			stkRequest.ID,
		)
	}

	mpesaReceiptNumber, transactionDate := extractSTKCallbackMetadata(
		callback,
	)

	if mpesaReceiptNumber == nil {
		return nil, errors.New(
			"successful mpesa callback missing receipt number",
		)
	}

	if err := s.repository.UpdateSTKCompleted(
		ctx,
		stkRequest.ID,
		resultCode,
		resultDescription,
		mpesaReceiptNumber,
		transactionDate,
	); err != nil {
		return nil, err
	}

	if stkRequest.PaymentID != nil &&
		*stkRequest.PaymentID != "" {

		_, err := s.paymentRepo.Confirm(
			ctx,
			stkRequest.TenantID,
			stkRequest.BranchID,
			*stkRequest.PaymentID,
			payment.ConfirmPaymentRequest{
				MPesaReceiptNumber: mpesaReceiptNumber,
			},
		)

		if err != nil {
			return nil, fmt.Errorf(
				"failed to confirm mpesa customer payment: %w",
				err,
			)
		}

		_, err = s.receiptRepo.Create(
			ctx,
			stkRequest.TenantID,
			stkRequest.BranchID,
			stkRequest.SaleID,
			stkRequest.PaymentID,
		)

		if err != nil {
			return nil, fmt.Errorf(
				"failed to create mpesa receipt: %w",
				err,
			)
		}
	}

	return s.repository.GetSTKRequest(
		ctx,
		stkRequest.ID,
	)
}

func (s *Service) processSubscriptionSTKCallback(
	ctx context.Context,
	stkRequest *SubscriptionSTKRequest,
	callback STKCallback,
) (*STKRequest, error) {
	if stkRequest == nil {
		return nil, ErrSTKRequestNotFound
	}

	resultCode := strconv.FormatInt(callback.ResultCode, 10)
	resultDescription := strings.TrimSpace(callback.ResultDesc)

	if callback.ResultCode != 0 {
		if err := s.repository.UpdateSubscriptionSTKFailed(
			ctx,
			stkRequest.ID,
			resultCode,
			resultDescription,
		); err != nil {
			return nil, err
		}

		if err := s.subscriptionRepo.MarkPaymentFailed(
			ctx,
			stkRequest.SubscriptionPaymentID,
			resultDescription,
		); err != nil {
			return nil, fmt.Errorf(
				"failed to mark subscription payment failed: %w",
				err,
			)
		}

		return subscriptionSTKAsSTKRequest(stkRequest), nil
	}

	mpesaReceiptNumber, transactionDate := extractSTKCallbackMetadata(
		callback,
	)

	if mpesaReceiptNumber == nil {
		return nil, errors.New(
			"successful mpesa subscription callback missing receipt number",
		)
	}

	if err := s.repository.UpdateSubscriptionSTKCompleted(
		ctx,
		stkRequest.ID,
		resultCode,
		resultDescription,
		mpesaReceiptNumber,
		transactionDate,
	); err != nil {
		return nil, err
	}

	if err := s.subscriptionRepo.MarkPaymentCompleted(
		ctx,
		stkRequest.SubscriptionPaymentID,
		*mpesaReceiptNumber,
	); err != nil {
		return nil, fmt.Errorf(
			"failed to complete subscription payment: %w",
			err,
		)
	}

	return subscriptionSTKAsSTKRequest(stkRequest), nil
}

func extractSTKCallbackMetadata(
	callback STKCallback,
) (*string, *time.Time) {
	var (
		mpesaReceiptNumber *string
		transactionDate    *time.Time
	)

	if callback.CallbackMetadata == nil {
		return nil, nil
	}

	for _, item := range callback.CallbackMetadata.Item {
		switch strings.TrimSpace(item.Name) {
		case "MpesaReceiptNumber":
			if value, ok := item.Value.(string); ok {
				value = strings.TrimSpace(value)

				if value != "" {
					mpesaReceiptNumber = &value
				}
			}

		case "TransactionDate":
			if value, ok := item.Value.(float64); ok {
				raw := strconv.FormatInt(
					int64(value),
					10,
				)

				parsed, err := time.Parse(
					"20060102150405",
					raw,
				)
				if err == nil {
					transactionDate = &parsed
				}
			}
		}
	}

	return mpesaReceiptNumber, transactionDate
}

func subscriptionSTKAsSTKRequest(
	request *SubscriptionSTKRequest,
) *STKRequest {
	if request == nil {
		return nil
	}

	return &STKRequest{
		ID:                     request.ID,
		TenantID:               request.TenantID,
		PhoneNumber:            request.PhoneNumber,
		Amount:                 request.Amount,
		AccountReference:       request.AccountReference,
		TransactionDescription: request.TransactionDescription,
		MerchantRequestID:      request.MerchantRequestID,
		CheckoutRequestID:      request.CheckoutRequestID,
		ResponseCode:           request.ResponseCode,
		ResponseDescription:    request.ResponseDescription,
		CustomerMessage:        request.CustomerMessage,
		Status:                 request.Status,
		CreatedAt:              request.CreatedAt,
		UpdatedAt:              request.UpdatedAt,
	}
}

// InitiateSubscriptionSTKPush starts an M-Pesa STK Push
// for a CyberSaaS subscription payment.
func (s *Service) InitiateSubscriptionSTKPush(
	ctx context.Context,
	tenantID string,
	phoneNumber string,
) (*SubscriptionSTKRequest, error) {
	if _, err := uuid.Parse(tenantID); err != nil {
		return nil, errors.New("invalid tenant id")
	}

	subscriptionRecord, err := s.subscriptionRepo.GetSubscriptionByTenant(
		ctx,
		tenantID,
	)
	if err != nil {
		return nil, err
	}

	if subscriptionRecord.IsLifetime {
		return nil, errors.New(
			"lifetime subscription does not require payment",
		)
	}

	if subscriptionRecord.PlanID == nil {
		return nil, errors.New(
			"subscription plan is not configured",
		)
	}

	plan, err := s.subscriptionRepo.GetPlan(
		ctx,
		*subscriptionRecord.PlanID,
	)
	if err != nil {
		return nil, err
	}

	if plan.IsLifetime {
		return nil, errors.New(
			"lifetime subscription does not require payment",
		)
	}

	amountText := strings.TrimSpace(plan.MonthlyPrice)

	if amountText == "" || amountText == "0" {
		return nil, errors.New(
			"subscription plan has no monthly price",
		)
	}

	formattedPhone, err := normalizeMpesaPhone(phoneNumber)
	if err != nil {
		return nil, err
	}

	amountFloat, err := strconv.ParseFloat(amountText, 64)
	if err != nil {
		return nil, errors.New(
			"invalid subscription amount",
		)
	}

	numericAmount := int64(amountFloat)

	if numericAmount <= 0 {
		return nil, errors.New(
			"subscription amount must be greater than zero",
		)
	}

	subscriptionPayment, err :=
		s.subscriptionRepo.CreateSubscriptionPayment(
			ctx,
			subscriptionRecord.ID,
			tenantID,
			amountText,
			formattedPhone,
			"mpesa_stk",
		)
	if err != nil {
		return nil, err
	}

	config, err := LoadPlatformDarajaConfig()
	if err != nil {
		return nil, err
	}

	stkID := uuid.New().String()

	err = s.repository.CreateSubscriptionSTKRequest(
		ctx,
		CreateSubscriptionSTKRequestParams{
			ID:                     stkID,
			TenantID:               tenantID,
			SubscriptionID:         subscriptionRecord.ID,
			SubscriptionPaymentID:  subscriptionPayment.ID,
			PhoneNumber:            formattedPhone,
			Amount:                 amountText,
			AccountReference:       config.AccountReference,
			TransactionDescription: "CyberSaaS subscription payment",
		},
	)
	if err != nil {
		return nil, err
	}

	darajaClient := NewDarajaClient(config)

	response, err := darajaClient.InitiateSTKPush(
		ctx,
		STKPushRequest{
			Amount:           numericAmount,
			PhoneNumber:      formattedPhone,
			AccountReference: config.AccountReference,
			TransactionDesc:  "CyberSaaS subscription payment",
		},
	)
	if err != nil {
		_ = s.repository.UpdateSubscriptionSTKFailed(
			ctx,
			stkID,
			"",
			err.Error(),
		)

		return nil, err
	}

	err = s.repository.UpdateSubscriptionSTKAccepted(
		ctx,
		stkID,
		response.MerchantRequestID,
		response.CheckoutRequestID,
		response.ResponseCode,
		response.ResponseDescription,
		response.CustomerMessage,
	)
	if err != nil {
		return nil, err
	}

	err = s.subscriptionRepo.SetSubscriptionPaymentCheckoutRequestID(
		ctx,
		subscriptionPayment.ID,
		response.CheckoutRequestID,
	)
	if err != nil {
		return nil, err
	}

	merchantRequestID := response.MerchantRequestID
	checkoutRequestID := response.CheckoutRequestID
	responseCode := response.ResponseCode
	responseDescription := response.ResponseDescription
	customerMessage := response.CustomerMessage

	return &SubscriptionSTKRequest{
		ID:                     stkID,
		TenantID:               tenantID,
		SubscriptionID:         subscriptionRecord.ID,
		SubscriptionPaymentID:  subscriptionPayment.ID,
		PhoneNumber:            formattedPhone,
		Amount:                 amountText,
		AccountReference:       config.AccountReference,
		TransactionDescription: "CyberSaaS subscription payment",
		MerchantRequestID:      &merchantRequestID,
		CheckoutRequestID:      &checkoutRequestID,
		ResponseCode:           &responseCode,
		ResponseDescription:    &responseDescription,
		CustomerMessage:        &customerMessage,
		Status:                 "accepted",
		CreatedAt:              time.Now().UTC(),
		UpdatedAt:              time.Now().UTC(),
	}, nil
}
