package skeleton

// LogEntry structure used to log each HTTP request and response status code
type LogEntry struct {
	Host       string
	RemoteAddr string
	Method     string
	RequestURI string
	Proto      string
	Status     int
	ContentLen int
	UserAgent  string
	Duration   string
	RequestID  string
}
