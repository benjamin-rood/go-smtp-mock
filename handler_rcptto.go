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
	if !session.config.multipleRcptto {
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
}

// RCPTTO message status resolver. Returns true when current RCPTTO status is true or
// when multiple RCPTTO scenario is enabled and message includes at least one successful
// RCPTTO response. Otherwise returns false
func (session *session) resolveMessageStatus(currentRcpttoStatus bool) bool {
	configuration, message := session.config, session.message
	multipleRcptto, msgRcpttoReceived := configuration.multipleRcptto, configuration.msgRcpttoReceived

	return currentRcpttoStatus || (multipleRcptto && message.isIncludesSuccessfulRcpttoResponse(msgRcpttoReceived))
}

// Writes handled RCPTTO result to session, message
func (session *session) writeRcpttoResult(isSuccessful bool, request, response string) {
	config, message := session.config, session.message
	if !isSuccessful {
		session.addError(errors.New(response))
	}

	message.rcpttoRequestResponse = append(message.rcpttoRequestResponse, []string{request, response})
	message.rcptto = session.resolveMessageStatus(isSuccessful)
	session.writeResponse(response, config.responseDelayRcptto)
}

// Invalid RCPTTO command sequence predicate. Returns true and writes result for case when RCPTTO
// command sequence is invalid, otherwise returns false
func (handler *handlerRcptto) isInvalidCmdSequence(message Message) bool {
	return (message.helo && message.mailfrom)
}

// Invalid RCPTTO command argument predicate. Returns true and writes result for case when RCPTTO
// command argument is invalid, otherwise returns false
func (handler *handlerRcptto) isInvalidCmdArg(request string) bool {
	return !matchRegex(request, validRcpttoComplexCmdRegexPattern)
}

// Returns email from RCPTTO request
func (handler *handlerRcptto) rcpttoEmail(request string) string {
	return regexCaptureGroup(request, validRcpttoComplexCmdRegexPattern, 3)
}

// Custom behavior for RCPTTO email. Returns true and writes result for case when
// RCPTTO email is included in configuration.blacklistedRcpttoEmails slice
func (handler *handlerRcptto) isBlacklistedEmail(request string) bool {
	return isIncluded(handler.blacklistedEmails, handler.rcpttoEmail(request))
}

// Custom behavior for RCPTTO email. Returns true and writes result for case when
// RCPTTO email is included in configuration.notRegisteredEmails slice
func (handler *handlerRcptto) isNotRegisteredEmail(request string) bool {
	return isIncluded(handler.notRegisteredEmails, handler.rcpttoEmail(request))
}
