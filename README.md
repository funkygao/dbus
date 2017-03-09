# dbus
    
          $$\       $$\                                       
          $$ |      $$ |                                      
     $$$$$$$ |      $$$$$$$\        $$\   $$\        $$$$$$$\ 
    $$  __$$ |      $$  __$$\       $$ |  $$ |      $$  _____|
    $$ /  $$ |      $$ |  $$ |      $$ |  $$ |      \$$$$$$\  
    $$ |  $$ |      $$ |  $$ |      $$ |  $$ |       \____$$\ 
    \$$$$$$$ |      $$$$$$$  |      \$$$$$$  |      $$$$$$$  |
     \_______|      \_______/        \______/       \_______/ 
                                                              
yet another databus that transfer/transform pipeline data between plugins.

### Plugins

#### Input

- MysqlbinlogInput
- RedisbinlogInput
- KafkaInput
- MockInput

#### Filter

- MysqlbinlogFilter
- MockFilter

#### Output

- KafkaOutput
- ESOutput
- MockOutput

### Roadmap

- pubsub audit reporter
- universal kafka listener and outputer

### Dependencies

dbus itself has no external dependencies. But the plugins might have.
For example, MysqlbinlogInput uses zookeeper for sharding/election.

### Configuration

- KafkaOutput async mode with batch=1024/500ms, ack=WaitForAll
- Mysql binlog positioner commit every 1s, channal buffer 100

### TODO

- [X] pipeline
  - 1 input, multiple output
  - filter to dispatch dbs of a single binlog to different output
- [X] kill Packet.input field
- [X] router metrics
- [ ] visualized flow throughput like nifi
- [ ] router finding matcher is slow
- [X] dbusd api server
- [ ] pack.Payload reuse memory, json.NewEncoder(os.Stdout)
- [ ] ugly design of Input/Output ack mechanism
  - we might learn from storm bolt ack
- [ ] sharding binlog across the dbusd cluster
- [X] logging
- [X] share zkzone instance
- [X] presence and standby mode
- [X] graceful shutdown
- [X] master must drain before leave cluster
- [X] KafkaOutput metrics
  -  binlog tps
  -  kafka tps
  -  lag
- [X] hub is shared, what if a plugin blocks others
  - currently, I have no idea how to solve this issue
- [X] Batcher padding
- [X] shutdown kafka
- [X] zk checkpoint vs kafka checkpoint
- [X] kafka follower stops replication
- [ ] integration with helix
  - place config to central zk znode and watch changes
- [X] can a mysql instance with miltiple databases have multiple Log/Position?
- [X] kafka sync produce in batch
- [X] DDL binlog
  - drop table y;
- [X] trace async producer Successes channel and mark as processed
- [ ] each Input have its own recycle chan, one block will not block others
- [X] metrics
- [X] telemetry and alert
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
  - [X] router stat wrong
    Total:142,535,625      0.00B speed:22,671/s      0.00B/s max: 0.00B/0.00B
  - [ ] ffjson marshalled bytes has NL before the ending bracket
- [ ] test cases
  - [X] restart mysql master
  - [X] mysql kill process
  - [X] race detection
  - [ ] tc drop network packets and high latency
  - [X] reset binlog pos, and check kafka didn't recv dup events
  - [X] MysqlbinlogInput max_event_length
  - [ ] min.insync.replicas=2, shutdown 1 kafka broker then start
- [ ] GTID

### Memo

- mysqlbinlog input peak with mock output
  - 140k event per second
  - 30k row event per second
  - 200Mb network bandwidth
  - it takes 2h25m to zero lag for platform of 2d lag

