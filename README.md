# dbus
another databus that listens for mysql binlog and distribute to sinks

### Plugins

#### Input

- MysqlbinlogInput
- MockInput

#### Output

- KafkaOutput
- PubOutput
- MockOutput

### TODO

- [X] logging
- [ ] graceful shutdown
- [X] metrics
- [X] telemetry and alert
- [ ] master/slave
- [ ] pack.Payload reuse memory
- [ ] what if replication conn broken
- [X] position will be stored in zk
- [X] play with binlog_row_image
- [ ] bug fix
  - next log position leads to failure after resume
- [ ] test cases
  - restart mysql master
  - mysql kill process
  - tc drop network packets and high latency
