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
- [X] share zkzone instance
- [ ] presence and standby mode
- [ ] graceful shutdown
- [ ] pack.Payload reuse memory
- [ ] DDL binlog
  - drop table y;
- [X] trace async producer Successes channel and mark as processed
- [X] metrics
- [X] telemetry and alert
- [X] what if replication conn broken
- [X] position will be stored in zk
- [X] play with binlog_row_image
- [ ] bug fix
  - next log position leads to failure after resume
  - when repliation stops, mysql show processlist still exists
  - table id issue
- [ ] test cases
  - [X] restart mysql master
  - [X] mysql kill process
  - [ ] race detection
  - [ ] tc drop network packets and high latency

