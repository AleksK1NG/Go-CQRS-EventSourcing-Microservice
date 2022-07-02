package elasticsearch_subscription

import (
	"context"
	"github.com/segmentio/kafka-go"
)

func (s *elasticSearchSubscription) commitMessage(ctx context.Context, r *kafka.Reader, m kafka.Message) {
	if err := r.CommitMessages(ctx, m); err != nil {
		s.log.Errorf("(elasticSearchSubscription) [CommitMessages] err: %v", err)
		return
	}
	s.log.KafkaLogCommittedMessage(m.Topic, m.Partition, m.Offset)
}

func (s *elasticSearchSubscription) commitErrMessage(ctx context.Context, r *kafka.Reader, m kafka.Message) {
	if err := r.CommitMessages(ctx, m); err != nil {
		s.log.Errorf("(elasticSearchSubscription) [CommitMessages] err: %v", err)
		return
	}
	s.log.KafkaLogCommittedMessage(m.Topic, m.Partition, m.Offset)
}

func (s *elasticSearchSubscription) logProcessMessage(m kafka.Message, workerID int) {
	s.log.KafkaProcessMessage(m.Topic, m.Partition, m.Value, workerID, m.Offset, m.Time)
}
