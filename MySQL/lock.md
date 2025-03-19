# 一、MySQL中的锁
MySQL中的锁是用来管理多个事务对数据库资源（表、行···）的并发访问，避免数据的不一致冲突。

## 1.1 按粒度分类

### 1.1.1 全局锁
作用于整个数据库实例。在需要全库备份时，防止写入操作导致数据不一致。使用命令：
```sql
flush tables with read lock(FTWRL)
```
使得全库进入只读模式，阻塞写入和DDL操作，容易导致其他事务超时或死锁。
可以用 **set global read_only = 1**替代允许读操作的临时维护场景。但不能代替全局锁的严格一致性保证。

### 1.1.2 表级锁
作用于整张表。对整表操作，防止结构被其他线程修改。

**（1）表锁（Table Lock）**
直接锁定整张表，其他线程对表的访问受限
```sql
lock tables table_name READ WRITE
```

**(2)元数据锁（Metadata lock， MDL）**
在访问表结构或执行DDL时隐式加锁。触发条件：**alter table**或**查询**时。

### 1.1.3 行级锁
作用于单行记录。其锁粒度小，冲突少，性能高。开销较大，需要维护更多锁的状态。MySQL的**InnoDB存储引擎**实现。
**（1）共享锁（S锁/Read Lock）**
多个事务可以共享读取权限。
```sql
select ... lock in share mode
```

**（2）排他锁（X锁/Write Lock）**
阻止其他事务读取或写入。
```sql
select ... for update
```

## 1.2 按功能分类

### 1.2.1 乐观锁
- 特点：假设冲突很少，主要通过版本号或时间戳实现。
- 实现方式：
    （1）在表中加入版本号字段**version**
    （2）更新数据时使用条件：**update table_name set ... where id=? and version=?**
- 适用场景：并发不高，适合查询多、更新少的场景。

### 1.2.2 悲观锁
- 特点：假设冲突频繁，操作时直接加锁。
- 实现方式：借助事务和行锁完成
- 实现方式：
    （1）**select ... for update** 对读取的行加排他锁
    （2）**lock tables ... write** 对整表加写锁
- 适用场景：并发高、冲突概率大的场景。

## 1.3 死锁问题与处理

**（1）死锁原因**
两个或多个事务占有锁资源，互相等待对方释放锁。

**（2）处理方法**
- **等待超时机制**：设置参数 **innodb_lock_wait_timeout** ，超时自动回滚事务。
- **死锁检测**： InnoDB存储引擎支持自动检测回滚代价较大的事务。

## 1.4 InnoDB锁机制

**（1）记录锁（Record Lock）**
锁住索引记录本身，不会影响范围以外的数据。

**（2）间隙锁（Gap Lock）**
锁住索引范围，主要用于防止"幻读（Phantom Read）"。**repeatable read**隔离级别下的**select...for update**触发

**(3)临建锁（Next-key Lock）**
记录锁+间隙锁，防止插入新记录导致数据不一致。

## 1.5 锁相关的查看命令
* **show engine innodb status\g**: 查看死锁信息。  
* **show processlist**: 查看当前线程状态。  
* **information_schema.innodb_locks**: 查看锁信息。  
* **information_schema.innodb_lock_waits**：等待锁信息

## 1.6 总结
* **全局锁适用于全库只读。**
* **表级锁适用于批量操作。**
* **行级锁适用于事务并发，避免大范围冲突。**
* **乐观锁和悲观锁根据实际冲突情况选择。**

# 二、MySQL锁使用场景实例

## 2.1 全局锁：备份全库
使用全局锁将数据库置为只读模式，防止写入导致数据不一致。
```sql
-- 加锁：防止其他事务进行写操作
flush tables with read lock;

-- 实例：执行全库备份
mysqldump -uroot -p --all-databases > backup.sql

-- 解锁：恢复正常操作
unlock tables;
```

## 2.2 表级锁：批量更新数据
在大量更新时防止其他事务修改表。
```sql
-- 加锁：将表锁定为只读或写模式
lock tables orders write;

-- 示例：批量插入数据
insert into orders(order_id, customer_id, amount) values
(101, 1, 100.0),
(102, 2, 200.0);

-- 解锁：允许其他操作
unlock tables;
```

## 2.3 行级锁：共享锁（S锁）
多个事务可以读取同一行，但无法修改。
```sql
-- 事务1： 加共享锁
start transaction;
select * from products where product_id = 1 lock in share mode;

-- 事务2：尝试更新同一行（会被阻塞）
start transaction;
update products set stock = stock - 1 where product_id = 1; 
```

## 2.4 行级锁：排他锁（X锁）
一个事务加锁后，其他事务即不能读取也不能修改。
```sql
-- 事务1：加排他锁
start transaction;
select * from products where product_id = 1 for update;

-- 事务2：尝试读取同一行（会被阻塞）
start transaction;
select * from products where product_id = 1;
```

## 2.5 乐观锁：基于版本号控制并发
在更新检查版本号是否一致，避免冲突。
```sql
-- 表结构：增加一个`version`字段
create table products (
    product_id int primary key,
    stock int not null,
    version int not null
);

-- 事务1：读取并更新数据
start transaction;
select stock, version from products where product_id = 1;

-- 假设 version 为1，此时更新时验证版本号
update products
set stock = stock - 1, version = version + 1
where product_id = 1 and version = 1;

commit;
-- 如果版本号不匹配（被其他事务修改），更新将失败。
```

## 2.7 悲观锁：基于行锁的并发控制
直接加锁，确保操作期间数据不被其他事务修改。
```sql
-- 事务1：加悲观锁
start transaction;
select stock from products where product_id = 1 for update;

-- 示例：执行安全更新
update products set stock = stock - 1 where product_id = 1;

commit;
```

## 2.8 死锁场景及解决
**(1)示例：死锁发生**
```sql
-- 事务1：加锁顺序
start transaction;
select * from orders where order_id = 1 for update;
-- 此时事务2持有order_id = 2 的锁

-- 事务2：加锁顺序
start transaction;
select * from orders where order_id = 2 for update;
-- 此时事务1持有 order_id = 1 的锁

-- 事务1再尝试加锁 order_id = 2，事务2尝试加锁 order_id = 1，死锁产生。
```

**(2)解决方法**
- 超时机制：  
```sql
set innodb_lock_wait_timeout = 5; -- 等待锁超时时间
```
- 避免死锁：  
统一加锁顺序：确保所有事务以相同的顺序请求锁。

# 三、Go + MySQL锁 案例

**数据准备**  
```sql
create table products (
    pproduct_id int primary key,
    stock int not null,
    version int not null default 0
)
```  
  
**场景说明**  
* 库存扣减系统： 电商系统中需要扣减商品库存，涉及高并发操作。  
* 防止超卖： 使用 MySQL 的行锁（悲观锁）或乐观锁确保数据一致性。  
* 结合 Go 数据库操作： 使用 database/sql 包实现。  

```go
package main

import (
    "data/sql"
    "fmt"
    "log"
    _"github.com/go-sql-driver/mysql"  // 导入 MySQL 驱动
)

// 数据库配置
const (
    dsn = "root:cmx1014@tcp(127.0.0.1:3306)/ecommerce"
)

func main() {
    // 初始化数据库连接
    db, err := sql.Open("mysql", dsn)
    if err != nil {
        log.Fatalf("db connect failed： %v", err)
    }
    defer db.Close()

    // 使用悲观锁实现库存扣减
    err = deductStockWithPessimisticLock(db, 1, 10)
    if err != nil {
        log.Printf("deduct stock failed : %v", err)
    } else {
        log.Printf("deduct stock success!")
    }

    // 使用乐观锁实现库存扣减
    err = deductStockWithOptimisticLock(db, 1, 10)
    if err != nil {
        log.Printf("deduct stock failed : %v", err)
    } else {
        log.Printf("deduct stock success!")
    }
}

// 使用悲观锁扣减库存
func deductStockWithPessimisticLock(db *sql.DB, productID int, quantity int) err {
    tx, err := db.Begin()
    if err != nil {
        return fmt.Errorf("start transaction failed : %v", err)
    }
    defer func() {
        if err != nil {
            tx.Rollback()
        } else {
            tx.Commit()
        }
    }()

    // 查询库存并加行锁
    var stock int 
    query := "select stock from products where product_id = ? for update"
    err = tx.QueryRow(query, productID).Scan(&stock)
    if err != nil {
        return fmt.Errorf("query stock failed : %v", err)
    }

    if stock < quantity {
        return fmt.Errorf("stock not enough")
    }

    // 扣减库存
    update := "update products set stock = stock - ? where product_id = ?"
    _, err = tx.Exec(update, quantity, productID)
    if err != nil {
        return fmt.Errorf("deduct stock failed : %v", err)
    }

    return nil
}

// 使用乐观锁扣减库存
func deductStockWithOptimisticLock(db *sql.DB, productID int, quantity int) error {
    for {
        // 查询库存和版本号
        var stock, version int
        query := "select stock, version from products where product__id = ?"
        err := db.QueryRow(query,productID).Scan(&stock, &version)
        if err != nil {
            return fmt.Errorf("query stock failed: %v", err)
        }

        if stock < quantity {
            return fmt.Errorf("stock not enough")
        }

        // 更新库存时校验版本号
        update := "update products set stock = stock - ?, version = version + 1 where product_id = ? and version = ?"
        res, err := db.Exec(update, quantity, productID, version)
        if err != nil {
            return fmt.Errorf("deduct socket failed:%v", err)
        }

        // 检查是否有行被更新（受版本号影响）
        rowsAffected, err := res.RowsAffected()
        if err != nil {
            return fmt.Errorf("inspect updated result failed: %v", err)
        }
        if rowsAffected == 1 {
            return nil
        }

        // 若未更新，表示版本号冲突，重试
        log.Println("version number conflict, retry...")
    }
}
```