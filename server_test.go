package smtpmock

import (
	"errors"
	"fmt"
	"net"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestNewServer(t *testing.T) {
	t.Run("creates new server", func(t *testing.T) {
		configuration := createConfiguration()
		server := newServer(configuration)

		assert.Same(t, configuration, server.configuration)
		assert.Equal(t, []Message{}, server.Messages())
		assert.Equal(t, newLogger(configuration.logToStdout, configuration.logServerActivity), server.logger)
		assert.Nil(t, server.listener)
		assert.NotNil(t, server.wg)
		assert.Nil(t, server.quit)
		assert.False(t, server.isStarted)
		assert.Equal(t, 0, server.PortNumber)
	})
}

func TestNewMessageWithHeloContext(t *testing.T) {
	t.Run("returns new message with helo context from other message", func(t *testing.T) {
		message, heloRequest, heloResponse, helo := Message{}, "heloRequest", "heloResponse", true
		message.heloRequest, message.heloResponse, message.helo = heloRequest, heloResponse, helo
		newMessage := newMessageWithHeloContext(message)

		assert.Equal(t, heloRequest, newMessage.heloRequest)
		assert.Equal(t, heloResponse, newMessage.heloResponse)
		assert.Equal(t, helo, newMessage.helo)
	})
}

func TestServerIsInvalidCmd(t *testing.T) {
	availableComands := strings.Split("helo,ehlo,mail from:,rcpt to:,data,quit", ",")

	for _, validCommand := range availableComands {
		t.Run("when valid command", func(t *testing.T) {
			assert.False(t, isInvalidCmd(validCommand))
		})
	}

	t.Run("when invalid command", func(t *testing.T) {
		assert.True(t, isInvalidCmd("some invalid command"))
	})
}

func TestServerRecognizeCommand(t *testing.T) {
	t.Run("captures the first word divided by spaces, converts it to upper case", func(t *testing.T) {
		firstWord, secondWord := "first", " command"
		command := firstWord + secondWord

		assert.Equal(t, strings.ToUpper(firstWord), recognizeCommand(command))
	})
}

func TestServerAddToWaitGroup(t *testing.T) {
	waitGroup := new(waitGroupMock)
	server := &Server{wg: waitGroup}

	t.Run("increases count of goroutines by one", func(t *testing.T) {
		waitGroup.On("Add", 1).Once().Return(nil)
		server.addToWaitGroup()
	})
}

func TestServerRemoveFromWaitGroup(t *testing.T) {
	waitGroup := new(waitGroupMock)
	server := &Server{wg: waitGroup}

	t.Run("decreases count of goroutines by one", func(t *testing.T) {
		waitGroup.On("Done").Once().Return(nil)
		server.removeFromWaitGroup()
	})
}

func TestServerIsAbleToEndSession(t *testing.T) {
	t.Run("when quit command has been sent", func(t *testing.T) {
		server, message, session := newServer(createConfiguration()), Message{quitSent: true}, new(session)
		server.Start()
		server.messages.Append(message)

		assert.True(t, server.isAbleToEndSession(message, session))
	})

	t.Run("when quit command has not been sent, error has been found, fail fast scenario has been enabled", func(t *testing.T) {
		server, message, session := newServer(createConfiguration()), Message{}, new(session)
		server.Start()
		server.messages.Append(message)
		session.err = errors.New("some error")
		server.configuration.isCmdFailFast = true

		assert.True(t, server.isAbleToEndSession(message, session))
	})

	t.Run("when quit command has not been sent, no errors", func(t *testing.T) {
		server, message, session := newServer(createConfiguration()), Message{}, new(session)
		server.Start()
		server.messages.Append(message)

		assert.False(t, server.isAbleToEndSession(message, session))
	})

	t.Run("when quit command has not been sent, error has been found, fail fast scenario has not been enabled", func(t *testing.T) {
		server, message, session := newServer(createConfiguration()), Message{}, new(session)
		server.Start()
		server.messages.Append(message)
		session.err = errors.New("some error")

		assert.False(t, server.isAbleToEndSession(message, session))
	})
}

func TestServerHandleSession(t *testing.T) {
	t.Run("when complex successful session, multiple message receiving scenario disabled", func(t *testing.T) {
		session, configuration := &sessionMock{}, createConfiguration()
		server := newServer(configuration)
		server.Start()

		session.On("processResponse").Once().Return(Message{})

		session.On("writeResponse", configuration.msgGreeting, defaultSessionResponseDelay).Once().Return(nil)

		session.On("setTimeout", defaultSessionTimeout).Once().Return(nil)
		session.On("readRequest").Once().Return("helo example.com", nil)
		session.On("clearError").Once().Return(nil)
		session.On("writeResponse", configuration.msgHeloReceived, configuration.responseDelayHelo).Once().Return(nil)
		session.On("isErrorFound").Once().Return(false)

		session.On("setTimeout", defaultSessionTimeout).Once().Return(nil)
		session.On("readRequest").Once().Return("ehlo example.com", nil)
		session.On("clearError").Once().Return(nil)
		session.On("writeResponse", configuration.msgHeloReceived, configuration.responseDelayHelo).Once().Return(nil)
		session.On("isErrorFound").Once().Return(false)

		session.On("setTimeout", defaultSessionTimeout).Once().Return(nil)
		session.On("readRequest").Once().Return("rset", nil)
		session.On("clearError").Once().Return(nil)
		session.On("writeResponse", configuration.msgRsetReceived, configuration.responseDelayRset).Once().Return(nil)
		session.On("isErrorFound").Once().Return(false)

		session.On("setTimeout", defaultSessionTimeout).Once().Return(nil)
		session.On("readRequest").Once().Return("mail from: receiver@example.com", nil)
		session.On("clearError").Once().Return(nil)
		session.On("writeResponse", configuration.msgMailfromReceived, configuration.responseDelayMailfrom).Once().Return(nil)
		session.On("isErrorFound").Once().Return(false)

		session.On("setTimeout", defaultSessionTimeout).Once().Return(nil)
		session.On("readRequest").Once().Return("rcpt to: sender@example.com", nil)
		session.On("clearError").Once().Return(nil)
		session.On("writeResponse", configuration.msgRcpttoReceived, configuration.responseDelayRcptto).Once().Return(nil)
		session.On("isErrorFound").Once().Return(false)

		session.On("setTimeout", defaultSessionTimeout).Once().Return(nil)
		session.On("readRequest").Once().Return("data", nil)
		session.On("clearError").Once().Return(nil)
		session.On("writeResponse", configuration.msgDataReceived, configuration.responseDelayData).Once().Return(nil)
		session.On("isErrorFound").Once().Return(false)

		session.On("readBytes").Once().Return([]uint8(".some message"), nil)
		session.On("readBytes").Once().Return([]uint8(".\r\n"), nil)
		session.On("writeResponse", configuration.msgMsgReceived, configuration.responseDelayMessage).Once().Return(nil)

		session.On("setTimeout", defaultSessionTimeout).Once().Return(nil)
		session.On("readRequest").Once().Return("quit", nil)
		session.On("writeResponse", configuration.msgQuitCmd, configuration.responseDelayQuit).Once().Return(nil)
		session.On("isErrorFound").Once().Return(false)

		session.On("finish").Once().Return(nil)

		server.handleSession(session)
		assert.Equal(t, 1, len(server.Messages()))
	})

	const (
		heloExample     = "helo example.com"
		helo42          = "helo 42"
		ehloExample     = "ehlo example.com"
		mailfromExample = "mail from: receiver@example.com"
		rcpttoExample   = "rcpt to: sender1@example.com"
		data            = "data"
		rset            = "rset"
		quit            = "quit"

		notImplementedCmd = "not implemented cmd"
	)

	t.Run("when complex successful session, multiple message receiving scenario enabled", func(t *testing.T) {
		session, configuration := &sessionMock{}, createConfiguration()
		configuration.multipleMessageReceiving = true
		server := newServer(configuration)
		server.Start()

		session.On("writeResponse", configuration.msgGreeting, defaultSessionResponseDelay).Once().Return(nil)

		session.On("setTimeout", defaultSessionTimeout).Once().Return(nil)
		session.On("readRequest").Once().Return(heloExample, nil)
		session.On("processResponse", heloExample).Once().Return(Message{})
		session.On("clearError").Once().Return(nil)
		session.On("writeResponse", configuration.msgHeloReceived, configuration.responseDelayHelo).Once().Return(nil)
		session.On("isErrorFound").Once().Return(false)

		session.On("setTimeout", defaultSessionTimeout).Once().Return(nil)
		session.On("readRequest").Once().Return(ehloExample, nil)
		session.On("processResponse", ehloExample).Once().Return(Message{})
		session.On("clearError").Once().Return(nil)
		session.On("writeResponse", configuration.msgHeloReceived, configuration.responseDelayHelo).Once().Return(nil)
		session.On("isErrorFound").Once().Return(false)

		session.On("setTimeout", defaultSessionTimeout).Once().Return(nil)
		session.On("readRequest").Once().Return(mailfromExample, nil)
		session.On("processResponse", mailfromExample).Once().Return(Message{})
		session.On("clearError").Once().Return(nil)
		session.On("writeResponse", configuration.msgMailfromReceived, configuration.responseDelayMailfrom).Once().Return(nil)
		session.On("isErrorFound").Once().Return(false)

		session.On("processResponse").Once().Return(Message{}, nil)

		session.On("setTimeout", defaultSessionTimeout).Once().Return(nil)
		session.On("readRequest").Once().Return(rcpttoExample, nil)
		session.On("processResponse", rcpttoExample).Once().Return(Message{})
		session.On("clearError").Once().Return(nil)
		session.On("writeResponse", configuration.msgRcpttoReceived, configuration.responseDelayRcptto).Once().Return(nil)
		session.On("isErrorFound").Once().Return(false)

		session.On("processResponse").Once().Return(Message{}, nil)

		session.On("setTimeout", defaultSessionTimeout).Once().Return(nil)
		session.On("readRequest").Once().Return(data, nil)
		session.On("processResponse", data).Once().Return(Message{})
		session.On("clearError").Once().Return(nil)
		session.On("writeResponse", configuration.msgDataReceived, configuration.responseDelayData).Once().Return(nil)
		session.On("isErrorFound").Once().Return(false)

		session.On("readBytes").Once().Return([]uint8(".some message"), nil)
		session.On("readBytes").Once().Return([]uint8(".\r\n"), nil)
		session.On("writeResponse", configuration.msgMsgReceived, configuration.responseDelayMessage).Once().Return(nil)
		session.On("isErrorFound").Once().Return(false)

		session.On("setTimeout", defaultSessionTimeout).Once().Return(nil)
		session.On("readRequest").Once().Return(rset, nil)
		session.On("processResponse", rset).Once().Return(Message{})
		session.On("clearError").Once().Return(nil)
		session.On("writeResponse", configuration.msgRsetReceived, configuration.responseDelayRset).Once().Return(nil)
		session.On("isErrorFound").Once().Return(false)

		session.On("setTimeout", defaultSessionTimeout).Once().Return(nil)
		session.On("readRequest").Once().Return(mailfromExample, nil)
		session.On("processResponse", mailfromExample).Once().Return(Message{})
		session.On("clearError").Once().Return(nil)
		session.On("writeResponse", configuration.msgMailfromReceived, configuration.responseDelayMailfrom).Once().Return(nil)
		session.On("isErrorFound").Once().Return(false)

		session.On("setTimeout", defaultSessionTimeout).Once().Return(nil)
		session.On("readRequest").Once().Return(rcpttoExample, nil)
		session.On("processResponse", rcpttoExample).Once().Return(Message{})
		session.On("clearError").Once().Return(nil)
		session.On("writeResponse", configuration.msgRcpttoReceived, configuration.responseDelayRcptto).Once().Return(nil)
		session.On("isErrorFound").Once().Return(false)

		session.On("setTimeout", defaultSessionTimeout).Once().Return(nil)
		session.On("readRequest").Once().Return(data, nil)
		session.On("processResponse", data).Once().Return(Message{})
		session.On("clearError").Once().Return(nil)
		session.On("writeResponse", configuration.msgDataReceived, configuration.responseDelayData).Once().Return(nil)
		session.On("isErrorFound").Once().Return(false)

		session.On("readBytes").Once().Return([]uint8(".some message"), nil)
		session.On("readBytes").Once().Return([]uint8(".\r\n"), nil)
		session.On("writeResponse", configuration.msgMsgReceived, configuration.responseDelayMessage).Once().Return(nil)
		session.On("isErrorFound").Once().Return(false)

		session.On("setTimeout", defaultSessionTimeout).Once().Return(nil)
		session.On("readRequest").Once().Return(quit, nil)
		session.On("processResponse", quit).Once().Return(Message{})
		session.On("writeResponse", configuration.msgQuitCmd, configuration.responseDelayQuit).Once().Return(nil)
		session.On("isErrorFound").Once().Return(false)

		session.On("finish").Once().Return(nil)

		server.handleSession(session)
		assert.Equal(t, 1, len(server.Messages()))
	})

	t.Run("when invalid command, fail fast scenario disabled", func(t *testing.T) {
		session, configuration := &sessionMock{}, createConfiguration()
		server := newServer(configuration)
		server.Start()

		session.On("writeResponse", configuration.msgGreeting, defaultSessionResponseDelay).Once().Return(nil)

		session.On("setTimeout", defaultSessionTimeout).Once().Return(nil)
		session.On("readRequest").Once().Return("not implemented command", nil)
		session.On("writeResponse", configuration.msgInvalidCmd, defaultSessionResponseDelay).Once().Return(nil)

		session.On("setTimeout", defaultSessionTimeout).Once().Return(nil)
		session.On("readRequest").Once().Return("quit", nil)
		session.On("processResponse", "quit").Once().Return(Message{})
		session.On("writeResponse", configuration.msgQuitCmd, configuration.responseDelayQuit).Once().Return(nil)

		session.On("isErrorFound").Once().Return(true)
		session.On("finish").Once().Return(nil)

		server.handleSession(session)
	})

	t.Run("when invalid command, session error, fail fast scenario enabled", func(t *testing.T) {
		session, configuration := &sessionMock{}, newConfiguration(ConfigurationAttr{IsCmdFailFast: true})
		server, errorMessage := newServer(configuration), configuration.msgInvalidCmdHeloArg
		server.Start()

		session.On("writeResponse", configuration.msgGreeting, defaultSessionResponseDelay).Once().Return(nil)

		session.On("setTimeout", defaultSessionTimeout).Once().Return(nil)
		session.On("readRequest").Once().Return(notImplementedCmd, nil)
		session.On("writeResponse", configuration.msgInvalidCmd, defaultSessionResponseDelay).Once().Return(nil)

		session.On("setTimeout", defaultSessionTimeout).Once().Return(nil)
		session.On("readRequest").Once().Return(helo42, nil)
		session.On("processResponse", helo42).Once().Return(Message{})
		session.On("clearError").Once().Return(nil)
		session.On("addError", errors.New(errorMessage)).Once().Return(nil)
		session.On("writeResponse", errorMessage, defaultSessionResponseDelay).Once().Return(nil)

		session.On("isErrorFound").Once().Return(true)
		session.On("finish").Once().Return(nil)

		server.handleSession(session)
	})

	t.Run("when server quit channel was closed", func(t *testing.T) {
		session, configuration := &sessionMock{}, newConfiguration(ConfigurationAttr{IsCmdFailFast: true})
		server := newServer(configuration)
		server.quit = make(chan interface{})
		close(server.quit)

		session.On("writeResponse", configuration.msgGreeting, defaultSessionResponseDelay).Once().Return(nil)
		session.On("finish").Once().Return(nil)

		server.handleSession(session)
	})

	t.Run("when read request session error", func(t *testing.T) {
		session, configuration := &sessionMock{}, newConfiguration(ConfigurationAttr{IsCmdFailFast: true})
		server := newServer(configuration)

		session.On("writeResponse", configuration.msgGreeting, defaultSessionResponseDelay).Once().Return(nil)
		session.On("setTimeout", defaultSessionTimeout).Once().Return(nil)
		session.On("readRequest").Once().Return(emptyString, errors.New("some read request error"))
		session.On("finish").Once().Return(nil)

		server.handleSession(session)
	})
}

func TestServerStart(t *testing.T) {
	t.Run("when no errors happens during starting and running the server with default port", func(t *testing.T) {
		configuration := createConfiguration()
		server := newServer(configuration)

		assert.NoError(t, server.Start())
		err := runSuccessfulSMTPSession(configuration.hostAddress, server.PortNumber, false)
		assert.NoError(t, err)
		time.Sleep(5 * time.Millisecond)
		assert.NotEmpty(t, server.Messages())
		assert.NotNil(t, server.quit)
		assert.NotNil(t, server.quitTimeout)
		assert.True(t, server.isStarted)
		assert.Greater(t, server.PortNumber, 0)

		assert.NoError(t, server.Stop())
	})

	t.Run("when no errors happens during starting and running the server with custom port", func(t *testing.T) {
		configuration, portNumber := createConfiguration(), 2525
		configuration.portNumber = portNumber
		server := newServer(configuration)

		assert.NoError(t, server.Start())
		err := runSuccessfulSMTPSession(configuration.hostAddress, portNumber, false)
		assert.NoError(t, err)
		time.Sleep(5 * time.Millisecond)
		assert.NotEmpty(t, server.Messages())
		assert.NotNil(t, server.quit)
		assert.NotNil(t, server.quitTimeout)
		assert.True(t, server.isStarted)
		assert.Equal(t, portNumber, server.PortNumber)

		assert.NoError(t, server.Stop())
	})

	t.Run("when active server doesn't start current server", func(t *testing.T) {
		server := &Server{isStarted: true}

		assert.EqualError(t, server.Start(), serverStartErrorMsg)
		assert.Equal(t, 0, server.PortNumber)
	})

	t.Run("when listener error happens during starting the server doesn't start current server", func(t *testing.T) {
		configuration := createConfiguration()
		server, logger := newServer(configuration), new(loggerMock)
		listener, _ := net.Listen(networkProtocol, emptyString)
		portNumber := listener.Addr().(*net.TCPAddr).Port
		errorMessage := fmt.Sprintf("%s: %d", serverErrorMsg, portNumber)
		configuration.portNumber, server.logger = portNumber, logger
		logger.On("error", errorMessage).Once().Return(nil)

		assert.EqualError(t, server.Start(), errorMessage)
		assert.False(t, server.isStarted)
		assert.Equal(t, 0, server.PortNumber)
		listener.Close()
	})
}

func TestServerStop(t *testing.T) {
	t.Run("when server active stops current server, graceful shutdown case", func(t *testing.T) {
		logger, listener, waitGroup, quitChannel := new(loggerMock), new(listenerMock), new(waitGroupMock), make(chan interface{})
		server := &Server{
			configuration: createConfiguration(),
			logger:        logger,
			listener:      listener,
			wg:            waitGroup,
			quit:          quitChannel,
			isStarted:     true,
			quitTimeout:   make(chan interface{}),
			messages:      NewMessageList(),
		}
		listener.On("Close").Once().Return(nil)
		waitGroup.On("Wait").Once().Return(nil)
		logger.On("infoActivity", serverStopMsg).Once().Return(nil)

		assert.NoError(t, server.Stop())
		assert.False(t, server.isStarted)
		_, isChannelOpened := <-server.quit
		assert.False(t, isChannelOpened)
	})

	t.Run("when server active stops current server, force shutdown case", func(t *testing.T) {
		logger, listener, waitGroup, quitChannel := new(loggerMock), new(listenerMock), new(waitGroupMock), make(chan interface{})
		server := &Server{
			configuration: createConfiguration(),
			logger:        logger,
			listener:      listener,
			wg:            waitGroup,
			quit:          quitChannel,
			isStarted:     true,
			messages:      NewMessageList(),
		}
		listener.On("Close").Once().Return(nil)
		waitGroup.On("Wait").Once().Return(nil)
		logger.On("infoActivity", serverForceStopMsg).Once().Return(nil)

		assert.NoError(t, server.Stop())
		assert.False(t, server.isStarted)
		_, isChannelOpened := <-server.quit
		assert.False(t, isChannelOpened)
	})

	t.Run("when server is inactive doesn't stop current server", func(t *testing.T) {
		assert.EqualError(t, new(Server).Stop(), serverStopErrorMsg)
	})
}

func TestServerMessages(t *testing.T) {
	configuration := createConfiguration()

	t.Run("when there are no messages on the server", func(t *testing.T) {
		server := newServer(configuration)

		assert.Empty(t, server.Messages())
	})

	t.Run("when there are messages on the server", func(t *testing.T) {
		server := newServer(configuration)

		assert.Empty(t, server.Messages())

		message := Message{}
		go server.messages.Writer()
		server.messages.Append(message)
		server.messages.Stop() // Append will now exit
		time.Sleep(50 * time.Millisecond)

		assert.NotEmpty(t, server.Messages())
	})

	t.Run("message data are identical", func(t *testing.T) {
		server := newServer(configuration)

		assert.Empty(t, server.messages.Messages())
		assert.Empty(t, server.Messages())
		assert.NotSame(t, server.messages.Messages(), server.Messages())

		message := Message{}
		go server.messages.Writer()
		server.messages.Append(message)
		server.messages.Stop() // Append will now exit
		time.Sleep(50 * time.Millisecond)
		assert.Equal(t, []Message{message}, server.messages.Messages())
		assert.Equal(t, []Message{message}, server.Messages())
		assert.Equal(t, server.messages.Messages(), server.Messages())
		assert.NotSame(t, server.messages.Messages(), server.Messages())
	})
}
