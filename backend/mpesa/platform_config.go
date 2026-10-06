package mpesa

import (
	"errors"
	"os"
	"strings"
)

func LoadPlatformDarajaConfig() (DarajaConfig, error) {
	environment := strings.ToLower(
		strings.TrimSpace(
			os.Getenv("MPESA_PLATFORM_ENVIRONMENT"),
		),
	)

	if environment == "" {
		environment = string(DarajaSandbox)
	}

	switch DarajaEnvironment(environment) {
	case DarajaSandbox, DarajaProduction:
	default:
		return DarajaConfig{}, errors.New(
			"MPESA_PLATFORM_ENVIRONMENT must be sandbox or production",
		)
	}

	consumerKey := strings.TrimSpace(
		os.Getenv("MPESA_PLATFORM_CONSUMER_KEY"),
	)

	consumerSecret := strings.TrimSpace(
		os.Getenv("MPESA_PLATFORM_CONSUMER_SECRET"),
	)

	passkey := strings.TrimSpace(
		os.Getenv("MPESA_PLATFORM_PASSKEY"),
	)

	businessShortCode := strings.TrimSpace(
		os.Getenv("MPESA_PLATFORM_BUSINESS_SHORT_CODE"),
	)

	tillNumber := strings.TrimSpace(
		os.Getenv("MPESA_PLATFORM_TILL_NUMBER"),
	)

	paybillNumber := strings.TrimSpace(
		os.Getenv("MPESA_PLATFORM_PAYBILL_NUMBER"),
	)

	callbackURL := strings.TrimSpace(
		os.Getenv("MPESA_PLATFORM_CALLBACK_URL"),
	)

	accountReference := strings.TrimSpace(
		os.Getenv("MPESA_PLATFORM_ACCOUNT_REFERENCE"),
	)

	if consumerKey == "" {
		return DarajaConfig{}, errors.New(
			"MPESA_PLATFORM_CONSUMER_KEY is required",
		)
	}

	if consumerSecret == "" {
		return DarajaConfig{}, errors.New(
			"MPESA_PLATFORM_CONSUMER_SECRET is required",
		)
	}

	if businessShortCode == "" {
		return DarajaConfig{}, errors.New(
			"MPESA_PLATFORM_BUSINESS_SHORT_CODE is required",
		)
	}

	if tillNumber == "" && paybillNumber == "" {
		return DarajaConfig{}, errors.New(
			"either MPESA_PLATFORM_TILL_NUMBER or MPESA_PLATFORM_PAYBILL_NUMBER is required",
		)
	}

	if callbackURL == "" {
		return DarajaConfig{}, errors.New(
			"MPESA_PLATFORM_CALLBACK_URL is required",
		)
	}

	if !IsValidCallbackURL(callbackURL) {
		return DarajaConfig{}, errors.New(
			"MPESA_PLATFORM_CALLBACK_URL must be a valid HTTPS URL",
		)
	}

	if accountReference == "" {
		accountReference = "CYBERSAAS"
	}

	return DarajaConfig{
		Environment:       DarajaEnvironment(environment),
		ConsumerKey:       consumerKey,
		ConsumerSecret:    consumerSecret,
		Passkey:           passkey,
		BusinessShortCode: businessShortCode,
		TillNumber:        tillNumber,
		PaybillNumber:     paybillNumber,
		CallbackURL:       callbackURL,
		AccountReference:  accountReference,
	}, nil
}
