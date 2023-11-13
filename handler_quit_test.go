package smtpmock

/*

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNewHandlerQuit(t *testing.T) {
	t.Run("returns new handlerQuit", func(t *testing.T) {
		handler := newHandlerQuit()
	})
}

func TestHandlerQuitRun(t *testing.T) {
	t.Run("when successful QUIT request", func(t *testing.T) {
		request, session, message, configuration := "QUIT", new(sessionMock), new(Message), createConfiguration()
		receivedMessage := configuration.msgQuitCmd
		handler := newHandlerQuit()
		session.On("writeResponse", receivedMessage, configuration.responseDelayQuit).Once().Return(nil)
		handler.run(request)

		assert.True(t, message.quitSent)
	})

	t.Run("when failure QUIT request", func(t *testing.T) {
		request, session, message, configuration := "QUIT ", new(sessionMock), new(Message), createConfiguration()
		handler := newHandlerQuit()
		handler.run(request)

		assert.False(t, message.quitSent)
	})
}

func TestHandlerQuitIsInvalidRequest(t *testing.T) {
	handler := newHandlerQuit(new(session), new(Message), new(configuration))

	t.Run("when request includes invalid QUIT command", func(t *testing.T) {
		request := "QUIT "

		assert.True(t, handler.isInvalidRequest(request))
	})

	t.Run("when request includes valid QUIT command", func(t *testing.T) {
		request := "QUIT"

		assert.False(t, handler.isInvalidRequest(request))
	})
}

*/
