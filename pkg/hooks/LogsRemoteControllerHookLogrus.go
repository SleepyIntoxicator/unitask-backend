package hooks

import (
	"errors"
	log "github.com/sirupsen/logrus"
	"net"
	"time"
)

type logItem struct {
	Level   log.Level  `json:"level"`
	Time    time.Time  `json:"time"`
	Message string     `json:"message"`
	Data    log.Fields `json:"data"`
}

type LogsRemoteControllerHook struct {
	logs                     []log.Entry
	target                   net.Addr //Адрес удалённого лог-сервера
	isRemoteServerMustListen bool     //Должен ли быть включен удалённый сервер
	isHookEnabled            bool     //Включен ли хук
	SendOnline               bool     //Отсылать ли логи в режиме онлайн
}

func NewLogRemoteController(targetAddr net.Addr, mustListen bool, enabledOnStart bool) *LogsRemoteControllerHook {
	err := errors.New("")

	//listener, err := net.Listen("tcp", targetAddr.String())
	if err == net.ErrClosed {

	}
	//conn, err := listener.Accept()
	if err == net.ErrClosed {

	}

	net.Dial("tcp", ":8888")


	return &LogsRemoteControllerHook{
		target:                   targetAddr,
		isRemoteServerMustListen: mustListen,
		isHookEnabled:            enabledOnStart,
	}
}

func (h *LogsRemoteControllerHook) Levels() []log.Level {
	return log.AllLevels
}

func (h *LogsRemoteControllerHook) Fire(e *log.Entry) error {
	h.logs = append(h.logs, *e)

	if h.SendOnline {
		/*log := logItem{
			Level:   e.Level,
			Time:    e.Time,
			Message: e.Message,
			Data:    e.Data,
		}*/

		h.target.Network()
	}

	return nil
}

