package es

const (
	saveEventQuery = `INSERT INTO microservices.events as e (aggregate_id, aggregate_type, event_type, data, version, metadata, timestamp)
	VALUES ($1, $2, $3, $4, $5, $6, now())`

	getEventsQuery = `SELECT event_id, aggregate_id, aggregate_type, event_type, data, version, timestamp, metadata 
	FROM microservices.events e WHERE aggregate_id = $1 ORDER BY version ASC`

	getEventQuery = `SELECT aggregate_id FROM microservices.events e WHERE aggregate_id = $1`

	getEventsByVersionQuery = `SELECT event_id, aggregate_id, aggregate_type, event_type, data, version, timestamp, metadata 
	FROM microservices.events e WHERE aggregate_id = $1 AND version > $2 ORDER BY version ASC`

	saveSnapshotQuery = `INSERT INTO microservices.snapshots (aggregate_id, aggregate_type, data, version, timestamp) 
	VALUES ($1, $2, $3, $4, now()) ON CONFLICT (aggregate_id) DO UPDATE 
	SET data = $3, version = $4, timestamp = now()`

	getSnapshotQuery = `SELECT aggregate_id, aggregate_type, data, version FROM microservices.snapshots s WHERE aggregate_id = $1`

	handleConcurrentWriteQuery = `SELECT aggregate_id FROM microservices.events e WHERE e.aggregate_id = $1 LIMIT 1 FOR UPDATE`
)
