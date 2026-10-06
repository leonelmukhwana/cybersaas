package terminalpayment

import (
	"cybersaas/backend/middleware"

	"github.com/gin-gonic/gin"
)

func RegisterRoutes(
	router *gin.RouterGroup,
	handler *Handler,
	terminalAuthenticator middleware.TerminalAuthenticator,
) {
	terminalPayments := router.Group("/terminal-payments")

	terminalPayments.Use(
		middleware.TerminalAuthMiddleware(
			terminalAuthenticator,
		),
	)

	// Cash payment
	terminalPayments.POST(
		"/cash",
		handler.CreateCashPayment,
	)

	// M-Pesa STK Push
	terminalPayments.POST(
		"/mpesa",
		handler.CreateMpesaPayment,
	)

	// M-Pesa payment status
	terminalPayments.GET(
		"/mpesa/:paymentID/status",
		handler.GetMpesaPaymentStatus,
	)
}
