package constants

const (
	// Roles
	ENUM_ROLE_ADMIN    = "admin"
	ENUM_ROLE_CLIENT   = "client"
	ENUM_ROLE_DESIGNER = "designer"

	// Run Mode
	ENUM_RUN_PRODUCTION = "production"
	ENUM_RUN_TESTING    = "testing"

	// Pagination
	ENUM_PAGINATION_PER_PAGE = 10
	ENUM_PAGINATION_PAGE     = 1

	// Dependency Injection Keys
	DB         = "db"
	JWTService = "JWTService"

	// Project Request Status
	PROJECT_REQUEST_STATUS_OPEN        = "Open"
	PROJECT_REQUEST_STATUS_IN_PROGRESS = "InProgress"
	PROJECT_REQUEST_STATUS_COMPLETED   = "Completed"
	PROJECT_REQUEST_STATUS_CANCELLED   = "Cancelled"

	// Quotation Status
	QUOTATION_STATUS_PENDING  = "Pending"
	QUOTATION_STATUS_ACCEPTED = "Accepted"
	QUOTATION_STATUS_REJECTED = "Rejected"

	// Order Payment Status
	ORDER_PAYMENT_STATUS_UNPAID   = "Unpaid"
	ORDER_PAYMENT_STATUS_PAID     = "Paid"
	ORDER_PAYMENT_STATUS_REFUNDED = "Refunded"

	// Order Work Status
	ORDER_WORK_STATUS_ACTIVE    = "Active"
	ORDER_WORK_STATUS_COMPLETED = "Completed"
	ORDER_WORK_STATUS_CANCELLED = "Cancelled"

	// Payment Status (Midtrans)
	PAYMENT_STATUS_PENDING    = "Pending"
	PAYMENT_STATUS_SETTLEMENT = "Settlement"
	PAYMENT_STATUS_EXPIRE     = "Expire"
	PAYMENT_STATUS_CANCEL     = "Cancel"
	PAYMENT_STATUS_DENY       = "Deny"

	// Message Type
	MESSAGE_TYPE_TEXT  = "Text"
	MESSAGE_TYPE_IMAGE = "Image"
	MESSAGE_TYPE_FILE  = "File"
)
