package smtpmock

import (
	"bytes"
	"errors"
)

// Main message handler runner
func (session *session) runMessageHandler() {
	var request string
	var msgData []byte
	config := session.config

	for {
		line, err := session.readBytes()
		if err != nil {
			return
		}

		// Handles end of data denoted by lone period (\r\n.\r\n)
		if bytes.Equal(line, []byte(".\r\n")) {
			break
		}

		// Removes leading period (RFC 5321 section 4.5.2)
		if line[0] == '.' {
			line = line[1:]
		}

		// Enforces the maximum message size limit
		if len(msgData)+len(line) > config.msgSizeLimit {
			session.discardBufin()
			session.writeMessageResult(false, request, config.msgMsgSizeIsTooBig)
			return
		}

		msgData = append(msgData, line...)
	}

	session.writeMessageResult(true, string(msgData), config.msgMsgReceived)
}

// Writes handled message result to session, message
func (session *session) writeMessageResult(isSuccessful bool, request, response string) {
	config, message := session.config, session.message
	if !isSuccessful {
		session.addError(errors.New(response))
	}

	message.msgRequest, message.msgResponse, message.msg = request, response, isSuccessful
	session.writeResponse(response, config.responseDelayMessage)
}
