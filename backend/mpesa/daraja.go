package mpesa

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"sync"
	"time"
)

type DarajaEnvironment string

const (
	DarajaSandbox    DarajaEnvironment = "sandbox"
	DarajaProduction DarajaEnvironment = "production"
)

type DarajaConfig struct {
	Environment       DarajaEnvironment
	ConsumerKey       string
	ConsumerSecret    string
	Passkey           string
	BusinessShortCode string
	TillNumber        string
	PaybillNumber     string
	CallbackURL       string
	AccountReference  string
}

type STKPushRequest struct {
	Amount           int64
	PhoneNumber      string
	AccountReference string
	TransactionDesc  string
}

type STKPushResponse struct {
	MerchantRequestID   string `json:"MerchantRequestID"`
	CheckoutRequestID   string `json:"CheckoutRequestID"`
	ResponseCode        string `json:"ResponseCode"`
	ResponseDescription string `json:"ResponseDescription"`
	CustomerMessage     string `json:"CustomerMessage"`
}

type DarajaClient struct {
	config     DarajaConfig
	httpClient *http.Client

	mu          sync.Mutex
	accessToken string
	tokenExpiry time.Time
}

func NewDarajaClient(config DarajaConfig) *DarajaClient {
	return &DarajaClient{
		config: config,
		httpClient: &http.Client{
			Timeout: 30 * time.Second,
		},
	}
}

func NewDarajaClientWithHTTPClient(
	config DarajaConfig,
	httpClient *http.Client,
) *DarajaClient {
	if httpClient == nil {
		httpClient = &http.Client{
			Timeout: 30 * time.Second,
		}
	}

	return &DarajaClient{
		config:     config,
		httpClient: httpClient,
	}
}

func (c *DarajaClient) baseURL() string {
	if c.config.Environment == DarajaProduction {
		return "https://api.safaricom.co.ke"
	}

	return "https://sandbox.safaricom.co.ke"
}

func (c *DarajaClient) AccessToken(
	ctx context.Context,
) (string, error) {
	c.mu.Lock()
	defer c.mu.Unlock()

	if c.accessToken != "" &&
		time.Now().Before(c.tokenExpiry) {
		return c.accessToken, nil
	}

	if strings.TrimSpace(c.config.ConsumerKey) == "" {
		return "", errors.New("daraja consumer key is required")
	}

	if strings.TrimSpace(c.config.ConsumerSecret) == "" {
		return "", errors.New("daraja consumer secret is required")
	}

	credentials := base64.StdEncoding.EncodeToString(
		[]byte(
			c.config.ConsumerKey + ":" +
				c.config.ConsumerSecret,
		),
	)

	endpoint :=
		c.baseURL() +
			"/oauth/v1/generate?grant_type=client_credentials"

	req, err := http.NewRequestWithContext(
		ctx,
		http.MethodGet,
		endpoint,
		nil,
	)
	if err != nil {
		return "", err
	}

	req.Header.Set(
		"Authorization",
		"Basic "+credentials,
	)
	req.Header.Set(
		"Accept",
		"application/json",
	)

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return "", fmt.Errorf(
			"daraja oauth returned status %d",
			resp.StatusCode,
		)
	}

	var response struct {
		AccessToken string `json:"access_token"`
		ExpiresIn   string `json:"expires_in"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&response); err != nil {
		return "", err
	}

	if response.AccessToken == "" {
		return "", errors.New(
			"daraja returned an empty access token",
		)
	}

	expiry := 3600 * time.Second

	if response.ExpiresIn != "" {
		if seconds, err := strconv.Atoi(response.ExpiresIn); err == nil &&
			seconds > 0 {
			expiry = time.Duration(seconds) * time.Second
		}
	}

	c.accessToken = response.AccessToken

	// Refresh slightly before the real expiry.
	c.tokenExpiry = time.Now().Add(expiry - 30*time.Second)

	return c.accessToken, nil
}

func (c *DarajaClient) InitiateSTKPush(
	ctx context.Context,
	request STKPushRequest,
) (*STKPushResponse, error) {
	if request.Amount <= 0 {
		return nil, errors.New("stk amount must be greater than zero")
	}

	if strings.TrimSpace(request.PhoneNumber) == "" {
		return nil, errors.New("stk phone number is required")
	}

	if strings.TrimSpace(request.AccountReference) == "" {
		return nil, errors.New(
			"stk account reference is required",
		)
	}

	if strings.TrimSpace(request.TransactionDesc) == "" {
		return nil, errors.New(
			"stk transaction description is required",
		)
	}

	if strings.TrimSpace(c.config.BusinessShortCode) == "" {
		return nil, errors.New(
			"daraja business short code is required",
		)
	}

	if strings.TrimSpace(c.config.CallbackURL) == "" {
		return nil, errors.New(
			"daraja callback url is required",
		)
	}

	token, err := c.AccessToken(ctx)
	if err != nil {
		return nil, err
	}

	timestamp := time.Now().Format("20060102150405")

	password := base64.StdEncoding.EncodeToString(
		[]byte(
			c.config.BusinessShortCode +
				c.config.Passkey +
				timestamp,
		),
	)

	phone, err := normalizeDarajaPhone(
		request.PhoneNumber,
	)
	if err != nil {
		return nil, err
	}

	paymentType, partyB := c.resolvePaymentDestination()

	if partyB == "" {
		return nil, errors.New(
			"daraja till number or paybill number is required",
		)
	}

	payload := map[string]any{
		"BusinessShortCode": c.config.BusinessShortCode,
		"Password":          password,
		"Timestamp":         timestamp,
		"TransactionType":   paymentType,
		"Amount":            request.Amount,
		"PartyA":            phone,
		"PartyB":            partyB,
		"PhoneNumber":       phone,
		"CallBackURL":       c.config.CallbackURL,
		"AccountReference":  request.AccountReference,
		"TransactionDesc":   request.TransactionDesc,
	}

	body, err := json.Marshal(payload)
	if err != nil {
		return nil, err
	}

	endpoint :=
		c.baseURL() +
			"/mpesa/stkpush/v1/processrequest"

	req, err := http.NewRequestWithContext(
		ctx,
		http.MethodPost,
		endpoint,
		strings.NewReader(string(body)),
	)
	if err != nil {
		return nil, err
	}

	req.Header.Set(
		"Authorization",
		"Bearer "+token,
	)
	req.Header.Set(
		"Content-Type",
		"application/json",
	)
	req.Header.Set(
		"Accept",
		"application/json",
	)

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	responseBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return nil, fmt.Errorf(
			"daraja stk push returned status %d: %s",
			resp.StatusCode,
			strings.TrimSpace(string(responseBody)),
		)
	}

	var response STKPushResponse

	if err := json.Unmarshal(responseBody, &response); err != nil {
		return nil, err
	}

	if response.ResponseCode != "" &&
		response.ResponseCode != "0" {
		return nil, fmt.Errorf(
			"daraja stk push rejected: %s",
			response.ResponseDescription,
		)
	}

	return &response, nil
}
func (c *DarajaClient) resolvePaymentDestination() (
	string,
	string,
) {
	if strings.TrimSpace(c.config.TillNumber) != "" {
		return "CustomerBuyGoodsOnline",
			strings.TrimSpace(c.config.TillNumber)
	}

	if strings.TrimSpace(c.config.PaybillNumber) != "" {
		return "CustomerPayBillOnline",
			strings.TrimSpace(c.config.PaybillNumber)
	}

	return "", ""
}

func normalizeDarajaPhone(
	value string,
) (string, error) {
	phone := strings.TrimSpace(value)

	phone = strings.ReplaceAll(phone, " ", "")
	phone = strings.ReplaceAll(phone, "-", "")

	switch {
	case strings.HasPrefix(phone, "07") ||
		strings.HasPrefix(phone, "01"):
		if len(phone) != 10 {
			return "", errors.New(
				"invalid Kenyan phone number",
			)
		}

		phone = "254" + phone[1:]

	case strings.HasPrefix(phone, "254"):
		if len(phone) != 12 {
			return "", errors.New(
				"invalid Kenyan phone number",
			)
		}

	default:
		return "", errors.New(
			"invalid Kenyan phone number",
		)
	}

	for _, character := range phone {
		if character < '0' || character > '9' {
			return "", errors.New(
				"invalid Kenyan phone number",
			)
		}
	}

	return phone, nil
}

func IsValidCallbackURL(
	value string,
) bool {
	value = strings.TrimSpace(value)

	if value == "" {
		return false
	}

	parsed, err := url.Parse(value)
	if err != nil {
		return false
	}

	return parsed.Scheme == "https" &&
		parsed.Host != ""
}
