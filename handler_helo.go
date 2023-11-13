package smtpmock

import "errors"

// Main HELO handler runner
func (session *session) processHELO(request string) {
	config := session.config
	session.clearError()
	session.clearMessage()

	if isInvalidHeloCmdArg(request) {
		session.writeHeloResult(false, request, session.config.msgInvalidCmdHeloArg)
		return
	}
	if isIncluded(config.blacklistedHeloDomains, heloDomain(request)) {
		session.writeHeloResult(false, request, config.msgHeloBlacklistedDomain)
		return
	}

	session.writeHeloResult(true, request, config.msgHeloReceived)
}

// Erases all message data
func (session *session) clearMessage() {
	session.message = new(Message)
}

// Writes handled HELO result to session, message
func (session *session) writeHeloResult(isSuccessful bool, request, response string) {
	config, message := session.config, session.message
	if !isSuccessful {
		session.addError(errors.New(response))
	}

	message.heloRequest, message.heloResponse, message.helo = request, response, isSuccessful
	session.writeResponse(response, config.responseDelayHelo)
}

// Checks for invalid HELO command argument predicate. Returns true and writes result for case when HELO command
// argument is invalid, otherwise returns false
func isInvalidHeloCmdArg(request string) bool {
	return !matchRegex(request, validHeloComplexCmdRegexPattern)
}

// Returns domain from HELO request
func heloDomain(request string) string {
	return regexCaptureGroup(request, validHeloComplexCmdRegexPattern, 2)
}
