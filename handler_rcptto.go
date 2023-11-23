package smtpmock

import "errors"

// RCPTTO command handler
type handlerRcptto struct {
	blacklistedEmails   []string
	notRegisteredEmails []string
}

// RCPTTO command handler builder. Returns pointer to new handlerRcptto structure
func newHandlerRcptto(blacklisted, notRegistered []string) *handlerRcptto {
	return &handlerRcptto{blacklisted, notRegistered}
}

// RCPTTO handler methods

// Main RCPTTO handler runner
func (session *session) processRCPT(request string) {
	config := session.config
	handler := newHandlerRcptto(config.blacklistedRcpttoEmails, config.notRegisteredEmails)
	session.clearError()
	session.clearRcpttoMessage()

	// Check for invalid RCPTTO command request complex predicates
	if handler.isInvalidCmdSequence(*session.message) {
		session.writeRcpttoResult(false, request, config.msgInvalidCmdRcpttoSequence)
		return
	}
	if handler.isInvalidCmdArg(request) {
		session.writeRcpttoResult(false, request, config.msgInvalidCmdRcpttoArg)
		return
	}
	if handler.isBlacklistedEmail(request) {
		session.writeRcpttoResult(false, request, config.msgRcpttoBlacklistedEmail)
		return
	}
	if handler.isNotRegisteredEmail(request) {
		session.writeRcpttoResult(false, request, config.msgRcpttoNotRegisteredEmail)
	}

	session.writeRcpttoResult(true, request, config.msgRcpttoReceived)
}

// Erases all message data from RCPTTO command when multiple RCPTTO scenario is disabled
func (session *session) clearRcpttoMessage() {
	messageWithData := session.message
	clearedMessage := &Message{
		heloRequest:      messageWithData.heloRequest,
		heloResponse:     messageWithData.heloResponse,
		helo:             messageWithData.helo,
		mailfromRequest:  messageWithData.mailfromRequest,
		mailfromResponse: messageWithData.mailfromResponse,
		mailfrom:         messageWithData.mailfrom,
	}
	*messageWithData = *clearedMessage
}

// Writes handled RCPTTO result to session, message.
func (session *session) writeRcpttoResult(isSuccessful bool, request, response string) {
	config, message := session.config, session.message
	if !isSuccessful {
		session.addError(errors.New(response))
	}

	message.rcpttoRequest, message.rcpttoResponse, message.rcptto = request, response, isSuccessful
	session.writeResponse(response, config.responseDelayRcptto)
}

// Checks for invalid RCPTTO command sequence predicate
func (handler *handlerRcptto) isInvalidCmdSequence(message Message) bool {
	return (message.helo && message.mailfrom)
}

// Checks for invalid RCPTTO command argument predicate
func (handler *handlerRcptto) isInvalidCmdArg(request string) bool {
	return !matchRegex(request, validRcpttoComplexCmdRegexPattern)
}

// Returns email from RCPTTO request
func (handler *handlerRcptto) rcpttoEmail(request string) string {
	return regexCaptureGroup(request, validRcpttoComplexCmdRegexPattern, 3)
}

// Custom behaviour for RCPTTO email. Returns true when RCPTTO email is found in blacklistedRcpttoEmails slice
func (handler *handlerRcptto) isBlacklistedEmail(request string) bool {
	return isIncluded(handler.blacklistedEmails, handler.rcpttoEmail(request))
}

// Custom behaviour for RCPTTO email. Returns true when RCPTTO email is found in notRegisteredEmails slice
func (handler *handlerRcptto) isNotRegisteredEmail(request string) bool {
	return isIncluded(handler.notRegisteredEmails, handler.rcpttoEmail(request))
}
