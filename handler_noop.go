package smtpmock

// NOOP command handler
type handlerNoop struct {
}

// NOOP command handler builder. Returns pointer to new handlerNoop structure
func newHandlerNoop() *handlerNoop {
	return &handlerNoop{}
}

// NOOP handler methods

// Main processNOOP handler runner
func (session *session) processNOOP(request string) {
	config := session.config
	handler := newHandlerNoop()
	if handler.isInvalidRequest(request) {
		return
	}

	session.message.noop = true
	session.writeResponse(config.msgNoopReceived, config.responseDelayNoop)
}

// Invalid NOOP command predicate. Returns true when request is invalid, otherwise returns false
func (handler *handlerNoop) isInvalidRequest(request string) bool {
	return !matchRegex(request, validNoopCmdRegexPattern)
}
