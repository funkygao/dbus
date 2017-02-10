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
- [ ] topology in config
- [ ] trace async producer Successes channel and mark as processed
- [X] metrics
- [X] telemetry and alert
- [ ] presence and standby mode
- [ ] pack.Payload reuse memory
- [ ] what if replication conn broken
- [X] position will be stored in zk
- [X] play with binlog_row_image
- [ ] bug fix
  - next log position leads to failure after resume
  - when repliation stops, mysql show processlist still exists
- [ ] test cases
  - [X] restart mysql master
  - [X] mysql kill process
  - [ ] tc drop network packets and high latency

