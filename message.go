package smtpmock

import (
	"sync/atomic"
)

// Structure for storing the result of SMTP client-server interaction. Context-included
// commands should be represented as request/response structure fields
type Message struct {
	heloRequest, heloResponse                               string
	mailfromRequest, mailfromResponse                       string
	rcpttoRequestResponse                                   [][]string
	dataRequest, dataResponse                               string
	msgRequest, msgResponse                                 string
	rsetRequest, rsetResponse                               string
	helo, mailfrom, rcptto, data, msg, rset, noop, quitSent bool
}

type MessageNode struct {
	data Message
	next *MessageNode
}

// message getters

// Getter for heloRequest field
func (message Message) HeloRequest() string {
	return message.heloRequest
}

// Getter for heloResponse field
func (message Message) HeloResponse() string {
	return message.heloResponse
}

// Getter for helo field
func (message Message) Helo() bool {
	return message.helo
}

// Getter for mailfromRequest field
func (message Message) MailfromRequest() string {
	return message.mailfromRequest
}

// Getter for mailfromResponse field
func (message Message) MailfromResponse() string {
	return message.mailfromResponse
}

// Getter for mailfrom field
func (message Message) Mailfrom() bool {
	return message.mailfrom
}

// Getter for rcpttoRequestResponse field
func (message Message) RcpttoRequestResponse() [][]string {
	return message.rcpttoRequestResponse
}

// Getter for rcptto field
func (message Message) Rcptto() bool {
	return message.rcptto
}

// Getter for dataRequest field
func (message Message) DataRequest() string {
	return message.dataRequest
}

// Getter for dataResponse field
func (message Message) DataResponse() string {
	return message.dataResponse
}

// Getter for data field
func (message Message) Data() bool {
	return message.data
}

// Getter for msgRequest field
func (message Message) MsgRequest() string {
	return message.msgRequest
}

// Getter for msgResponse field
func (message Message) MsgResponse() string {
	return message.msgResponse
}

// Getter for msg field
func (message Message) Msg() bool {
	return message.msg
}

// Getter for rsetRequest field
func (message Message) RsetRequest() string {
	return message.rsetRequest
}

// Getter for rsetResponse field
func (message Message) RsetResponse() string {
	return message.rsetResponse
}

// Getter for rset field
func (message Message) Rset() bool {
	return message.rset
}

// Getter for noop field
func (message Message) Noop() bool {
	return message.noop
}

// Getter for quitSent field
func (message Message) QuitSent() bool {
	return message.quitSent
}

// Getter for message consistency status predicate. Returns true
// for case when message struct is consistent. It means that
// MAILFROM, RCPTTO, DATA commands and message context
// were successful. Otherwise returns false
func (message Message) IsConsistent() bool {
	return message.mailfrom && message.rcptto && message.data && message.msg
}

// Message pointer consistency status predicate. Returns true for case
// when message struct is consistent. It means that MAILFROM, RCPTTO, DATA
// commands and message context were successful. Otherwise returns false
func (message *Message) isConsistent() bool {
	return message.mailfrom && message.rcptto && message.data && message.msg
}

// Message RCPTTO successful response predicate. Returns true when at least one
// successful RCPTTO response exists. Otherwise returns false
func (message *Message) isIncludesSuccessfulRcpttoResponse(targetSuccessfulResponse string) bool {
	for _, slice := range message.rcpttoRequestResponse {
		if slice[1] == targetSuccessfulResponse {
			return true
		}
	}

	return false
}

// Pointer to empty message
var zeroMessage = &Message{}

// MessageList is an append-only concurrent-safe linked-list when using the provided methods.
// Oldest element is stored at the head, newest element stored at the tail.
// There is no need to manually read or set the `head` or `tail`, use the provided
// methods ONLY.
type MessageList struct {
	head    *MessageNode
	tail    atomic.Pointer[MessageNode] // points to a MessageNode
	q       chan Message
	stopped atomic.Bool
}

func NewMessageList() *MessageList {
	return &MessageList{
		q: make(chan (Message)),
	}
}

func (list *MessageList) Append(m Message) {
	list.q <- m
}

func (list *MessageList) Stop() {
	close(list.q)
	// atomically mark the MessageList as stopped
	list.stopped.Store(true)
}

// Writer should be run in a separate goroutine.
func (list *MessageList) Writer() {
	if list.q == nil {
		panic("MessageList.Writer: uninitialised MessageList")
	}
	if list.stopped.Load() == true {
		panic("MessageList.Writer: stopped")
	}
	for {
		select {
		case data, open := <-list.q:
			if open {
				newNode := &MessageNode{data: data, next: nil}
				if list.head == nil {
					// when the list is initially empty, set head to the new node.
					list.head = newNode
				} else {
					// otherwise we need to update the tail
					currentTail := list.tail.Load()
					currentTail.next = newNode
				}
				// atomically update the tail pointer
				list.tail.Store(newNode)
			} else {
				return // channel closed, exit
			}
		}
	}
}

func (list *MessageList) Messages() []Message {
	messages := []Message{}
	ptr := list.head
	for ptr != nil {
		msg := ptr.data
		messages = append(messages, msg)
		ptr = ptr.next
	}
	return messages
}
