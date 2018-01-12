package errs

// Constants for errors.
const (
	// Service related
	JSONDecodeFailed  = "ERR_JSON_DECODE_FAILED"
	JSONEncodeFailed  = "ERR_JSON_ENCODE_FAILED"
	MissingUserCookie = "ERR_MISSING_USER_COOKIE"
	MissingParameters = "ERR_MISSING_PARAMETERS"
	AccessDenied      = "ERR_ACCESS_DENIED"
	InvalidURL        = "ERR_INVALID_URL"
	UnknownParameter  = "ERR_UNKNOWN_PARAMETER"
	AgentNotFound     = "ERR_AGENT_NOT_FOUND"

	// Database related
	PersistFailed  = "ERR_DATABASE_PERSIST_FAILED"
	CreateFailed   = "ERR_DATABASE_CREATE_FAILED"
	DeleteFailed   = "ERR_DATABASE_DELETE_FAILED"
	QueryFailed    = "ERR_DATABASE_QUERY_FAILED"
	UpdateFailed   = "ERR_DATABASE_UPDATE_FAILED"
	QueryNoResults = "ERR_DATABASE_NO_RESULT"
)
