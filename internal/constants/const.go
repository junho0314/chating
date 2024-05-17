package constants

const (
	ServerName = "chating_service"
)

const (
	AccountIdField = "user_id"

	// 로그인
	AccountRoleCodeKey         = "account_role_code"
	ResponseCodeKey            = "response_code"
	DomainCodeKey              = "domain_code"
	PasswordExpirationReminder = "password_expiration_reminder"

	AccountDefaultStatus  = 1
	AccountStatusInactive = 0
	AccountStatusActive   = 1
	AccountStatusBlocked  = 2

	AuthDefaultStatus     = 0
	AuthStatusUncertified = 0
	AuthStatusCertified   = 1

	LoginRetryMaxCount = 5

	SettlementManagerCode = 1
	OperationsOfficerCode = 2

	CompanyAdminCodeMEV = 1000
	CompanyAdminCodeHEC = 2000

	BusinessNumberLength = 10

	PasswordMinLength = 8
	PasswordMaxLength = 20
)

// redis key
const (
	RefreshTokenKey = "refresh_"
)
