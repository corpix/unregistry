package server

import (
	"net"
	"net/http"

	echo "github.com/labstack/echo/v4"
	echomw "github.com/labstack/echo/v4/middleware"

	"git.backbone/corpix/unregistry/pkg/log"
	"git.backbone/corpix/unregistry/pkg/server/middleware"
	telemetry "git.backbone/corpix/unregistry/pkg/telemetry/registry"
)

type (
	HTTPServer      = http.Server
	HTTPOption      = func(*HTTPServer)
	HTTPHandler     = http.Handler
	HTTPHandlerFunc = http.HandlerFunc

	MiddlewareFunc = echo.MiddlewareFunc
	HandlerFunc    = echo.HandlerFunc

	Server struct{ *echo.Echo }

	Headers        = http.Header
	Context        = echo.Context
	Request        = http.Request
	Response       = echo.Response
	ResponseWriter = http.ResponseWriter

	Route  = echo.Route
	Router interface {
		Use(...MiddlewareFunc)
		Router(prefix string, m ...MiddlewareFunc) Router
		CONNECT(path string, h HandlerFunc, m ...MiddlewareFunc) *Route
		DELETE(path string, h HandlerFunc, m ...MiddlewareFunc) *Route
		GET(path string, h HandlerFunc, m ...MiddlewareFunc) *Route
		HEAD(path string, h HandlerFunc, m ...MiddlewareFunc) *Route
		OPTIONS(path string, h HandlerFunc, m ...MiddlewareFunc) *Route
		PATCH(path string, h HandlerFunc, m ...MiddlewareFunc) *Route
		POST(path string, h HandlerFunc, m ...MiddlewareFunc) *Route
		PUT(path string, h HandlerFunc, m ...MiddlewareFunc) *Route
		TRACE(path string, h HandlerFunc, m ...MiddlewareFunc) *Route
		Any(path string, h HandlerFunc, m ...MiddlewareFunc) []*Route
	}

	router struct{ *echo.Group }
)

const (
	StatusContinue           = http.StatusContinue
	StatusSwitchingProtocols = http.StatusSwitchingProtocols
	StatusProcessing         = http.StatusProcessing
	StatusEarlyHints         = http.StatusEarlyHints

	StatusOK                   = http.StatusOK
	StatusCreated              = http.StatusCreated
	StatusAccepted             = http.StatusAccepted
	StatusNonAuthoritativeInfo = http.StatusNonAuthoritativeInfo
	StatusNoContent            = http.StatusNoContent
	StatusResetContent         = http.StatusResetContent
	StatusPartialContent       = http.StatusPartialContent
	StatusMultiStatus          = http.StatusMultiStatus
	StatusAlreadyReported      = http.StatusAlreadyReported
	StatusIMUsed               = http.StatusIMUsed

	StatusMultipleChoices   = http.StatusMultipleChoices
	StatusMovedPermanently  = http.StatusMovedPermanently
	StatusFound             = http.StatusFound
	StatusSeeOther          = http.StatusSeeOther
	StatusNotModified       = http.StatusNotModified
	StatusUseProxy          = http.StatusUseProxy
	StatusTemporaryRedirect = http.StatusTemporaryRedirect
	StatusPermanentRedirect = http.StatusPermanentRedirect

	StatusBadRequest                   = http.StatusBadRequest
	StatusUnauthorized                 = http.StatusUnauthorized
	StatusPaymentRequired              = http.StatusPaymentRequired
	StatusForbidden                    = http.StatusForbidden
	StatusNotFound                     = http.StatusNotFound
	StatusMethodNotAllowed             = http.StatusMethodNotAllowed
	StatusNotAcceptable                = http.StatusNotAcceptable
	StatusProxyAuthRequired            = http.StatusProxyAuthRequired
	StatusRequestTimeout               = http.StatusRequestTimeout
	StatusConflict                     = http.StatusConflict
	StatusGone                         = http.StatusGone
	StatusLengthRequired               = http.StatusLengthRequired
	StatusPreconditionFailed           = http.StatusPreconditionFailed
	StatusRequestEntityTooLarge        = http.StatusRequestEntityTooLarge
	StatusRequestURITooLong            = http.StatusRequestURITooLong
	StatusUnsupportedMediaType         = http.StatusUnsupportedMediaType
	StatusRequestedRangeNotSatisfiable = http.StatusRequestedRangeNotSatisfiable
	StatusExpectationFailed            = http.StatusExpectationFailed
	StatusTeapot                       = http.StatusTeapot
	StatusMisdirectedRequest           = http.StatusMisdirectedRequest
	StatusUnprocessableEntity          = http.StatusUnprocessableEntity
	StatusLocked                       = http.StatusLocked
	StatusFailedDependency             = http.StatusFailedDependency
	StatusTooEarly                     = http.StatusTooEarly
	StatusUpgradeRequired              = http.StatusUpgradeRequired
	StatusPreconditionRequired         = http.StatusPreconditionRequired
	StatusTooManyRequests              = http.StatusTooManyRequests
	StatusRequestHeaderFieldsTooLarge  = http.StatusRequestHeaderFieldsTooLarge
	StatusUnavailableForLegalReasons   = http.StatusUnavailableForLegalReasons

	StatusInternalServerError           = http.StatusInternalServerError
	StatusNotImplemented                = http.StatusNotImplemented
	StatusBadGateway                    = http.StatusBadGateway
	StatusServiceUnavailable            = http.StatusServiceUnavailable
	StatusGatewayTimeout                = http.StatusGatewayTimeout
	StatusHTTPVersionNotSupported       = http.StatusHTTPVersionNotSupported
	StatusVariantAlsoNegotiates         = http.StatusVariantAlsoNegotiates
	StatusInsufficientStorage           = http.StatusInsufficientStorage
	StatusLoopDetected                  = http.StatusLoopDetected
	StatusNotExtended                   = http.StatusNotExtended
	StatusNetworkAuthenticationRequired = http.StatusNetworkAuthenticationRequired

	//

	HeaderAccept                          = echo.HeaderAccept
	HeaderAcceptEncoding                  = echo.HeaderAcceptEncoding
	HeaderAllow                           = echo.HeaderAllow
	HeaderAuthorization                   = echo.HeaderAuthorization
	HeaderContentDisposition              = echo.HeaderContentDisposition
	HeaderContentEncoding                 = echo.HeaderContentEncoding
	HeaderContentLength                   = echo.HeaderContentLength
	HeaderContentType                     = echo.HeaderContentType
	HeaderCookie                          = echo.HeaderCookie
	HeaderSetCookie                       = echo.HeaderSetCookie
	HeaderIfModifiedSince                 = echo.HeaderIfModifiedSince
	HeaderLastModified                    = echo.HeaderLastModified
	HeaderLocation                        = echo.HeaderLocation
	HeaderUpgrade                         = echo.HeaderUpgrade
	HeaderVary                            = echo.HeaderVary
	HeaderWWWAuthenticate                 = echo.HeaderWWWAuthenticate
	HeaderXForwardedFor                   = echo.HeaderXForwardedFor
	HeaderXForwardedProto                 = echo.HeaderXForwardedProto
	HeaderXForwardedProtocol              = echo.HeaderXForwardedProtocol
	HeaderXForwardedSsl                   = echo.HeaderXForwardedSsl
	HeaderXUrlScheme                      = echo.HeaderXUrlScheme
	HeaderXHTTPMethodOverride             = echo.HeaderXHTTPMethodOverride
	HeaderXRealIP                         = echo.HeaderXRealIP
	HeaderXRequestID                      = echo.HeaderXRequestID
	HeaderXRequestedWith                  = echo.HeaderXRequestedWith
	HeaderServer                          = echo.HeaderServer
	HeaderOrigin                          = echo.HeaderOrigin
	HeaderAccessControlRequestMethod      = echo.HeaderAccessControlRequestMethod
	HeaderAccessControlRequestHeaders     = echo.HeaderAccessControlRequestHeaders
	HeaderAccessControlAllowOrigin        = echo.HeaderAccessControlAllowOrigin
	HeaderAccessControlAllowMethods       = echo.HeaderAccessControlAllowMethods
	HeaderAccessControlAllowHeaders       = echo.HeaderAccessControlAllowHeaders
	HeaderAccessControlAllowCredentials   = echo.HeaderAccessControlAllowCredentials
	HeaderAccessControlExposeHeaders      = echo.HeaderAccessControlExposeHeaders
	HeaderAccessControlMaxAge             = echo.HeaderAccessControlMaxAge
	HeaderStrictTransportSecurity         = echo.HeaderStrictTransportSecurity
	HeaderXContentTypeOptions             = echo.HeaderXContentTypeOptions
	HeaderXXSSProtection                  = echo.HeaderXXSSProtection
	HeaderXFrameOptions                   = echo.HeaderXFrameOptions
	HeaderContentSecurityPolicy           = echo.HeaderContentSecurityPolicy
	HeaderContentSecurityPolicyReportOnly = echo.HeaderContentSecurityPolicyReportOnly
	HeaderXCSRFToken                      = echo.HeaderXCSRFToken
	HeaderReferrerPolicy                  = echo.HeaderReferrerPolicy

	QueryDelimiter = "?"
)

func StatusText(code int) string            { return http.StatusText(code) }
func WrapHandler(h HTTPHandler) HandlerFunc { return echo.WrapHandler(h) }

//

func (r *router) Router(prefix string, m ...MiddlewareFunc) Router {
	return &router{Group: r.Group.Group(prefix, m...)}
}

func (s *Server) Router(prefix string, m ...MiddlewareFunc) Router {
	return &router{Group: s.Echo.Group(prefix, m...)}
}

//

func ComposeMiddleware(mw ...MiddlewareFunc) MiddlewareFunc {
	return func(h HandlerFunc) HandlerFunc {
		var (
			n       = len(mw) - 1
			handler = h
		)
		for n >= 0 {
			handler = mw[n](handler)
			n--
		}

		return handler
	}
}

//

func HTTPTimeoutOption(c TimeoutConfig) HTTPOption {
	return func(s *HTTPServer) {
		s.ReadTimeout = c.Read
		s.WriteTimeout = c.Write
	}
}

func NewHTTPServer(addr string, options ...HTTPOption) *HTTPServer {
	s := &HTTPServer{Addr: addr}
	for _, fn := range options {
		fn(s)
	}
	return s
}

func New(c Config, subsystem string, l log.Logger, r *telemetry.Registry) (*Server, error) {
	e := echo.New()
	e.HideBanner = true
	e.Logger = &middleware.Logger{Logger: l}
	e.HTTPErrorHandler = DefaultHTTPErrorHandler

	ipExtractorOptions := make([]TrustOption, len(c.IPExtractor.TrustCIDR))
	for n, cidr := range c.IPExtractor.TrustCIDR {
		_, ipnet, err := net.ParseCIDR(cidr)
		if err != nil {
			return nil, err
		}

		ipExtractorOptions[n] = TrustIPRange(ipnet)
	}

	e.IPExtractor = echo.IPExtractor(ExtractIPFromRealIPHeader(ipExtractorOptions...))

	//

	e.Use(echomw.RequestID())
	e.Use(middleware.NewLogger(l, ""))
	e.Use(middleware.NewTelemetry(r, subsystem))
	e.Use(middleware.NewRecover(nil, l))

	return &Server{Echo: e}, nil
}
