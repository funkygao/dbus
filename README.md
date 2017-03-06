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

    TPS         ACK     mode    msgsize
    ===         ===     ====    =======
    1200        all     sync    10KB
    2700        all     sync     1KB
    2100        local   sync    10KB
    4600        local   sync     1KB

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
- [ ] hub is shared, what if a plugin blocks others
- [ ] Batcher padding
- [ ] shutdown kafka
- [ ] filter to dispatch dbs of a single binlog to different output
- [ ] zk checkpoint vs kafka checkpoint
  - discard MarkAsProcessed
- [ ] visualized flow throughput like nifi
- [ ] kafka follower stops replication
    [2017-02-17 07:59:52,581] INFO Reconnect due to socket error: java.lang.OutOfMemoryError: Direct buffer memory (kafka.consumer.SimpleConsumer)
    [2017-02-17 07:59:52,780] WARN [ReplicaFetcherThread-2-1], Error in fetch Name: FetchRequest; Version: 0; CorrelationId: 4; ClientId: ReplicaFetcherThread-2-1; ReplicaId: 0; MaxWait: 500 ms; MinBytes: 1 bytes; RequestInfo: [pubaudit,0] -> PartitionFetchInfo(257984707,1048576). Possible cause: java.lang.OutOfMemoryError: Direct buffer memory (kafka.server.ReplicaFetcherThread)

    [2017-02-06 11:45:07,387] INFO [ReplicaFetcherManager on broker 6] Removed fetcher for partitions [nginx_access_log,4] (kafka.server.ReplicaFetcherManager)
    [2017-02-17 07:59:52,781] INFO Reconnect due to socket error: java.nio.channels.ClosedChannelException (kafka.consumer.SimpleConsumer)
- [ ] integration with helix
  - place config to central zk znode and watch changes
- [ ] can a mysql instance with miltiple databases have multiple Log/Position?
- [ ] pack.Payload reuse memory, json.NewEncoder(os.Stdout)
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
  - [ ] MysqlbinlogInput max_event_length
  - [ ] min.insync.replicas=2, shutdown 1 kafka broker then start
- [ ] GTID
- [ ] ugly globals register myslave_key

### Memo

- mysqlbinlog input peak with mock output
  - 140k event per second
  - 30k row event per second
  - 200Mb network bandwidth
  - it takes 2h25m to zero lag for platform of 2d lag

