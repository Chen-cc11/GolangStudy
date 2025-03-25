# Golang 实现 MySQL 数据库事务

## 一、MySQL事务
MySQL事务是指一组数据库操作，它们被视为逻辑单元，并且它们要么全部成功执行，要么全部回滚（撤销）。
事务是数据库管理系统提供的一种机制，用于确保数据的一致性和完整性。

**事务具有以下原则（ACID）：**
    1. 原子性（Atomicity）：事务中的所有操作要么全部成功执行，要么全部回滚，不存在部分执行的情况。
    2. 一致性（Consistency）：事务的执行会使数据库从一个一致的状态转换到另一个一致的状态。这意味着在事务开始和结束时，数据必须满足预定义的完整性约束。
    3. 隔离性（Isolation）：事务的执行是相互隔离的，即一个事务的操作在提交之前对其他事务是不可见的。并发事务之间的相互影响被隔离，以避免数据损坏和不一致的结果。

    级别	            脏读	不可重复读	   幻读	  性能
    Read Uncommitted	✔️	     ✔️	         ✔️	  最高
    Read Committed	    ✖️	     ✔️	         ✔️	   高
    Repeatable Read	    ✖️	     ✖️	         ✔️	   中
    Serializable	    ✖️	     ✖️	         ✖️	   最低 

    脏读：     单条数据未提交值    修改+未提交       读取其他事务未提交的临时数据
    不可重复读：单条数据版本变化    修改+已提交       同一数据在十五内多次读取结果不同
    幻读：     范围数据行数变化    插入/删除+已提交   同一条件查询结果行数前后不同

    4. 持久性（Durability）：一旦事务提交成功，其对数据库的更改将永久保存，即使在系统故障或重启之后也能保持数据的持久性。

```sql
-- 在MySQL中，开始一个事务的语句
start transation;

-- 在事务中，可以执行一系列的数据库操作，如插入、更新和删除等。最后提交事务或回滚事务：

-- 提交事务
commit;

-- 回滚事务
rollback;

-- 通过使用事务，可以确保数据库操作的一致性和完整性，尤其在处理涉及多个相关操作的复杂业务逻辑时非常有用。
```

## 二、MySQL事务示例
```sql
start transaction;

-- 在事务中执行一系列数据库操作
insert into users(name, age) values('Alice', 25); 
update accounts set balance = balance - 100 where user_id = 1;
delete from logs where user_id= 1;

-- 如果一切正常，提交事务
commit;
```

如果在事务执行的过程中出现任何错误或异常情况，可以使用ROLLBACK语句回滚事务，使所有操作都被撤销，数据库恢复到事务开始之前的状态
```sql
业务规则检查（例如余额不能为负）
IF (SELECT balance FROM accounts WHERE user_id = 2) < 0 THEN
  ROLLBACK;  -- 主动回滚
ELSE
  COMMIT;    -- 提交
END IF;
```
事务的关键在于将多个相关的数据库操作组织在一起，并以原子性和一致性的方式进行提交或回滚。这确保了数据的完整性和一致性，同时也提供了灵活性和错误恢复机制。

## 三、MySQL事务引擎
* **1.InnoDB**： MySQL默认的事务引擎。它支持事务、行级锁定、外键约束和崩溃恢复等功能。适用于需要强调数据完整性和并发性能的应用程序。
* **2.MyISAM**： MySQL另一个常见的事务引擎。它不支持事务和行级锁定，但具有较高的插入和查询性能。适用于读密集型应用程序，如日志记录和全文搜索。
* **3.NDB CluSter**：MySQL的集群事务引擎，适用于需要高可用性和可扩展性的分布式应用程序。它具有自动分片、数据冗余和故障恢复等功能。
* **4.Memory**： （也称为Heap引擎）将表数据存储在内存中，提供非常高的插入和查询性能。但由于数据存储在内存中，因此在数据库重新启动时数据会丢失。Memory引擎是用于临时数据或缓存数据的存储。

在创建表时，可以指定所需的事务引擎。不同的事务引擎可能会有不同的配置和限制。
```sql
create table mytable (
    id int primary key,
    name varchar(20)
) engine = innoDB;
```

## 四、事务实例

### 1.开启事务Begin源码
```go
func (db *DB) BeginTx(ctx context.Context, opts *TxOptions) (*Tx, error) {
    var tx *Tx
    var err error
    // 通过重试机制获取连接
    err = db.retry(func(strategy connReuseStrategy) error {
        tx, err = db.begin(ctx,opts,strategy) // 核心事务初始化
        return err
    })
    return tx, err
}

func (db *DB) Begin() (*Tx, error) {
    return db.BeginTx(context.Background(), nil) // 默认调用BeginTx
}
```
* **（1）ctx context.Context**
- 绑定事务生命周期到上下文，支持超时、取消和链路跟踪。me
- 超时控制： 设置事务最长执行时间 context.WithTimeout
- 主动取消： 当用户中断请求时，通过ctx.Cancel()回滚事务
- 分布式追踪： 传递请求链路标识 OpenTelemetry

* **(2) opts *TxOptions**
- 配置事务的隔离级别和只读模式
- Isolation： 隔离级别（eg LevelReadCommited、LevelSerializable）
- ReadOnly
```go
opts := &sql.TxOptions{
    Isolation: sql.LevelSerializable,
    ReadOnly : false,
}
tx, err := db.BeginTx(ctx, opts)
```

* **(3) db.begin**
- 获取连接 ：调用db.conn(ctx, strategy)从连接池获取可用连接（driverConn）
- 初始化事务：通过db.benginDc将连接绑定到事务对象sql.Tx
- 驱动初始化：调用 ctxDriverBegin(ctx, opts, dc.ci)，最终由 MySQL 驱动发送 START TRANSACTION 到数据库

* **（4） sql.Tx结构体**
- dc *driverConn ： 事务绑定的物理数据库连接
- releaseConn func(error): 释放连接回连接池的回调函数
- txi driver.Tx : 底层驱动的事务接口

### 2.中止事务Rollback源码
```go
func (tx *Tx) rollback(discardConn bool) error {
    // 1.方法签名与原子状态检查
    // 原子变量tx.done 标记事务是否已完成
    if !tx.done.CompareAndSwap(false, true) {
        return ErrTxDone
    }

    // 2.钩子函数 跟踪事务回滚事件
    if rollbackHook != nil {
        rollbackHook()
    }

    // 3.上下文取消和锁管理
    tx.cancel()
    // 取消与事务关联的 context.Context，释放所有阻塞的查询（如正在执行的 Query 或 Exec），避免死锁。
    // closemu 读写锁sync.RWMutex 确保回滚期间无其它操作干扰
    tx.closemu.Lock()
    // 加写锁：tx.closemu.Lock() 防止其他协程修改连接状态。
    tx.closemu.Unlock()
    // 锁的作用是确保 tx.cancel() 的同步。

    // 4.驱动层回滚执行
    var err error
    withLock(tx.dc, func() {
        err = tx.txi.Rollback()
        // 调用具体数据库驱动（如 MySQL、PostgreSQL）实现的回滚方法，发送 ROLLBACK 命令到数据库
    })
    
    // 5.预处理语句关闭与错误处理
    if !errors.Is(err, driver.ErrBadConn) {
        tx.closePrepared()
    }
    // 5.连接丢弃逻辑
    if discardConn {
        err = driver.ErrBadConn
    }
    tx.close(err)
    return err
}
// 公共 Rollback 方法
func (tx *TX) Rollback() error {
    return tx.rollback(false) //表示正常回滚，不强制丢弃连接。
}
// 应优先调用 Rollback()，仅在驱动内部处理严重错误时使用 discardConn=true。
```

### 3.提交事务Commit源码
```go
func (tx *TX) Commit() error {
    // 1.上下文状态检查
    select {
    // 当 tx.ctx.Done() 通道有数据时，执行 case <-tx.ctx.Done()；否则立即执行 default（无操作）
    default :
    // 监听上下文tx.ctx的完成信号通道， 若已关闭（上下文失效），执行该case
    case <-tx.ctx.Done():
        if tx.doneLoad() {
            return ErrTxDone
        }
        return tx.ctx.Err()
    }shiwu
    // 2.原子状态标记
    if  !tx.done.CompareAndSwap(false, true) {
        // 原子操作确保事务仅能提交一次
        return ErrTxDone
    }
    // 3.上下文取消与锁管理
    tx.cancel()
    tx.closemu.Lock()
    tx.closemu.UnLock()

    // 4.驱动层提交执行
    var err error
    withLock(tx.dc, func() {
        err = tx.txi.Commit()
    })
    // 5.预处理语句关闭与错误处理
    if !errors.Is(err, driver.ErrBadConn) {
        tx.closePrepared()
    }
    tx.close(err)
    return err
}
```
### 4.example
```go
package main

import (
    "database/sql"
    "fmt"
    "time"
    _"github.com/go-sql-driver/mysql" // 匿名导入 自动执行 init()
)

var db *sql.DB

func initMySQL() (err error) {
    // DSN(data source name)
    dsn := "root:cmx1014@tcp(127.0.0.1:3306)/sql_test"
    db, err = sql.Open("mysql", dsn) // 只对格式进行校验，并不会真正连接数据库
    if err != nil {
        return err
    }
    // 数值需要根据业务具体情况来确定
    db.SetConnMaxLifetime(time.Second * 10) // 设置可以重用连接的最长时间
    db.SetConnMaxIdleTime(time.Second * 5)  // 设置连接可能处于空闲状态的最长时间
    db.SetMaxOpenConns(20)                  // 设置与数据库的最大打开连接数
    db.SetMaxIdleConns(10)                  // 设置空闲连接池中的最大连接数
    return nil
}

type user struct {
    id int
    age int
    name string
}

// 事务操作
func transactionDemo() {
    // 启动事务。默认隔离级别取决于驱动程序。
    tx, err := db.Begin() // 开启事务
    if err != nil {
        if tx != nil {
            tx.Rollback() // 回滚 中止事务
        }
        fmt.Printf("begin trans failed, err : %v\n", err)
        return
    }
    sqlStr1 := "update user set age= ? where id= ?"
    ret1, err := tx.Exec(sqlStr1, 21, 1)
    if err != nil {
        tx.Rollback()
        fmt.Printf("exec sql1 failed, err : %v\n", err)
        return
    }
    // RowsAffected 返回受更新、插入或删除影响的行数。
    affRow1, err := ret1.RowsAffected()
    if err != nil {
        tx.Rollback()
        fmt.Printf("exec ret1.RowsAffected() failed, err : %v\n", err)
        return
    }

    sqlStr2 := "update user set age=? where id=?"
    ret2, err := tx.Exec(sqlStr2, 100, 5)
    if err != nil {
        tx.Rollback()
        fmt.Printf("exec sql2 failed, err : %v\n", err)
        return
    }
    affRow2, err := ret2.RowsAffected()
    if err != nil {
        tx.Rollback()
        fmt.Printf("exec ret2.RoesAffected() failed, err : %v\n", err)
        return
    }

    fmt.Println(affRow1, affRow2)
    if affRow1 == 1 && affRow2 == 1 {
        fmt.Println("事务提交.....")
        tx.Commit()
    } else {
        tx.RollBack()
        fmt.Println("事务回滚")
    }

    fmt.Println("exec trans successfully!")
}

func main() {
    if err := initMySQL(); err != nil {
        fmt.Printf("connect to db failed, err : %v\n", err)
    }
    defer db.Close()

    fmt.Println("connect to db successfully!")

    transactionDemo()
}
```
**运行结果**
```sql
mysql> desc user;
+-------+-------------+------+-----+---------+----------------+
| Field | Type        | Null | Key | Default | Extra          |
+-------+-------------+------+-----+---------+----------------+
| id    | int         | NO   | PRI | NULL    | auto_increment |
| name  | varchar(50) | NO   |     | NULL    |                |
| age   | int         | YES  |     | NULL    |                |
+-------+-------------+------+-----+---------+----------------+
3 rows in set (0.04 sec)

mysql> insert into user (id, name, age) values (1,"Kitty", 12), (5, "Alice",23);
Query OK, 2 rows affected (0.02 sec) Records: 2  Duplicates: 0  Warnings: 0

mysql> select * from user;
+----+-------+------+
| id | name  | age  |
+----+-------+------+
|  1 | Kitty |   21 |
|  5 | Alice |  100 |
+----+-------+------+
2 rows in set (0.00 sec)
```