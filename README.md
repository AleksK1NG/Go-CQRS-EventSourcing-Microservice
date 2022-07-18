Go EventSourcing and CQRS with PostgreSQL, Kafka, MongoDB and ElasticSearch 👋✨💫

#### 👨‍💻 Full list what has been used:
[PostgeSQL](https://github.com/jackc/pgx) as event store<br/>
[Kafka](https://github.com/segmentio/kafka-go) as messages broker<br/>
[gRPC](https://github.com/grpc/grpc-go) Go implementation of gRPC<br/>
[Jaeger](https://www.jaegertracing.io/) open source, end-to-end distributed [tracing](https://opentracing.io/)<br/>
[Prometheus](https://prometheus.io/) monitoring and alerting<br/>
[Grafana](https://grafana.com/) for to compose observability dashboards with everything from Prometheus<br/>
[MongoDB](https://github.com/mongodb/mongo-go-driver) Web and API based SMTP testing<br/>
[Elasticsearch](https://github.com/elastic/go-elasticsearch) Elasticsearch client for Go.<br/>
[Echo](https://github.com/labstack/echo) web framework<br/>
[Kibana](https://github.com/labstack/echo) Kibana is user interface that lets you visualize your Elasticsearch<br/>
[Migrate](https://github.com/golang-migrate/migrate) for migrations<br/>


### Jaeger UI:

http://localhost:16686

### Prometheus UI:

http://localhost:9090

### Grafana UI:

http://localhost:3005

### Kibana UI:

http://localhost:5601/app/home#/


For local development 🙌👨‍💻🚀:

```
make local // for run docker compose
make run_es // run microservice
```
or
```
make develop // run all in docker compose with hot reload
```