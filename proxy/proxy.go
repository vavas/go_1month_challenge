package proxy

import (
	"encoding/json"
	"github.com/gin-gonic/gin"
	"github.com/vavas/go_1month_challenge/config"
	"github.com/vavas/go_services/services/extsrv"

	"log"
	"net/http"
	"regexp"
)

type Proxy struct {
	config   *config.Config
	handlers []*Handler
}

func New(config *config.Config) (*Proxy, error) {

	proxy := &Proxy{
		config:   config,
		handlers: make([]*Handler, 0),
	}
	for _, handlerConfig := range config.Handlers {
		proxy.bindHandler(handlerConfig)
	}

	return proxy, nil
}

// bindHandler is used by New to store handler information on a proxy instance
func (p *Proxy) bindHandler(handlerConfig *config.Handler) {
	handler := &Handler{
		Name:            handlerConfig.Name,
		ExternalSubject: handlerConfig.ExternalSubject,
		InternalSubject: handlerConfig.InternalSubject,
	}

	handler.RequestRegexes = []([2]*regexp.Regexp){}

	for _, rr := range handlerConfig.RequestRegexes {
		var host *regexp.Regexp
		if len(rr[0]) > 0 {
			host = regexp.MustCompile(rr[0])
		}
		var path *regexp.Regexp
		if len(rr[1]) > 0 {
			path = regexp.MustCompile(rr[1])
		}
		handler.RequestRegexes = append(handler.RequestRegexes, [2]*regexp.Regexp{host, path})
	}

	p.handlers = append(p.handlers, handler)
}

// HandleRequest is gin middleware that transforms the HTTP request into a Nats
// message, then dispatches it to the correct service. HandleRequest must be
// run after RegisterHandlers otherwise it will cause a panic.
func (p *Proxy) HandleRequest(c *gin.Context) {
	var handler *Handler
	if _handler, ok := c.Get("handler"); ok {
		handler = _handler.(*Handler)
	} else {
		c.AbortWithStatus(http.StatusNotFound)
		return
	}
	log.Printf("Dispatching request to %v service\n", handler.Name)

	log.Println("Building Nats request from HTTP request")

	body, _ := c.Get("body")

	natsRequest := &extsrv.Request{
		Service:   handler.ExternalSubject,
		RequestID: c.MustGet("requestID").(string),
		RequestIP: c.ClientIP(),
		Method:    c.Request.Method,
		Path:      c.Request.URL.Path,
		Query:     c.Request.URL.Query(),
		Body:      body,
		Header:    c.Request.Header,
	}

	rawAuthenticationResponseBody, exists := c.Get("rawAuthenticationResponseBody")
	if exists {
		natsRequest.RawAuth = rawAuthenticationResponseBody.(json.RawMessage)
	}
	log.Println("Nats request built")

	natsResponse := &extsrv.Response{}
	log.Printf("Dispatching request to %v service via Nats\n", handler.Name)
	if err := extsrv.RequestReply(natsRequest, natsResponse); err != nil {
		// TODO Handle error
		return
	}
	log.Printf("Recieved response from %v service\n", handler.Name)

}

// GetHandler is gin middleware that populates the current handler (if present).
// GetHandler must be used before HandleRequest.
func (p *Proxy) GetHandler(c *gin.Context) {
	log.Println("path:" + c.Request.URL.Path)
	log.Println("Finding destination service for incoming request")
	handler := p.getHandlerFromGinContext(c)
	if handler != nil {
		c.Set("handler", handler)
		log.Printf("Service found: %v\n", handler.Name)
	} else {
		log.Println("Service not found")
	}

	c.Next()
}

// getHandlerFromGinContext takes a gin context and attempts to match a
// service handler to it. If a handler is found it a pointer to it is returned.
func (p *Proxy) getHandlerFromGinContext(c *gin.Context) *Handler {
	for _, handler := range p.handlers {
		for _, handlerRegexes := range handler.RequestRegexes {
			hostRegex := handlerRegexes[0]
			pathRegex := handlerRegexes[1]

			isHostMatch := hostRegex == nil || hostRegex.MatchString(c.Request.Host)
			isPathMatch := pathRegex == nil || pathRegex.MatchString(c.Request.URL.Path)

			if isHostMatch && isPathMatch {
				return handler
			}
		}
	}

	return nil
}

// Handler contains information required for matching HTTP requests to Nats
// services. It is used primarily by proxy.go, but is referenced by the status
// middleware as well.
type Handler struct {
	Name            string
	RequestRegexes  [][2]*regexp.Regexp
	ExternalSubject string
	InternalSubject string
}
