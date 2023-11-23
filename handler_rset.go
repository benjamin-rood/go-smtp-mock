package smtpmock

import "errors"

// RSET command handler
type handlerRset struct {
}

// RSET command handler builder. Returns pointer to new handlerRset structure
func newHandlerRset() *handlerRset {
	return &handlerRset{}
}

// RSET handler methods

// Main RSET handler runner
func (session *session) processRSET(request string) {
	config := session.config
	session.clearError()
	session.clearMessageRSET()
	handler := newHandlerRset()

	// Check for invalid RSET command request
	if handler.isInvalidCmdSequence(*session.message) {
		session.writeRsetResult(false, request, config.msgInvalidCmdRsetSequence)
		return
	}
	if handler.isInvalidCmdArg(request) {
		session.writeRsetResult(false, request, config.msgInvalidCmdRsetArg)
		return
	}

	session.writeRsetResult(true, request, config.msgRsetReceived)
}

// Erases all message data except HELO/EHLO command context and changes cleared status to true
// for case when not multiple message receiving condition
func (session *session) clearMessageRSET() {
	messageWithData, config := session.message, session.config

	if !(config.multipleMessageReceiving && messageWithData.isConsistent()) {
		clearedMessage := &Message{
			heloRequest:  messageWithData.heloRequest,
			heloResponse: messageWithData.heloResponse,
			helo:         messageWithData.helo,
		}
		*messageWithData = *clearedMessage
	}
}

// Writes handled RSET result to session, message. Always returns true
func (session *session) writeRsetResult(isSuccessful bool, request, response string) bool {
	config, message := session.config, session.message
	if !isSuccessful {
		session.addError(errors.New(response))
	}

	message.rsetRequest, message.rsetResponse, message.rset = request, response, isSuccessful
	session.writeResponse(response, config.responseDelayRset)
	return true
}

// Invalid RSET command sequence predicate. Returns true and writes result for case when
// RSET command sequence is invalid (HELO command was failure), otherwise returns false
func (handler *handlerRset) isInvalidCmdSequence(message Message) bool {
	return !message.helo
}

// Invalid RSET command argument predicate. Returns true and writes result for case when
// RSET command argument is invalid, otherwise returns false
func (handler *handlerRset) isInvalidCmdArg(request string) bool {
	return !matchRegex(request, validRsetCmdRegexPattern)
}
