package hooks

import (
	log "github.com/sirupsen/logrus"
)

type LogsStoreHook struct {
	logs	[]log.Entry
}

func (h *LogsStoreHook) Levels() []log.Level {
	return log.AllLevels
}

func (h *LogsStoreHook) Fire(e *log.Entry) error {
	h.logs = append(h.logs, *e)

	return nil
}

func (h *LogsStoreHook) GetLogsText() []string {
	var msg []string
	for i := range h.logs {
		m, _ := h.logs[i].String()
		msg = append(msg, m)
	}
	return msg
}

func (h *LogsStoreHook) GetAllItems() []log.Entry {
	return h.logs
}