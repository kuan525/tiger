package logger

var (
	defaultOptions = Options{
		logDir:     "/home/www/logs/applogs",
		fileName:   "default.log",
		maxSize:    500,
		maxAge:     1,
		maxBackups: 10,
		callerSkip: 1,
	}
)

type Options struct {
	logDir     string
	fileName   string
	maxSize    int
	maxBackups int
	maxAge     int
	compress   bool
	callerSkip int
}

type Option interface {
	apply(*Options)
}

type OptionFunc func(*Options)

func (o OptionFunc) apply(opts *Options) {
	o(opts)
}

func WithLogDir(dir string) Option {
	return OptionFunc(func(options *Options) {
		options.logDir = dir
	})
}

func WithHistoryLogFileName(fileName string) Option {
	return OptionFunc(func(options *Options) {
		options.fileName = fileName
	})
}

func WithMaxSize(size int) Option {
	return OptionFunc(func(options *Options) {
		options.maxSize = size
	})
}

func WithMaxBackups(backup int) Option {
	return OptionFunc(func(options *Options) {
		options.maxBackups = backup
	})
}

func WithMaxAge(maxAge int) Option {
	return OptionFunc(func(options *Options) {
		options.maxAge = maxAge
	})
}

func WithCompress(b bool) Option {
	return OptionFunc(func(options *Options) {
		options.compress = b
	})
}

func WithCallerSkip(skip int) Option {
	return OptionFunc(func(options *Options) {
		options.callerSkip = skip
	})
}
