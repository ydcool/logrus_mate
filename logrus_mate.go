package logrus_mate

import (
	"sync"

	"github.com/Sirupsen/logrus"
)

var (
	defaultMate         *LogrusMate
	defaultMateInitOnce sync.Once
)

type LogrusMate struct {
	loggersLock sync.Mutex
	initalOnce  sync.Once

	loggers map[string]*logrus.Logger
}

func Logger(loggerName ...string) (logger *logrus.Logger) {
	if defaultMate == nil {
		defaultMateInitOnce.Do(func() {
			if defaultMate == nil {
				defaultMate = defaultLogrusMate()
			}
		})
	}

	return defaultMate.Logger(loggerName...)
}

func NewLogger(name string, conf LoggerConfig) (logger *logrus.Logger, err error) {
	return defaultMate.NewLogger(name, conf)
}

func (p LogrusMate) NewLogger(name string, conf LoggerConfig) (logger *logrus.Logger, err error) {
	tmpLogger := logrus.New()

	if conf.Formatter.Name == "" {
		conf.Formatter.Name = "text"
		conf.Formatter.Options = nil
	}

	var formatter logrus.Formatter
	if formatter, err = NewFormatter(conf.Formatter.Name, conf.Formatter.Options); err != nil {
		return
	}

	tmpLogger.Formatter = formatter

	if conf.Hooks != nil {
		for hookName, hookOpt := range conf.Hooks {
			var hook logrus.Hook
			if hook, err = NewHook(hookName, hookOpt); err != nil {
				return
			}
			tmpLogger.Hooks.Add(hook)
		}
	}

	var lvl = logrus.DebugLevel
	if lvl, err = logrus.ParseLevel(conf.Level); err != nil {
		return
	} else {
		tmpLogger.Level = lvl
	}

	logger = tmpLogger

	p.loggers[name] = logger

	return
}

func NewLogrusMate(mateConf LogrusMateConfig) (logrusMate *LogrusMate, err error) {
	mate := new(LogrusMate)

	if err = mate.inital(mateConf); err != nil {
		return
	}

	logrusMate = mate

	return
}

func (p *LogrusMate) inital(mateConf LogrusMateConfig) (err error) {
	p.loggersLock.Lock()
	defer p.loggersLock.Unlock()

	if err = mateConf.Validate(); err != nil {
		return
	}

	p.loggers = make(map[string]*logrus.Logger, len(mateConf.Loggers))

	p.initalOnce.Do(func() {

		runEnv := mateConf.RunEnv()

		for _, loggerConfs := range mateConf.Loggers {
			var conf LoggerConfig
			if loggerConf, exist := loggerConfs.Config[runEnv]; exist {
				conf = loggerConf
			} else {
				conf = defaultLoggerConfig()
			}

			if _, err = p.NewLogger(loggerConfs.Name, conf); err != nil {
				return
			}
		}

	})

	return
}

func (p *LogrusMate) Logger(loggerName ...string) (logger *logrus.Logger) {
	p.loggersLock.Lock()
	defer p.loggersLock.Unlock()

	name := ""
	if loggerName != nil && len(loggerName) == 1 {
		name = loggerName[0]
	}

	logger, _ = p.loggers[name]

	return
}

func defaultLogrusMate() (logrusMate *LogrusMate) {
	if mate, err := NewLogrusMate(defaultLogrusMateConfig()); err != nil {
		panic(err)
	} else {
		logrusMate = mate
	}
	return
}

func defaultLoggerConfig() LoggerConfig {
	return LoggerConfig{
		Level:     "debug",
		Formatter: FormatterConfig{Name: "text", Options: nil},
	}
}

func defaultLogrusMateConfig() LogrusMateConfig {
	return LogrusMateConfig{
		EnvironmentKeys: Environments{RunEnv: "development"},
		Loggers: []LoggerItem{
			{
				Name:   "",
				Config: map[string]LoggerConfig{"development": defaultLoggerConfig()}}},
	}
}
