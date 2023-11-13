package smtpmock

// QUIT command handler
type handlerQuit struct {
}

// QUIT command handler builder. Returns pointer to new handlerQuit structure
func newHandlerQuit() *handlerQuit {
	return &handlerQuit{}
}

// QUIT handler methods

// Main QUIT handler runner
func (session *session) runQuitHandler(request string) {
	config, message := session.config, session.message
	handler := newHandlerQuit()
	if handler.isInvalidRequest(request) {
		return
	}

	message.quitSent = true
	session.writeResponse(config.msgQuitCmd, config.responseDelayQuit)
}

// Invalid QUIT command predicate. Returns true when request is invalid, otherwise returns false
func (handler *handlerQuit) isInvalidRequest(request string) bool {
	return !matchRegex(request, validQuitCmdRegexPattern)
}
