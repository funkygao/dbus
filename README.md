# dbus
another databus that listens for mysql binlog and distribute to sinks

### Plugins

#### Input

- MysqlbinlogInput
- MockInput

#### Output

- KafkaOutput
- MockOutput

### TODO

- [ ] logging
- [ ] graceful shutdown
- [ ] metrics
- [ ] telemetry and alert
- [ ] master/slave
- [ ] pack.Payload reuse memory
