package comfy_errors

type ResponseErrorCode uint8

const (
	UnknownError ResponseErrorCode = iota
	LoginRequestError
	SessionStoreError
	EntityNotFoundError
	EntityAlreadyExistsError
	EntityValidationError
)
