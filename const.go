package eudore


const ( 
	// Header
	HeaderAccept				=	"Accept"
	HeaderAcceptEncoding		=	"Accept-Encoding"
	HeaderContentType			=	"Content-Type"
	HeaderContentLength			=	"Content-Length"
	HeaderContentEncoding		=	"Content-Encoding"
	HeaderHost					=	"Host"
	HeaderMethod				=	"Method"
	HeaderReferer				=	"Referer"
	HeaderCookie				=	"Cookie"
	HeaderAuthorization			=	"Authorization"
	HeaderLocation				=	"Location"
	HeaderUpgrade				=	"Upgrade"
	HeaderConnection			=	"Connection"
	HeaderVary					=	"Vary"
	HeaderXForwardedFor			=	"X-Forwarded-For"
	HeaderUserAgent				=	"User-Agent"
	HeaderXRequestID			=	"X-Request-ID"
	HeaderXParentID				=	"X-Parent-ID"
	// Mime
	MimeCharsetUtf8					=	"charset=utf-8"
	MimeText						=	"text/*"
	MimeTextPlain					=	"text/plain"
	MimeTextPlainCharsetUtf8		=	MimeTextPlain + "; " + MimeCharsetUtf8
	MimeTextHTML					=	"text/html"
	MimeTextHTMLCharsetUtf8			=	MimeTextHTML + "; " + MimeCharsetUtf8
	MimeTextCss						=	"text/css"
	MimeTextCssUtf8					=	MimeTextCss + "; " + MimeCharsetUtf8
	MimeTextJavascript				=	"text/javascript"
	MimeTextJavascriptUtf8			=	MimeTextJavascript + "; " + MimeCharsetUtf8
	MimeTextMarkdown				=	"text/markdown"
	MimeTextMarkdownUtf8			=	MimeTextMarkdown + "; " + MimeCharsetUtf8
	MimeTextXml						=	"text/xml"
	MimeTextXmlCharsetUtf8			=	MimeTextXml + "; " + MimeCharsetUtf8
	MimeApplicationJson				=	"application/json"
	MimeApplicationJsonUtf8			=	MimeApplicationJson + "; " + MimeCharsetUtf8
	MimeApplicationXml				=	"application/xml"
	MimeApplicationxmlCharsetUtf8	=	MimeApplicationXml + "; " + MimeCharsetUtf8
	MimeApplicationForm				=	"application/x-www-form-urlencoded"
	MimeMultipartForm				=	"multipart/form-data"
	// Param
	ParamAction				=	"Action"
	ParamRam				=	"Ram"
	ParamTemplate			=	"Template"
	ParamRoute				=	"Route"
	ParamRoutes				=	"Routes"
	ParamRoutePath			=	"Route-Path"
	ParamRouteMethod		=	"Route-Method"
	// Param value
	ValueJwt				=	"jwt"
	ValueSession			=	"session"
)
