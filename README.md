# dbus
    
          $$\       $$\                                       
          $$ |      $$ |                                      
     $$$$$$$ |      $$$$$$$\        $$\   $$\        $$$$$$$\ 
    $$  __$$ |      $$  __$$\       $$ |  $$ |      $$  _____|
    $$ /  $$ |      $$ |  $$ |      $$ |  $$ |      \$$$$$$\  
    $$ |  $$ |      $$ |  $$ |      $$ |  $$ |       \____$$\ 
    \$$$$$$$ |      $$$$$$$  |      \$$$$$$  |      $$$$$$$  |
     \_______|      \_______/        \______/       \_______/ 

### What is dbus?

dbus = distributed data bus

It is yet another lightweight versatile databus system that transfer/transform pipeline data between plugins.

dbus works by building a DAG of structured data out of the different plugins: from data input, via filter(optional), to the output.

Similar projects

- logstash
- flume
- nifi
- camel
- beats
- kettle
- zapier
- google cloud dataflow
- canal
- storm
- yahoo pipes (dead)

### Status

dbus is not yet a 1.0.
We're writing more tests, fixing bugs, working on TODOs.

### Use Case

- mysql binlog dispatcher
- multiple DC kafka mirror

### Features

dbus supports powerful and scalable directed graphs of data routing, transformation and
system mediation logic.

- Designed for extension
  - plugin architecture
  - build your own plugins and more
  - enables rapid development and effective testing
- Data Provenance
  - track dataflow from beginning to end
  - visualized dataflow
  - rich metrics feed into tsdb
  - online manual mediation of the dataflow
  - RESTful API
  - monitoring with alert
- Distributed Deployment
  - shard/balance/auto rebalance
  - linear scale
- Delivery Guarantee
  - loss tolerant
  - high throuput vs low latency
  - back pressure
- Robustness
  - race condition detected
  - edge cases fully covered
  - network jitter tested
  - dependent components failure tested
- Systemic Quality
  - hot reload
  - dryrun throughput 1.9M packets/s
- Cluster Support
  - modelling borrowed from helix+kafka controller
  - currently only leader/standby with sharding, without replica
  - easy to write a distributed plugin

### Getting Started

#### 1. Installing

To start using dbus, install Go and run `go get`:

```sh
$ go get -u github.com/funkygao/dbus
```

#### 2. Create config file

Please find sample config files in etc/ directory.

#### 3. Run the server

```sh
$ $GOPATH/dbusd -conf $myfile
```

### Dependencies

dbus uses zookeeper for sharding/balance/election.

### Plugins

More plugins are listed under [dbus-plugin](https://github.com/dbus-plugin).

#### Input

- MysqlbinlogInput
- KafkaInput
- MockInput
- StreamInput
  
#### Filter

- MysqlbinlogFilter
- MockFilter

#### Output

- KafkaOutput
- ESOutput
- MockOutput
- StreamOutput

### Configuration

- KafkaOutput async mode with batch=1024/500ms, ack=WaitForAll
- KafkaOutput retry=3, retry.backoff=350ms
- Mysql binlog positioner commit every 1s, channal buffer 100

### FAQ

#### mysql binlog

- is it totally data loss tolerant?
  
  if the binlog exceeds 1MB, it will be discarded(lost)

#### why not canal?

- no Delivery Guarantee
- no Data Provenance
- no integration with kafka
- only hot standby deployment mode, we need sharding load
- dbus is a dataflow engine, while canal only support mysql binlog pipeline

#### compared with logstash

- logstash has better ecosystem
- dbus is cluster aware, provides delivery guarantee, data provenance

#### can there be more than 1 leaders at the same time?

Yes.

For example, 3 participants with 1 being the leader. Then 1 is network partitioned and
zk session expires, [2, 3] found this event and re-elect 2 as new leader.
Before 1 regain new zk session, [1] and [2] are leaders both.
If [1] and [2] both found resources changes, they will both rebalance the cluster.

dbus uses epoch to solve this issue.

#### what if

- zookeeper crash

  dbus continues to work, but Ack will not be able to persist

### Memo

- mysqlbinlog input peak with mock output
  - 140k event per second
  - 30k row event per second
  - 260Mb network bandwidth
  - KafkaOutput 35K msg per second
  - it takes 2h25m to zero lag for platform of 2d lag

- dryrun MockInput -> MockOutput
  - 2.1M packet/s

