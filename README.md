# dbus
yet another databus that listens for mysql binlog and distribute to sinks

### Plugins

#### Input

- MockInput
- MysqlbinlogInput
- RedisbinlogInput

#### Output

- MockOutput
- KafkaOutput

### TODO

- [X] logging
- [X] share zkzone instance
- [X] presence and standby mode
- [X] graceful shutdown
- [X] master must drain before leave cluster
- [ ] KafkaOutput metrics
  -  binlog tps
  -  kafka tps
  -  lag
- [ ] zk checkpoint vs kafka checkpoint
- [ ] can a mysql instance with miltiple databases have multiple Log/Position?
- [ ] pack.Payload reuse memory
- [ ] kafka sync produce in batch
- [ ] DDL binlog
  - drop table y;
- [X] trace async producer Successes channel and mark as processed
- [X] metrics
- [X] telemetry and alert
- [X] visualize pipeline

  - ![dashboard](https://github.com/funkygao/dbus/blob/master/misc/resources/diagram.png)

- [X] what if replication conn broken
- [X] position will be stored in zk
- [X] play with binlog_row_image
- [ ] project feature for multi-tenant
- [ ] bug fix
  - [ ] next log position leads to failure after resume
  - [ ] when replication stops, mysql show processlist still exists
  - [ ] table id issue
  - [ ] what if invalid position
- [ ] test cases
  - [X] restart mysql master
  - [X] mysql kill process
  - [X] race detection
  - [ ] tc drop network packets and high latency
  - [X] reset binlog pos, and check kafka didn't recv dup events
- [ ] GTID
- [ ] ugly globals register myslave_key
