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

### Roadmap

- pubsub audit reporter
- universal kafka listener and outputer

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
  - discard MarkAsProcessed
- [ ] integration with helix
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
  - [ ] kill dbusd, dbusd-slave did not leave cluster
  - [ ] router stat wrong
    Total:142,535,625      0.00B speed:22,671/s      0.00B/s max: 0.00B/0.00B
- [ ] test cases
  - [X] restart mysql master
  - [X] mysql kill process
  - [X] race detection
  - [ ] tc drop network packets and high latency
  - [X] reset binlog pos, and check kafka didn't recv dup events
- [ ] GTID
- [ ] ugly globals register myslave_key

### Memo

- mysqlbinlog input peak with mock output
  - 140k event per second
  - 30k row event per second
  - 200Mb network bandwidth
  - it takes 2h25m to zero lag for platform of 2d lag

