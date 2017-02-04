com.alibaba.otter.canal.protocol.Event

Event[
    logIdentity=LogIdentity[sourceAddress=/192.168.47.128:3306,slaveId=-1],
    entry=header {
             version: 1
             logfileName: "mysql-bin.000061"
             logfileOffset: 247
             serverId: 1
             serverenCode: "UTF-8"
             executeTime: 1469589006000
             sourceType: MYSQL
             schemaName: "test1"
             tableName: "test1"
             eventLength: 44
             eventType: INSERT
    }
    entryType: ROWDATA
    storeValue: "\bG\020\001P\000bC\022\033\b\000\020\001\032\002id \000(\0010\000B\0011R\bchar(30)\022$\b\001\020\f\032\004name \000(\0010\000B\005test1R\vvarchar(20)"
]

storeValue 数据反序列化：
    tableId: 71
    eventType: INSERT
    isDdl: false
    rowDatas {
        afterColumns {
               index: 0
               sqlType: 1
               name: "id"
               isKey: false
               updated: true
               isNull: false
               value: "1"
               mysqlType: "char(30)"
        }
        afterColumns {
               index: 1
               sqlType: 12
               name: "name"
               isKey: false
               updated: true
               isNull: false
               value: "test1"
               mysqlType: "varchar(20)"
        }
    }
