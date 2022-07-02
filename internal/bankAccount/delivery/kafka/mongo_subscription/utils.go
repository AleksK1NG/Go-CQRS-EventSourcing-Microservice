package mongo_subscription

import (
	"context"
	"github.com/segmentio/kafka-go"
)

func (s *mongoSubscription) commitMessage(ctx context.Context, r *kafka.Reader, m kafka.Message) {
	if err := r.CommitMessages(ctx, m); err != nil {
		s.log.Errorf("(mongoSubscription) [CommitMessages] err: %v", err)
		return
	}
	s.log.KafkaLogCommittedMessage(m.Topic, m.Partition, m.Offset)
}

func (s *mongoSubscription) commitErrMessage(ctx context.Context, r *kafka.Reader, m kafka.Message) {
	if err := r.CommitMessages(ctx, m); err != nil {
		s.log.Errorf("(mongoSubscription) [CommitMessages] err: %v", err)
		return
	}
	s.log.KafkaLogCommittedMessage(m.Topic, m.Partition, m.Offset)
}

func (s *mongoSubscription) logProcessMessage(m kafka.Message, workerID int) {
	s.log.KafkaProcessMessage(m.Topic, m.Partition, m.Value, workerID, m.Offset, m.Time)
}
