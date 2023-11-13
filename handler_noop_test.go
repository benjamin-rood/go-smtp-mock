package smtpmock

import (
	"bufio"
	"bytes"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNewHandlerNoop(t *testing.T) {
	t.Skip("returns new handleNoop")
}

func TestHandlerNoopRun(t *testing.T) {
	t.Run("when successful NOOP request", func(t *testing.T) {
		request := "NOOP"
		configuration := createConfiguration()
		connectionAddress := "127.0.0.1:25"
		connection, address, logger := netConnectionMock{}, netAddressMock{}, new(loggerMock)
		address.On("String").Once().Return(connectionAddress)
		connection.On("RemoteAddr").Once().Return(address)
		response := "250 Ok"
		logger.On("infoActivity", sessionResponseMsg+response).Once().Return(nil)
		session := newSession(configuration, connection, logger)
		binaryData := bytes.NewBufferString("")
		session.bufout = bufio.NewWriter(binaryData)
		session.processNOOP(request)

		assert.True(t, session.message.noop)
		assert.Equal(t, response+"\r\n", binaryData.String())
		assert.NoError(t, session.err)
	})

	t.Run("when failure NOOP request", func(t *testing.T) {
		request := "NOOP "
		configuration := createConfiguration()
		connectionAddress := "127.0.0.1:25"
		connection, address, logger := netConnectionMock{}, netAddressMock{}, new(loggerMock)
		address.On("String").Once().Return(connectionAddress)
		connection.On("RemoteAddr").Once().Return(address)
		session := newSession(configuration, connection, logger)
		session.processNOOP(request)

		assert.False(t, session.message.noop)
	})
}

func TestHandlerNoopIsInvalidRequest(t *testing.T) {
	handler := newHandlerNoop()

	t.Run("when request includes invalid NOOP command", func(t *testing.T) {
		request := "NOOP "

		assert.True(t, handler.isInvalidRequest(request))
	})

	t.Run("when request includes valid NOOP command", func(t *testing.T) {
		request := "NOOP"

		assert.False(t, handler.isInvalidRequest(request))
	})
}
