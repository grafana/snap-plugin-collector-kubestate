package plugin

import (
	"errors"
	"fmt"
	"time"

	"google.golang.org/grpc/metadata"

	"golang.org/x/net/context"

	"github.com/intelsdi-x/snap-plugin-lib-go/v1/plugin/rpc"
	log "github.com/sirupsen/logrus"
)

const (
	defaultMaxCollectDuration = 10 * time.Second
	defaultMaxMetricsBuffer   = 0
)

type StreamProxy struct {
	pluginProxy
	plugin StreamCollector
	ctx    context.Context

	// maxMetricsBuffer is the maximum number of metrics the plugin is buffering before sending metrics.
	// Defaults to zero what means send metrics immediately.
	maxMetricsBuffer int64

	// maxCollectionDuration sets the maximum duration (always greater than 0s) between collections before metrics are sent.
	// Defaults to 10s what means that after 10 seconds no new metrics are received, the plugin should send
	// whatever data it has in the buffer instead of waiting longer.
	maxCollectDuration time.Duration

	sendChan chan []Metric
	recvChan chan []Metric
	errChan  chan string
}

func (p *StreamProxy) GetMetricTypes(ctx context.Context, arg *rpc.GetMetricTypesArg) (*rpc.MetricsReply, error) {
	cfg := fromProtoConfig(arg.Config)

	r, err := p.plugin.GetMetricTypes(cfg)
	if err != nil {
		return nil, err
	}
	metrics := []*rpc.Metric{}
	for _, mt := range r {
		// We can ignore this error since we are not returning data from GetMetricTypes.
		metric, _ := toProtoMetric(mt)
		metrics = append(metrics, metric)
	}
	reply := &rpc.MetricsReply{
		Metrics: metrics,
	}
	return reply, nil
}

func (p *StreamProxy) StreamMetrics(stream rpc.StreamCollector_StreamMetricsServer) error {
	log.WithFields(
		log.Fields{
			"_block": "StreamMetrics",
		},
	).Debug("streaming started")
	if stream == nil {
		return errors.New("Stream metrics server is nil")
	}

	// Error channel where we will forward plugin errors to snap where it
	// can report/handle them.
	p.errChan = make(chan string)
	// Metrics into the plugin from snap.
	p.recvChan = make(chan []Metric)
	// Metrics out of the plugin into snap.
	p.sendChan = make(chan []Metric)
	// context for communicating that the stream has been closed to the plugin author

	md, ok := metadata.FromIncomingContext(stream.Context())
	if !ok {
		log.Debug("No metadata")
	}

	taskID := "not-set"
	if tempVal, ok := md["task-id"]; ok {
		if len(tempVal) == 1 {
			taskID = tempVal[0]
		} else {
			log.Debug("Skipping assignment of metadata")
		}
	}

	go p.metricSend(taskID, p.sendChan, stream)
	go p.errorSend(p.errChan, stream)
	go p.streamRecv(taskID, p.recvChan, stream)

	return p.plugin.StreamMetrics(stream.Context(), p.recvChan, p.sendChan, p.errChan)

}

func (p *StreamProxy) errorSend(errChan chan string, stream rpc.StreamCollector_StreamMetricsServer) {
	for {
		select {
		case <-stream.Context().Done():
			return
		case r := <-errChan:
			reply := &rpc.CollectReply{
				Error: &rpc.ErrReply{Error: r},
			}
			if err := stream.Send(reply); err != nil {
				fmt.Println(err.Error())
			}
		}
	}
}

func (p *StreamProxy) metricSend(taskID string, ch chan []Metric, stream rpc.StreamCollector_StreamMetricsServer) {
	log.WithFields(
		log.Fields{
			"_block":             "metricSend",
			"task-id":            taskID,
			"maxMetricsBuffer":   p.maxMetricsBuffer,
			"maxCollectDuration": p.maxCollectDuration,
		},
	).Debug("starting routine for sending metrics")
	metrics := []*rpc.Metric{}

	afterCollectDuration := time.After(p.maxCollectDuration)
	for {
		select {
		case mts := <-ch:
			if len(mts) == 0 {
				break
			}

			for _, mt := range mts {
				metric, err := toProtoMetric(mt)
				if err != nil {
					fmt.Println(err.Error())
					break
				}
				metrics = append(metrics, metric)

				// send metrics if maxMetricsBuffer is reached
				// (notice it is only possible for maxMetricsBuffer greater than 0)
				if p.maxMetricsBuffer == int64(len(metrics)) {
					sendReply(taskID, metrics, stream)
					metrics = []*rpc.Metric{}
				}
			}

			// send all available metrics immediately for maxMetricsBuffer is 0 (defaults)
			if p.maxMetricsBuffer == 0 {
				sendReply(taskID, metrics, stream)
				metrics = []*rpc.Metric{}
				afterCollectDuration = time.After(p.maxCollectDuration)
			}

		case <-afterCollectDuration:
			// send metrics if maxCollectDuration is reached
			sendReply(taskID, metrics, stream)
			metrics = []*rpc.Metric{}
			afterCollectDuration = time.After(p.maxCollectDuration)
		case <-stream.Context().Done():
			return
		}
	}
}

func (p *StreamProxy) streamRecv(taskID string, ch chan []Metric, stream rpc.StreamCollector_StreamMetricsServer) {
	logger := log.WithFields(
		log.Fields{
			"_block":  "streamRecv",
			"task-id": taskID,
		},
	)
	logger.Debug("starting routine for receiving metrics")
	for {
		select {
		case <-stream.Context().Done():
			close(ch)
			return
		default:

			s, err := stream.Recv()
			if err != nil {
				logger.Error(err)
				break
			}
			if s != nil {
				if s.MaxMetricsBuffer > 0 {
					logger.WithFields(log.Fields{
						"option": "max-metrics-buffer",
						"value":  s.MaxMetricsBuffer,
					}).Debug("setting max metrics buffer option")
					p.setMaxMetricsBuffer(s.MaxMetricsBuffer)
				}
				if s.MaxCollectDuration > 0 {
					logger.WithFields(log.Fields{
						"option": "max-collect-duration",
						"value":  fmt.Sprintf("%v seconds", time.Duration(s.MaxCollectDuration).Seconds()),
					}).Debug("setting max collect duration option")
					p.setMaxCollectDuration(time.Duration(s.MaxCollectDuration))
				}
				if s.Metrics_Arg != nil {
					metrics := []Metric{}
					for _, mt := range s.Metrics_Arg.Metrics {
						metric := fromProtoMetric(mt)
						metrics = append(metrics, metric)
					}
					// send requested metrics to be collected into the stream plugin
					ch <- metrics
				}
			}
		}
	}
}

func (p *StreamProxy) setMaxCollectDuration(d time.Duration) {
	p.maxCollectDuration = d
}

func (p *StreamProxy) setMaxMetricsBuffer(i int64) {
	p.maxMetricsBuffer = i
}

func sendReply(taskID string, metrics []*rpc.Metric, stream rpc.StreamCollector_StreamMetricsServer) {
	logger := log.WithFields(
		log.Fields{
			"_block":  "sendReply",
			"task-id": taskID,
		},
	)
	if len(metrics) == 0 {
		logger.Debug("No metrics available to send")
		return
	}

	reply := &rpc.CollectReply{
		Metrics_Reply: &rpc.MetricsReply{Metrics: metrics},
	}

	if err := stream.Send(reply); err != nil {
		logger.Error(err)
	}

	logger.WithFields(
		log.Fields{
			"count": len(metrics),
		},
	).Debug("sending metrics")
}
