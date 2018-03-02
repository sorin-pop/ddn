package errs

// Constants for errors.
const (
	// Service related
	JSONDecodeFailed       = "ERR_JSON_DECODE_FAILED"
	JSONEncodeFailed       = "ERR_JSON_ENCODE_FAILED"
	MissingUserCookie      = "ERR_MISSING_USER_COOKIE"
	MissingParameters      = "ERR_MISSING_PARAMETERS"
	AccessDenied           = "ERR_ACCESS_DENIED"
	InvalidURL             = "ERR_INVALID_URL"
	UnknownParameter       = "ERR_UNKNOWN_PARAMETER"
	AgentNotFound          = "ERR_AGENT_NOT_FOUND"
	NoAgentsAvailable      = "ERR_NO_AGENTS_AVAILABLE"
	FailedListingDirectory = "ERR_DIR_LIST_FAILED"
	NoFoldersMounted       = "ERR_NO_FOLDER_MOUNTED"
	FileIOFailed           = "ERR_FILE_IO_FAILED"

	// Database related
	PersistFailed  = "ERR_DATABASE_PERSIST_FAILED"
	CreateFailed   = "ERR_DATABASE_CREATE_FAILED"
	ImportFailed   = "ERR_DATABASE_IMPORT_FAILED"
	DropFailed     = "ERR_DATABASE_DROP_FAILED"
	ExportFailed   = "ERR_DATABASE_EXPORT_FAILED"
	QueryFailed    = "ERR_DATABASE_QUERY_FAILED"
	UpdateFailed   = "ERR_DATABASE_UPDATE_FAILED"
	QueryNoResults = "ERR_DATABASE_NO_RESULT"
)
