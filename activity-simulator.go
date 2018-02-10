package prom_http_simulator

import (
	"time"
	"sync"
	"math/rand"
	"github.com/prometheus/client_golang/prometheus"
)

var (
	httpRequestsTotal = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "http_requests_total",
			Help: "Total number of http requests by response status code",
		},
		[]string{"endpoint", "status"},
	)

	httpRequestDurationMs = prometheus.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    "http_request_duration_milliseconds",
			Help:    "Http request latency histogram",
			Buckets: prometheus.ExponentialBuckets(25, 2, 7),
		},
		[]string{"endpoint", "status"},
	)
)

func init() {
	prometheus.MustRegister(httpRequestsTotal, httpRequestDurationMs)
}

type Simulator struct {
	opts  SimulatorOpts
	mutex *sync.Mutex
	rng   *rand.Rand

	spikeMode bool
}

func NewSimulator(opts SimulatorOpts) *Simulator {
	return &Simulator{
		opts,
		&sync.Mutex{},
		rand.New(rand.NewSource(time.Now().UnixNano())),
		false,
	}
}

func (s *Simulator) UpdateOpts(opts SimulatorOpts) {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	s.opts = opts
}

func (s *Simulator) SetSpikeMode(mode string) {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	switch mode {
	case "on":
		s.opts.SpikeModeStatus = ON
	case "off":
		s.opts.SpikeModeStatus = OFF
	case "random":
		s.opts.SpikeModeStatus = RANDOM
	}
}

func (s *Simulator) SetErrorRate(rate int) {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	if rate > 100 {
		rate = 100
	}
	if rate < 0 {
		rate = 0
	}
	s.opts.ErrorRate = rate
}

func (s *Simulator) GetOpts() SimulatorOpts {
	return s.opts
}

func (s *Simulator) Run() {
	for {
		s.simulateActivity()
		time.Sleep(1 * time.Second)
	}
}

func (s *Simulator) simulateActivity() {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	requestRate := s.opts.RequestRate
	if s.giveSpikeMode() {
		requestRate *= 5 + s.rng.Intn(10)
	}

	nbRequests := s.giveWithUncertainty(requestRate, s.opts.RequestRateUncertainty)
	var nbErrors int
	for i := 0; i < nbRequests; i++ {
		statusCode := s.giveStatusCode()
		endpoint := s.giveEndpoint()
		httpRequestsTotal.WithLabelValues(endpoint, statusCode).Inc()
		latency := s.giveLatency(statusCode)
		if s.spikeMode {
			latency *= 2
		}
		if statusCode == "500" {
			nbErrors++
		}
		httpRequestDurationMs.WithLabelValues(endpoint, statusCode).Observe(float64(latency))
	}
}

func (s *Simulator) giveSpikeMode() bool {
	switch s.opts.SpikeModeStatus {
	case ON:
		s.spikeMode = true
	case OFF:
		s.spikeMode = false
	case RANDOM:
		n := s.rng.Intn(100)
		if !s.spikeMode && n < s.opts.SpikeStartChance {
			s.spikeMode = true
		} else if s.spikeMode && n < s.opts.SpikeEndChance {
			s.spikeMode = false
		}
	}
	return s.spikeMode
}

func (s *Simulator) giveWithUncertainty(n int, u int) int {
	delta := s.rng.Intn(n*u/50) - (n * u / 100)
	return n + delta
}

func (s *Simulator) giveStatusCode() string {
	if s.rng.Intn(100) < s.opts.ErrorRate {
		return "500"
	} else {
		return "200"
	}
}

func (s *Simulator) giveEndpoint() string {
	n := s.rng.Intn(len(s.opts.Endpoints))
	return s.opts.Endpoints[n]
}

func (s *Simulator) giveLatency(statusCode string) int {
	if statusCode != "200" {
		return 5+s.rng.Intn(50)
	}

	p := s.rng.Intn(100)
	if p < 50 {
		return s.giveWithUncertainty(s.opts.LatencyMin+s.rng.Intn(s.opts.LatencyP50-s.opts.LatencyMin), s.opts.LatencyUncertainty)
	}
	if p < 90 {
		return s.giveWithUncertainty(s.opts.LatencyP50+s.rng.Intn(s.opts.LatencyP90-s.opts.LatencyP50), s.opts.LatencyUncertainty)
	}
	if p < 99 {
		return s.giveWithUncertainty(s.opts.LatencyP90+s.rng.Intn(s.opts.LatencyP99-s.opts.LatencyP90), s.opts.LatencyUncertainty)
	}
	return s.giveWithUncertainty(s.opts.LatencyP99+s.rng.Intn(s.opts.LatencyMax-s.opts.LatencyP99), s.opts.LatencyUncertainty)
}

type SpikeMode int
const (
	OFF SpikeMode = iota
	ON
	RANDOM
)

type SimulatorOpts struct {
	// Endpoints Weighted map of endpoints to simulate
	Endpoints []string

	// RequestRate requests per second
	RequestRate int

	// RequestRateUncertainty Percentage of uncertainty when generating requests (+/-)
	RequestRateUncertainty int

	// LatencyMin in milliseconds
	LatencyMin int
	// LatencyP50 in milliseconds
	LatencyP50 int
	// LatencyP90 in milliseconds
	LatencyP90 int
	// LatencyP99 in milliseconds
	LatencyP99 int
	// LatencyMax in milliseconds
	LatencyMax int
	// LatencyUncertainty Percentage of uncertainty when generating latency (+/-)
	LatencyUncertainty int

	// ErrorRate Percentage of chance of requests causing 500
	ErrorRate int

	// SpikeModeStatus ON/OFF/RANDOM
	SpikeModeStatus SpikeMode

	// SpikeStartChance Percentage of chance of entering spike mode
	SpikeStartChance int

	// SpikeStartChance Percentage of chance of exiting spike mode
	SpikeEndChance int
}
