package smtpmock

import "errors"

// DATA command handler
type handlerData struct {
}

// DATA command handler builder. Returns pointer to new handlerData structure
func newHandlerData() *handlerData {
	return &handlerData{}
}

// DATA handler methods

// Main DATA handler runner
func (session *session) processDATA(request string) {
	config := session.config
	session.clearError()
	session.clearMessageDATA()
	handler := newHandlerData()

	// Check for invalid DATA command predicate
	if handler.isInvalidCmdSequence(*session.message) {
		session.writeDataResult(false, request, config.msgInvalidCmdDataSequence)
		return
	}
	if handler.isInvalidCmd(request) {
		session.writeDataResult(false, request, config.msgInvalidCmd)
		return
	}

	session.writeDataResult(true, request, config.msgDataReceived)
	session.processIncomingMessage()
}

// Erases all message data from DATA command
func (session *session) clearMessageDATA() {
	messageWithData := session.message
	clearedMessage := &Message{
		heloRequest:      messageWithData.heloRequest,
		heloResponse:     messageWithData.heloResponse,
		helo:             messageWithData.helo,
		mailfromRequest:  messageWithData.mailfromRequest,
		mailfromResponse: messageWithData.mailfromResponse,
		mailfrom:         messageWithData.mailfrom,
		rcpttoRequest:    messageWithData.rcpttoRequest,
		rcpttoResponse:   messageWithData.rcpttoResponse,
		rcptto:           messageWithData.rcptto,
	}
	session.message = clearedMessage
}

// Reads and saves message body context using handlerMessage under the hood
func (session *session) processIncomingMessage() {
	session.runMessageHandler()
}

// Writes handled DATA result to session, message
func (session *session) writeDataResult(isSuccessful bool, request, response string) {
	config, message := session.config, session.message
	if !isSuccessful {
		session.addError(errors.New(response))
	}

	message.dataRequest, message.dataResponse, message.data = request, response, isSuccessful
	session.writeResponse(response, config.responseDelayData)
}

// Invalid DATA command sequence predicate. Returns true and writes result for case
// when DATA command sequence is invalid, otherwise returns false
func (handler *handlerData) isInvalidCmdSequence(message Message) bool {
	return (message.helo && message.mailfrom && message.rcptto)
}

// Invalid DATA command predicate. Returns true and writes result for case
// when DATA command is invalid, otherwise returns false
func (handler *handlerData) isInvalidCmd(request string) bool {
	return !matchRegex(request, validDataCmdRegexPattern)
}
