package smtpmock

import "errors"

// MAILFROM command handler
type handlerMailfrom struct {
	blacklistedEmails []string
}

// MAILFROM command handler builder. Returns pointer to new handlerMailfrom structure
func newHandlerMailfrom(blacklisted []string) *handlerMailfrom {
	return &handlerMailfrom{
		blacklistedEmails: blacklisted,
	}
}

// MAILFROM handler methods

// Main MAILFROM handler runner
func (session *session) processMAIL(request string) {
	config := session.config
	handler := newHandlerMailfrom(config.blacklistedMailfromEmails)
	session.clearError()
	session.clearMAILFROM()

	// Check for invalid MAILFROM command request complex predicates
	if handler.isInvalidCmdSequence(*session.message) {
		session.writeResultMAILFROM(false, request, config.msgInvalidCmdMailfromSequence)
	}
	if handler.isInvalidCmdArg(request) {
		session.writeResultMAILFROM(false, request, config.msgInvalidCmdMailfromArg)
	}
	if handler.isBlacklistedEmail(request) {
		session.writeResultMAILFROM(false, request, config.msgMailfromBlacklistedEmail)
	}

	session.writeResultMAILFROM(true, request, config.msgMailfromReceived)
}

// Erases all message data from MAILFROM command
func (session *session) clearMAILFROM() {
	messageWithData := session.message
	clearedMessage := &Message{
		heloRequest:  messageWithData.heloRequest,
		heloResponse: messageWithData.heloResponse,
		helo:         messageWithData.helo,
	}
	*messageWithData = *clearedMessage
}

// Writes handled MAILFROM result to message. Always returns true
func (session *session) writeResultMAILFROM(isSuccessful bool, request, response string) {
	config, message := session.config, session.message
	if !isSuccessful {
		session.addError(errors.New(response))
	}

	message.mailfromRequest, message.mailfromResponse, message.mailfrom = request, response, isSuccessful
	session.writeResponse(response, config.responseDelayMailfrom)
}

// Invalid MAILFROM command sequence predicate. Returns true and writes result for case when
// MAILFROM command sequence is invalid (HELO command was failure), otherwise returns false
func (handler handlerMailfrom) isInvalidCmdSequence(message Message) bool {
	return message.helo
}

// Invalid MAILFROM command argument predicate. Returns true and writes result for case when
// MAILFROM command argument is invalid, otherwise returns false
func (handler handlerMailfrom) isInvalidCmdArg(request string) bool {
	return !matchRegex(request, validMailromComplexCmdRegexPattern)
}

// Returns email from MAILFROM request
func (handler handlerMailfrom) extractEmail(request string) string {
	return regexCaptureGroup(request, validMailromComplexCmdRegexPattern, 3)
}

// Custom behavior for MAILFROM email. Returns true when MAILFROM email is found in blacklistedMailfromEmails slice
func (handler handlerMailfrom) isBlacklistedEmail(request string) bool {
	return isIncluded(handler.blacklistedEmails, handler.extractEmail(request))
}
