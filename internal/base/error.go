package base

type ErrorType uint8

const (
	ErrorTypeProtocol  = ErrorType(1)
	ErrorTypeTransport = ErrorType(2)
	ErrorTypeReply     = ErrorType(3)
	ErrorTypeRuntime   = ErrorType(4)
	ErrorTypeKernel    = ErrorType(5)
	ErrorTypeSecurity  = ErrorType(6)
)

type ErrorLevel uint8

const (
	ErrorLevelWarn  = ErrorLevel(1)
	ErrorLevelError = ErrorLevel(2)
	ErrorLevelFatal = ErrorLevel(3)
)

type ErrorCode uint32

type Error struct {
	code    uint64
	message string
}

func defineError(
	kind ErrorType,
	code ErrorCode,
	level ErrorLevel,
	message string,
) *Error {
	errCode := (uint64(kind) << 56) & (uint64(level) << 48) & (uint64(code) << 16)
	return &Error{
		code:    errCode,
		message: message,
	}
}

// DefineProtocolError ...
func DefineProtocolError(code ErrorCode, level ErrorLevel, msg string) *Error {
	return defineError(ErrorTypeProtocol, code, level, msg)
}

// DefineTransportError ...
func DefineTransportError(code ErrorCode, level ErrorLevel, msg string) *Error {
	return defineError(ErrorTypeTransport, code, level, msg)
}

// DefineReplyError ...
func DefineReplyError(code ErrorCode, level ErrorLevel, msg string) *Error {
	return defineError(ErrorTypeReply, code, level, msg)
}

// DefineRuntimeError ...
func DefineRuntimeError(code ErrorCode, level ErrorLevel, msg string) *Error {
	return defineError(ErrorTypeRuntime, code, level, msg)
}

// DefineKernelError ...
func DefineKernelError(code ErrorCode, level ErrorLevel, msg string) *Error {
	return defineError(ErrorTypeKernel, code, level, msg)
}

// DefineSecurityError ...
func DefineSecurityError(code ErrorCode, level ErrorLevel, msg string) *Error {
	return defineError(ErrorTypeSecurity, code, level, msg)
}

func (p *Error) GetType() ErrorType {
	return ErrorType(p.code >> 56)
}

func (p *Error) GetLevel() ErrorLevel {
	return ErrorLevel((p.code >> 48) & 0xFF)
}

func (p *Error) GetCode() ErrorCode {
	return ErrorCode((p.code >> 16) & 0xFFFFFFFF)
}

func (p *Error) GetMessage() string {
	return p.message
}

func (p *Error) AddDebug(debug string) *Error {
	ret := &Error{}

	if p.code%2 == 0 {
		ret.code = p.code + 1
	} else {
		ret.code = p.code
	}

	if p.message == "" {
		ret.message = debug
	} else {
		ret.message = ConcatString(p.message, "\n", debug)
	}

	return ret
}

func (p *Error) getErrorTypeString() string {
	switch p.GetType() {
	case ErrorTypeProtocol:
		return "Protocol"
	case ErrorTypeTransport:
		return "Transport"
	case ErrorTypeReply:
		return "Reply"
	case ErrorTypeRuntime:
		return "Runtime"
	case ErrorTypeKernel:
		return "Kernel"
	case ErrorTypeSecurity:
		return "Security"
	default:
		return ""
	}
}

func (p *Error) getErrorLevelString() string {
	switch p.GetLevel() {
	case ErrorLevelWarn:
		return "Warn"
	case ErrorLevelError:
		return "Error"
	case ErrorLevelFatal:
		return "Fatal"
	default:
		return ""
	}
}

func (p *Error) Error() string {
	return ConcatString(
		p.getErrorTypeString(),
		p.getErrorLevelString(),
		":",
		p.message,
	)
}
