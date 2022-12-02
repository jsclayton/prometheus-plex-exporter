package plex

// ErrorInvalidToken a constant to help check invalid token errors
const (
	ErrorInvalidToken       = "invalid token"
	ErrorNotAuthorized      = "you are not authorized to access that server"
	ErrorCommon             = "error: %s"
	ErrorKeyIsRequired      = "key is required"
	ErrorTitleRequired      = "a title is required"
	ErrorServerReplied      = "server replied with %d status code"
	ErrorMissingSessionKey  = "missing sessionKey"
	ErrorUrlTokenRequired   = "url or a token is required"
	ErrorServer             = "server error: %s"
	ErrorPINNotAuthorized   = "pin is not authorized yet"
	ErrorLinkAccount        = "failed to link account: %s"
	ErrorFailedToSetWebhook = "failed to set webhook"
	ErrorWebhook            = "webhook error: %s"
)
