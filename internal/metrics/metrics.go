package metrics

import "github.com/prometheus/client_golang/prometheus"

type WorkerMetrics struct {
	TasksProcessed prometheus.Counter
	TasksFailed    prometheus.Counter
	TaskRetries    prometheus.Counter
	TaskDuration   prometheus.Histogram
	TaskQueueSize  prometheus.Gauge
}

func NewWorkerMetrics() *WorkerMetrics {
	return &WorkerMetrics{
		TasksProcessed: prometheus.NewCounter(
			prometheus.CounterOpts{
				Name: "tasks_processed_total",
				Help: "Total number of successfully processed tasks",
			},
		),

		TasksFailed: prometheus.NewCounter(
			prometheus.CounterOpts{
				Name: "tasks_failed_total",
				Help: "Total number of failed task processing attempts",
			},
		),

		TaskRetries: prometheus.NewCounter(
			prometheus.CounterOpts{
				Name: "task_retries_total",
				Help: "Total number of task retries",
			},
		),

		TaskDuration: prometheus.NewHistogram(
			prometheus.HistogramOpts{
				Name: "task_processing_duration_seconds",
				Help: "Task processing duration in seconds",
			},
		),

		TaskQueueSize: prometheus.NewGauge(
			prometheus.GaugeOpts{
				Name: "task_queue_size",
				Help: "Current number of tasks waiting in the Redis queue",
			},
		),
	}
}

func (m *WorkerMetrics) TaskProcessed() {
	m.TasksProcessed.Inc()
}

func (m *WorkerMetrics) TaskFailed() {
	m.TasksFailed.Inc()
}

func (m *WorkerMetrics) TaskRetried() {
	m.TaskRetries.Inc()
}

func (m *WorkerMetrics) ObserveTaskDuration(seconds float64) {
	m.TaskDuration.Observe(seconds)
}

func (m *WorkerMetrics) SetQueueSize(size int) {
	m.TaskQueueSize.Set(float64(size))
}

func (m *WorkerMetrics) Register() {
	prometheus.MustRegister(
		m.TasksProcessed,
		m.TasksFailed,
		m.TaskRetries,
		m.TaskDuration,
		m.TaskQueueSize,
	)
}
