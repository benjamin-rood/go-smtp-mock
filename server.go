package smtpmock

import (
	"errors"
	"fmt"
	"net"
	"strings"
	"sync"
	"time"
)

// WaitGroup interface
type waitGroup interface {
	Add(int)
	Done()
	Wait()
}

// Server structure which implements SMTP mock server
type Server struct {
	configuration *configuration
	messages      *MessageList
	logger        logger
	listener      net.Listener
	wg            waitGroup
	quit          chan interface{}
	started       bool
	portNumber    int
	quitTimeout   chan interface{}
	sync.Mutex
}

// SMTP mock server builder, creates new server
func newServer(configuration *configuration) *Server {
	return &Server{
		configuration: configuration,
		messages:      NewMessageList(),
		logger:        newLogger(configuration.logToStdout, configuration.logServerActivity),
		wg:            new(sync.WaitGroup),
	}
}

// server methods

// Start binds and runs SMTP mock server on specified port or random free port. Returns error for
// case when server is active. Server port number will be assigned after successful start only
func (server *Server) Start() (err error) {
	if server.isStarted() {
		return errors.New(serverStartErrorMsg)
	}

	configuration, logger := server.configuration, server.logger
	portNumber := configuration.portNumber

	listener, err := net.Listen(networkProtocol, serverWithPortNumber(configuration.hostAddress, portNumber))
	if err != nil {
		errorMessage := fmt.Sprintf("%s: %d", serverErrorMsg, portNumber)
		logger.error(errorMessage)
		return errors.New(errorMessage)
	}

	portNumber = listener.Addr().(*net.TCPAddr).Port
	server.setListener(listener)
	server.setPortNumber(portNumber)
	server.start()
	server.quit, server.quitTimeout = make(chan interface{}), make(chan interface{})
	logger.infoActivity(fmt.Sprintf("%s: %d", serverStartMsg, portNumber))

	go server.messages.Writer()
	server.addToWaitGroup()
	go func() {
		defer server.removeFromWaitGroup()
		for {
			connection, err := server.listener.Accept()
			if err != nil {
				if _, ok := <-server.quit; !ok {
					logger.warning(serverNotAcceptNewConnectionsMsg)
				}
				return
			}

			server.addToWaitGroup()
			go func() {
				server.handleSession(newSession(server.configuration, connection, logger))
				server.removeFromWaitGroup()
			}()

			logger.infoActivity(sessionStartMsg)
		}
	}()

	return err
}

// Stop shutdowns server gracefully or force by timeout.
// Returns error for case when server is not active
func (server *Server) Stop() (err error) {
	if server.isStarted() {
		server.messages.Stop()
		close(server.quit)
		server.listener.Close()

		go func() {
			server.wg.Wait()
			server.quitTimeout <- true
			server.stop()
			server.logger.infoActivity(serverStopMsg)
		}()

		select {
		case <-server.quitTimeout:
		case <-time.After(time.Duration(server.configuration.shutdownTimeout) * time.Second):
			server.stop()
			server.logger.infoActivity(serverForceStopMsg)
		}

		return
	}

	return errors.New(serverStopErrorMsg)
}

// Public interface to get access to server messages.
// Returns slice with copy of messages
func (server *Server) Messages() []Message {
	return server.messages.Messages()
}

// Thread-safe getter of server port.
// Returns server.portNumber
func (server *Server) PortNumber() int {
	server.Lock()
	defer server.Unlock()
	return server.portNumber
}

// Thread-safe getter to check if server has been started.
// Returns server.started
func (server *Server) isStarted() bool {
	server.Lock()
	defer server.Unlock()
	return server.started
}

// Thread-safe setter of server.listener
func (server *Server) setListener(listener net.Listener) {
	server.Lock()
	defer server.Unlock()
	server.listener = listener
}

// Thread-safe setter of server.portNumber
func (server *Server) setPortNumber(port int) {
	server.Lock()
	defer server.Unlock()
	server.portNumber = port
}

// Thread-safe setter of started-flag to indicate server has been started
func (server *Server) start() {
	server.Lock()
	defer server.Unlock()
	server.started = true
}

// Thread-safe setter of started-flag to indicate server has been stopped
func (server *Server) stop() {
	server.Lock()
	defer server.Unlock()
	server.started = false
}

// Creates and assigns new message with helo context from other message to server.messages
func newMessageWithHeloContext(otherMessage Message) *Message {
	return &Message{
		heloRequest:  otherMessage.heloRequest,
		heloResponse: otherMessage.heloResponse,
		helo:         otherMessage.helo,
	}
}

// Invalid SMTP command predicate. Returns true when command is invalid, otherwise returns false
func isInvalidCmd(request string) bool {
	return !matchRegex(request, availableCmdsRegexPattern)
}

// Recognizes command implemented commands. Captures the first word divided by spaces,
// converts it to upper case
func recognizeCommand(request string) string {
	command := strings.Split(request, " ")[0]
	return strings.ToUpper(command)
}

// Addes goroutine to WaitGroup
func (server *Server) addToWaitGroup() {
	server.wg.Add(1)
}

// Removes goroutine from WaitGroup
func (server *Server) removeFromWaitGroup() {
	server.wg.Done()
}

// Checks ability to end current session
func (server *Server) isAbleToEndSession(message Message, session *session) bool {
	return message.quitSent || (session.isErrorFound() && server.configuration.isCmdFailFast)
}

//nolint:gocyclo // SMTP client-server session handler
func (server *Server) handleSession(session *session) {
	defer session.finish()
	// message, configuration := server.newMessage(), server.configuration

	for {
		select {
		case <-server.quit:
			return
		default:
			sentMsg, err := session.ProcessRequest()
			if err != nil {
				return
			}
			server.messages.Append(sentMsg)

			if server.isAbleToEndSession(sentMsg, session) {
				return
			}
		}
	}
}
