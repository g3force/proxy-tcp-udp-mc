package proxy

import (
	"log"
	"sync"
	"time"
)

const maxDatagramSize = 8192

type Proxy interface {
	SetName(name string)
	SetVerbose(verbose bool)
	Start()
	Stop()
}

type StatsPrinter struct {
	numMessages    map[string]int
	lastReportTime time.Time
	reportInterval time.Duration
	mutex          sync.Mutex
}

func NewStatsPrinter() *StatsPrinter {
	return &StatsPrinter{
		numMessages:    map[string]int{},
		lastReportTime: time.Now(),
		reportInterval: time.Minute,
	}
}

func (s *StatsPrinter) NewMessage(classifier string) {
	s.mutex.Lock()
	defer s.mutex.Unlock()
	s.numMessages[classifier]++
	if time.Since(s.lastReportTime) > s.reportInterval {
		s.lastReportTime = time.Now()
		log.Printf("Message counts within %v: %v", s.reportInterval, s.numMessages)
		s.numMessages = map[string]int{}
	}
}
