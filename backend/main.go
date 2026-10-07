package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"strconv"
	"time"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"

	"cybersaas/backend/attendant"
	"cybersaas/backend/auth"
	"cybersaas/backend/branch"
	"cybersaas/backend/config"
	"cybersaas/backend/customer"
	"cybersaas/backend/database"
	"cybersaas/backend/email"
	"cybersaas/backend/expenses"
	"cybersaas/backend/middleware"
	"cybersaas/backend/mpesa"
	"cybersaas/backend/payment"
	"cybersaas/backend/platform"
	"cybersaas/backend/receipt"
	"cybersaas/backend/report"
	"cybersaas/backend/sale"
	"cybersaas/backend/service"
	"cybersaas/backend/session"
	"cybersaas/backend/subscription"
	"cybersaas/backend/sync"
	"cybersaas/backend/tenant"
	"cybersaas/backend/terminal"
	"cybersaas/backend/terminalhealth"
	"cybersaas/backend/terminalpayment"
)

func main() {
	// Load .env during local development.
	if err := godotenv.Load(); err != nil {
		log.Println("No .env file found, using environment variables")
	}

	// Load application configuration.
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("Configuration error: %v", err)
	}

	// Connect to PostgreSQL.
	db, err := database.Connect(cfg.DatabaseURL)
	if err != nil {
		log.Fatalf("Database error: %v", err)
	}
	defer db.Close()

	router := gin.Default()

	// --------------------------------------------------
	// CORS
	// --------------------------------------------------

	router.Use(cors.New(cors.Config{
		AllowOrigins: []string{
			"http://localhost:3000",
		},
		AllowMethods: []string{
			http.MethodGet,
			http.MethodPost,
			http.MethodPut,
			http.MethodPatch,
			http.MethodDelete,
			http.MethodOptions,
		},
		AllowHeaders: []string{
			"Origin",
			"Content-Type",
			"Accept",
			"Authorization",
			"X-Client-Operation-ID",
		},
		ExposeHeaders: []string{
			"Content-Length",
		},
		AllowCredentials: true,
		MaxAge:           12 * time.Hour,
	}))

	if err := router.SetTrustedProxies(nil); err != nil {
		log.Fatalf("Failed to configure trusted proxies: %v", err)
	}

	// --------------------------------------------------
	// Health check
	// --------------------------------------------------

	router.GET("/health", func(c *gin.Context) {
		ctx, cancel := context.WithTimeout(
			c.Request.Context(),
			3*time.Second,
		)
		defer cancel()

		if err := db.Ping(ctx); err != nil {
			c.JSON(http.StatusServiceUnavailable, gin.H{
				"status":   "error",
				"service":  "cybercafe-api",
				"database": "disconnected",
			})
			return
		}

		c.JSON(http.StatusOK, gin.H{
			"status":   "ok",
			"service":  "cybercafe-api",
			"database": "connected",
		})
	})

	// --------------------------------------------------
	// JWT configuration
	// --------------------------------------------------

	jwtSecret := os.Getenv("JWT_SECRET")

	if len(jwtSecret) < 32 {
		log.Fatal("JWT_SECRET must be at least 32 characters")
	}

	jwtTTL := 15 * time.Minute

	if value := os.Getenv("JWT_TTL_MINUTES"); value != "" {
		minutes, err := strconv.Atoi(value)
		if err != nil || minutes <= 0 {
			log.Fatal("JWT_TTL_MINUTES must be a positive number")
		}

		jwtTTL = time.Duration(minutes) * time.Minute
	}

	resetURL := os.Getenv("RESET_URL")

	if resetURL == "" {
		resetURL = "http://localhost:3000"
	}

	// --------------------------------------------------
	// Authentication
	// --------------------------------------------------

	tokenManager := auth.NewTokenManager(
		jwtSecret,
		jwtTTL,
	)

	authRepository := auth.NewRepository(db)

	emailService := email.NewService(email.Config{
		APIKey:    os.Getenv("RESEND_API_KEY"),
		FromEmail: cfg.SMTPFromEmail,
		FromName:  cfg.SMTPFromName,
	})

	authService := auth.NewService(
		authRepository,
		tokenManager,
		resetURL,
	)

	authService.SetEmailSender(emailService)

	authHandler := auth.NewHandler(authService)

	// --------------------------------------------------
	// API routes
	// --------------------------------------------------

	api := router.Group("/api")

	// Authentication routes.
	auth.RegisterRoutes(api, authHandler)

	// --------------------------------------------------
	// Platform Admin
	// --------------------------------------------------

	platformRepository := platform.NewRepository(db)

	platformService := platform.NewService(
		platformRepository,
	)

	platformHandler := platform.NewHandler(
		platformService,
	)

	platform.RegisterRoutes(
		api,
		platformHandler,
		tokenManager,
	)

	// --------------------------------------------------
	// Subscription and Payment
	// --------------------------------------------------

	subscriptionRepository := subscription.NewRepository(db)

	subscriptionService := subscription.NewService(
		subscriptionRepository,
	)

	subscriptionHandler := subscription.NewHandler(
		subscriptionService,
	)

	subscription.RegisterRoutes(
		api,
		subscriptionHandler,
		tokenManager,
	)

	// --------------------------------------------------
	// Subscription access middleware
	//
	// This is used ONLY by web/business APIs that require
	// an active trial or paid subscription.
	//
	// Subscription-management routes remain accessible after
	// expiry so the owner can reactivate the account.
	//
	// Terminal APIs are NOT protected by this middleware.
	// --------------------------------------------------

	requireActiveSubscription := middleware.RequireActiveSubscription(
		subscriptionService,
	)

	// --------------------------------------------------
	// Branches
	// --------------------------------------------------

	branchRepository := branch.NewRepository(db)

	branchService := branch.NewService(
		branchRepository,
		subscriptionService,
	)

	branchHandler := branch.NewHandler(branchService)

	branch.RegisterRoutes(
		api,
		branchHandler,
		tokenManager,
		requireActiveSubscription,
	)

	// --------------------------------------------------
	// Attendant
	// --------------------------------------------------

	attendantRepository := attendant.NewRepository(db)
	attendantService := attendant.NewService(attendantRepository)
	attendantHandler := attendant.NewHandler(attendantService)

	attendant.RegisterRoutes(
		api,
		attendantHandler,
		tokenManager,
		requireActiveSubscription,
	)

	// --------------------------------------------------
	// Terminals
	// --------------------------------------------------

	terminalRepository := terminal.NewRepository(db)

	terminalService := terminal.NewService(
		terminalRepository,
		subscriptionService,
	)
	terminalHandler := terminal.NewHandler(
		terminalService,
	)

	terminal.RegisterRoutes(
		api,
		terminalHandler,
		tokenManager,
		terminalService,
		requireActiveSubscription,
	)

	// --------------------------------------------------
	// Customer registration and handling
	// --------------------------------------------------

	customerRepository := customer.NewRepository(db)

	customerService := customer.NewService(
		customerRepository,
	)

	customerHandler := customer.NewHandler(
		customerService,
	)

	customer.RegisterRoutes(
		api,
		customerHandler,
		tokenManager,
		terminalService,
		requireActiveSubscription,
	)

	// --------------------------------------------------
	// Services
	// --------------------------------------------------

	serviceRepository := service.NewRepository(db)
	serviceManager := service.NewService(serviceRepository)
	serviceHandler := service.NewHandler(serviceManager)

	service.RegisterRoutes(
		api,
		serviceHandler,
		tokenManager,
		requireActiveSubscription,
	)

	// --------------------------------------------------
	// Sales
	// --------------------------------------------------

	saleRepository := sale.NewRepository(db)
	saleService := sale.NewService(saleRepository)
	saleHandler := sale.NewHandler(saleService)

	sale.RegisterRoutes(
		api,
		saleHandler,
		tokenManager,
		requireActiveSubscription,
	)

	// --------------------------------------------------
	// Payments
	// --------------------------------------------------

	paymentRepository := payment.NewRepository(db)
	paymentService := payment.NewService(paymentRepository)
	paymentHandler := payment.NewHandler(paymentService)

	payment.RegisterRoutes(
		api,
		paymentHandler,
		tokenManager,
		requireActiveSubscription,
	)

	// --------------------------------------------------
	// Receipts
	// --------------------------------------------------

	receiptRepository := receipt.NewRepository(db)
	receiptService := receipt.NewService(receiptRepository)
	receiptHandler := receipt.NewHandler(receiptService)

	receipt.RegisterRoutes(
		api,
		receiptHandler,
		tokenManager,
		requireActiveSubscription,
	)
	// --------------------------------------------------
	// M-Pesa
	//
	// IMPORTANT:
	// This must be initialized BEFORE terminalpayment,
	// because terminalpayment.Service depends on mpesa.Service.
	// --------------------------------------------------

	mpesaEncryptor, err := mpesa.NewEncryptor()
	if err != nil {
		log.Fatal(err)
	}

	mpesaRepository := mpesa.NewRepository(db)

	mpesaService := mpesa.NewService(
		mpesaRepository,
		mpesaEncryptor,
		subscriptionRepository,
		paymentRepository,
		receiptRepository,
	)

	mpesaHandler := mpesa.NewHandler(mpesaService)

	mpesa.RegisterRoutes(
		api,
		mpesaHandler,
		tokenManager,
		requireActiveSubscription,
	)

	// --------------------------------------------------
	// Sessions
	// --------------------------------------------------

	sessionRepository := session.NewRepository(db)
	sessionService := session.NewService(sessionRepository)
	sessionHandler := session.NewHandler(sessionService)

	session.RegisterRoutes(
		api,
		sessionHandler,
		tokenManager,
		terminalService,
		requireActiveSubscription,
	)

	// --------------------------------------------------
	// Terminal Payments
	// --------------------------------------------------

	terminalPaymentRepository := terminalpayment.NewRepository(db)

	terminalPaymentService := terminalpayment.NewService(
		terminalPaymentRepository,
		mpesaService,
	)

	terminalPaymentHandler := terminalpayment.NewHandler(
		terminalPaymentService,
	)

	terminalpayment.RegisterRoutes(
		api,
		terminalPaymentHandler,
		terminalService,
	)

	// --------------------------------------------------
	// Computer Health
	// --------------------------------------------------

	terminalHealthRepository := terminalhealth.NewRepository(db)

	terminalHealthService := terminalhealth.NewService(
		terminalHealthRepository,
	)

	terminalHealthHandler := terminalhealth.NewHandler(
		terminalHealthService,
	)

	terminalhealth.RegisterRoutes(
		api,
		terminalHealthHandler,
		tokenManager,
		terminalService,
		requireActiveSubscription,
	)
	// --------------------------------------------------
	// Sync
	// --------------------------------------------------

	syncRepository := sync.NewRepository(db)
	syncService := sync.NewService(syncRepository)
	syncHandler := sync.NewHandler(syncService)

	sync.RegisterRoutes(
		api,
		syncHandler,
		terminalRepository,
	)

	// --------------------------------------------------
	// Tenant / Business
	// --------------------------------------------------

	tenantRepository := tenant.NewRepository(
		db,
		subscriptionRepository,
	)

	tenantService := tenant.NewService(
		tenantRepository,
	)

	tenantHandler := tenant.NewHandler(
		tenantService,
	)

	tenant.RegisterRoutes(
		api,
		tenantHandler,
		tokenManager,
	)

	// --------------------------------------------------
	// Reports
	// --------------------------------------------------

	reportRepository := report.NewRepository(db)

	reportService := report.NewService(
		reportRepository,
	)

	reportHandler := report.NewHandler(
		reportService,
	)

	report.RegisterRoutes(
		api,
		report.RouteDependencies{
			Handler: reportHandler,
		},
		tokenManager,
		requireActiveSubscription,
	)

	// --------------------------------------------------
	// Expenses
	// --------------------------------------------------

	expenseRepository := expenses.NewRepository(db)

	expenseService := expenses.NewService(
		expenseRepository,
		customerRepository,
	)

	expenseHandler := expenses.NewHandler(
		expenseService,
	)

	expenses.RegisterRoutes(
		api,
		expenseHandler,
		tokenManager,
		requireActiveSubscription,
	)

	//cyber attendant report

	attendantReportRepository := report.NewAttendantRepository(db)

	attendantReportService := report.NewAttendantReportService(
		attendantReportRepository,
	)

	attendantReportHandler := report.NewAttendantReportHandler(
		attendantReportService,
	)

	report.RegisterAttendantRoutes(
		api,
		attendantReportHandler,
		tokenManager,
		requireActiveSubscription,
	)

	// --------------------------------------------------
	// Owner Reports
	// --------------------------------------------------

	ownerReportRepository := report.NewOwnerRepository(db)

	ownerReportService := report.NewOwnerReportService(
		ownerReportRepository,
	)

	ownerReportHandler := report.NewOwnerReportHandler(
		ownerReportService,
	)

	report.RegisterOwnerRoutes(
		api,
		ownerReportHandler,
		tokenManager,
		requireActiveSubscription,
	)

	// --------------------------------------------------
	// Start server
	// --------------------------------------------------

	log.Printf("Server running on port %s", cfg.Port)

	if err := router.Run(":" + cfg.Port); err != nil {
		log.Fatalf("Server failed: %v", err)
	}
}
