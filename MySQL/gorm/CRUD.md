# CRUD
CRUD 是数据库的增删改查操作，go 中可以使用GORM实现创建、查询、更新和删除操作。
首先我们先创建 承载 *gorm.DB 对象的 一个变量 db 。
```go
import (
    "github.com/jinzhu/gorm"
    _"github.com/jinzhu/gorm/dialects/mysql"
)
func main() {
    db, err := gorm.Open("mysql", "user:password@/dbname?charset=utf8&parseTime=true&loc=Local")
    defer db.Close()
}
```

## 1 Create

### 1.1 create record

```go
type User struct {
    ID int64
    Name string
    Age int64
}
// 使用 NewRecord() 查询主键是否存在，主键为空使用 Create() 创建记录：
user := User{Name: "q1mi", Age: 18}

db.NewRecord(user) // 主键为空返回`true`
db.Create(&user) // 创建user
db.NewRecord(user) // 创建`user`后返回`false`
```

### 1.2 default value
可以通过 tag 定义字段的默认值
```go
type User struct {
    ID int64
    Name string `gorm:"default:'小丸子'"`
    Age int64
}

var user = User{Name: "", Age: 99}
db.Create(&user)
```
上面实际执行的SQL语句是 INSERT INTO users("age") values('99');
排除了零值字段 Name， 
**Notice** ： 所有字段的零值， 比如 0， "" ， false 或者其他零值， 都不会保存到数据库内，但会使用它们的默认值。 若想避免这种情况，可以考虑 **指针** 或者实现 **Scanner/Valuer接口** 。

#### 1.2.1 Implementing zero value storage in the DB using pointer method
```go
type User struct {
    ID int64
    Name *string `gorm:"default:'小丸子'"`
    Age int64
}
user1 := User{Name: new(string), Age: 18}
user2 := User{Age: 19}
db.Create(&user1) // 此时数据库该条记录name字段的值就是''
db.Create(&user2) // 此时数据库该条记录name字段的值是默认值
```
#### 1.2.2 Implementing zero value storage in the database using Scanner/Valuer interface
```go
type User struct {
    ID int64
    Name sql.NullString `gorm:"default:'小丸子'"` // sql.NullString 实现了Scanner/Value 接口
    Age int64
}
user := User{Name: sql.NullString{"", true}, Age:18}
db.Create(&user) // 此时数据库该条记录name字段的值就是''
```

## 2 Query

### 2.1 General Query
```go
// 根据主键查询第一条记录
db.First(&user)

// 随机获取一条记录
db.Take(&user)

// 根据主键查询最后一条记录
db.Last(&user)

// 查询所有的记录
db.Find(&users)

// 查询指定的某条记录
db.First(&user, 10)
```
### 2.2 Where condition

#### 2.2.1 普通SQL查询
```go
// Get first matched record
db.Where("name = ?", "jinzhu").First(&user)

// Get all matched records
db.Where("name = ?", "jinzhu").First(&Users)

// <> exclude
db.Where("name <> ?", "jinzhu").Find(&users)

// IN
db.Where("name IN (?)", []string{"jinzhu","jiznhu1"}).Find(&users)

// LIKE 
db.Where("name Like ?", "%jin%").Find(&users)

// AND
db.Where("name = ? AND age >= ?", "jinzhu", "22").Find(&users)

// Time 
db.Where("updated_at > ?", lastWeek).Find(&users)

// BETWEEN
db.Where("created_at BETWEEN ? AND ?", lastweek, today).Find(&users)
```

#### 2.2.2 Struct & Map query
```go
// Struct
db.Where(&User{Name:"jinzhu", Age: 20}).First(&user)

// Map
db.Where(map[string]interface{}{"name": "jinzhu","age": 20}).Find(&users)

// Primary key's slices
db.Where([]int64{20, 21, 22}).Find(&users)
```
**Notice**: 当通过结构体进行查询时， GORM 将会只通过非零值字段查询。 这意味着若字段存在 零值， 将不会被用于构建查询条件。such：
```go
db.Where(&User{Name: "jinzhu", Age: 0}).Find(&users)
//// SELECT * FROM users WHERE name = "jinzhu";
```
但我们能通过使用指针或实现 Scanner/Valuer 接口来避免这个问题：
```go
// Pointer
type User struct {
  gorm.Model
  Name string
  Age  *int
}

// Scanner/Valuer
type User struct {
  gorm.Model
  Name string
  Age  sql.NullInt64  // sql.NullInt64 实现了 Scanner/Valuer 接口
}
```

### 2.3 Not Condition
作用与 Where 类似
```go
db.Not("name", "jinzhu").First(&user)
//// SELECT * FROM users WHERE name <> "jinzhu" LIMIT 1;

// Not In
db.Not("name", []string{"jinzhu", "jinzhu 2"}).Find(&users)
//// SELECT * FROM users WHERE name NOT IN ("jinzhu", "jinzhu 2");

// Not In slice of primary keys
db.Not([]int64{1,2,3}).First(&user)
//// SELECT * FROM users WHERE id NOT IN (1,2,3);

db.Not([]int64{}).First(&user)
//// SELECT * FROM users;

// Plain SQL
db.Not("name = ?", "jinzhu").First(&user)
//// SELECT * FROM users WHERE NOT(name = "jinzhu");

// Struct
db.Not(User{Name: "jinzhu"}).First(&user)
//// SELECT * FROM users WHERE name <> "jinzhu";
```

### 2.4 Or Condition
```go
db.Where("role = ?", "admin").Or("role = ?", "super_admin").Find(&users)
//// SELECT * FROM users WHERE role = 'admin' OR role = 'super_admin';

// Struct
db.Where("name = 'jinzhu'").Or(User{Name: "jinzhu 2"}).Find(&users)
//// SELECT * FROM users WHERE name = 'jinzhu' OR name = 'jinzhu 2';

// Map
db.Where("name = 'jinzhu'").Or(map[string]interface{}{"name": "jinzhu 2"}).Find(&users)
//// SELECT * FROM users WHERE name = 'jinzhu' OR name = 'jinzhu 2';
```
### 2.5 内联条件
作用与Where查询类似，当内联条件与多个立即执行方法一起使用时, 内联条件不会传递给后面的立即执行方法。
```go
// 根据主键获取记录 (只适用于整形主键)
db.First(&user, 23)
//// SELECT * FROM users WHERE id = 23 LIMIT 1;
// 根据主键获取记录, 如果它是一个非整形主键
db.First(&user, "id = ?", "string_primary_key")
//// SELECT * FROM users WHERE id = 'string_primary_key' LIMIT 1;

// Plain SQL
db.Find(&user, "name = ?", "jinzhu")
//// SELECT * FROM users WHERE name = "jinzhu";

db.Find(&users, "name <> ? AND age > ?", "jinzhu", 20)
//// SELECT * FROM users WHERE name <> "jinzhu" AND age > 20;

// Struct
db.Find(&users, User{Age: 20})
//// SELECT * FROM users WHERE age = 20;

// Map
db.Find(&users, map[string]interface{}{"age": 20})
//// SELECT * FROM users WHERE age = 20;
```

### 2.6 额外查询选项
```go
// 为查询 SQL 添加额外的 SQL 操作
db.Set("gorm:query_option", "FOR UPDATE").First(&user, 10)
//// SELECT * FROM users WHERE id = 10 FOR UPDATE;
```
**Set("gorm:query_option", "FOR UPDATE")：**悲观锁

### 2.7 FirstOrInit
**获取匹配的第一条记录，否则根据给定的条件初始化一个新的对象 (仅支持 struct 和 map 条件)***
```go
// 未找到
db.FirstOrInit(&user, User{Name: "non_existing"})
//// user -> User{Name: "non_existing"}

// 找到
db.Where(User{Name: "Jinzhu"}).FirstOrInit(&user)
//// user -> User{Id: 111, Name: "Jinzhu", Age: 20}
db.FirstOrInit(&user, map[string]interface{}{"name": "jinzhu"})
//// user -> User{Id: 111, Name: "Jinzhu", Age: 20}
```

#### 2.7.1 Attrs
**如果记录未找到，将使用参数初始化 struct.**
```go
// 未找到
db.Where(User{Name: "non_existing"}).Attrs(User{Age: 20}).FirstOrInit(&user)
//// SELECT * FROM USERS WHERE name = 'non_existing';
//// user -> User{Name: "non_existing", Age: 20}

db.Where(User{Name: "non_existing"}).Attrs("age", 20).FirstOrInit(&user)
//// SELECT * FROM USERS WHERE name = 'non_existing';
//// user -> User{Name: "non_existing", Age: 20}

// 找到
db.Where(User{Name: "Jinzhu"}).Attrs(User{Age: 30}).FirstOrInit(&user)
//// SELECT * FROM USERS WHERE name = jinzhu';
//// user -> User{Id: 111, Name: "Jinzhu", Age: 20}
```

#### 2.7.2 Assign
**不管记录是否找到，都将参数赋值给 struct.**
```go
// 未找到
db.Where(User{Name: "non_existing"}).Assign(User{Age: 20}).FirstOrInit(&user)
//// user -> User{Name: "non_existing", Age: 20}

// 找到
db.Where(User{Name: "Jinzhu"}).Assign(User{Age: 30}).FirstOrInit(&user)
//// SELECT * FROM USERS WHERE name = jinzhu';
//// user -> User{Id: 111, Name: "Jinzhu", Age: 30}
```

### 2.8 FirstOrCreate
**获取匹配的第一条记录, 否则根据给定的条件创建一个新的记录 (仅支持 struct 和 map 条件)**
```go
// 未找到
db.FirstOrCreate(&user, User{Name: "non_existing"})
//// INSERT INTO "users" (name) VALUES ("non_existing");
//// user -> User{Id: 112, Name: "non_existing"}

// 找到
db.Where(User{Name: "Jinzhu"}).FirstOrCreate(&user)
//// user -> User{Id: 111, Name: "Jinzhu"}
```

#### 2.8.1 Attrs
**如果记录未找到，将使用参数创建 struct 和记录.**
```go
 // 未找到
db.Where(User{Name: "non_existing"}).Attrs(User{Age: 20}).FirstOrCreate(&user)
//// SELECT * FROM users WHERE name = 'non_existing';
//// INSERT INTO "users" (name, age) VALUES ("non_existing", 20);
//// user -> User{Id: 112, Name: "non_existing", Age: 20}

// 找到
db.Where(User{Name: "jinzhu"}).Attrs(User{Age: 30}).FirstOrCreate(&user)
//// SELECT * FROM users WHERE name = 'jinzhu';
//// user -> User{Id: 111, Name: "jinzhu", Age: 20}
```

#### 2.8.2 Assign
**不管记录是否找到，都将参数赋值给 struct 并保存至数据库.**
```go
// 未找到
db.Where(User{Name: "non_existing"}).Assign(User{Age: 20}).FirstOrCreate(&user)
//// SELECT * FROM users WHERE name = 'non_existing';
//// INSERT INTO "users" (name, age) VALUES ("non_existing", 20);
//// user -> User{Id: 112, Name: "non_existing", Age: 20}

// 找到
db.Where(User{Name: "jinzhu"}).Assign(User{Age: 30}).FirstOrCreate(&user)
//// SELECT * FROM users WHERE name = 'jinzhu';
//// UPDATE users SET age=30 WHERE id = 111;
//// user -> User{Id: 111, Name: "jinzhu", Age: 30}
```

### 2.9 Advanced query

#### 2.9.1 SubQuery
```go
db.Where("amount > ?", 
    db.Table("orders")         // 子查询操作 orders 表
     .Select("AVG(amount)")    // 计算 amount 字段的平均值
     .Where("state = ?", "paid") // 过滤 state = 'paid' 的记录
     .SubQuery()               // 将整个查询转换为子查询表达式
).Find(&orders)
// SELECT * FROM "orders"  WHERE "orders"."deleted_at" IS NULL AND (amount > (SELECT AVG(amount) FROM "orders"  WHERE (state = 'paid')));
```

#### 2.9.2 Select Field
**Select，指定你想从数据库中检索出的字段，默认会选择全部字段。**
```go
db.Select("name, age").Find(&users)
//// SELECT name, age FROM users;

db.Select([]string{"name", "age"}).Find(&users)
//// SELECT name, age FROM users;

db.Table("users").Select("COALESCE(age,?)", 42).Rows()
//// SELECT COALESCE(age,'42') FROM users;
```

#### 2.9.3 Order
**Order，指定从数据库中检索出记录的顺序。设置第二个参数 reorder 为 true ，可以覆盖前面定义的排序条件。**
```go
db.Order("age desc, name").Find(&users)
//// SELECT * FROM users ORDER BY age desc, name;

// 多字段排序
db.Order("age desc").Order("name").Find(&users)
//// SELECT * FROM users ORDER BY age desc, name;

// 覆盖排序
db.Order("age desc").Find(&users1).Order("age", true).Find(&users2)
//// SELECT * FROM users ORDER BY age desc; (users1)
//// SELECT * FROM users ORDER BY age; (users2)
```

#### 2.9.4 Limit 数量
**Limit，指定从数据库检索出的最大记录数。**
```go
db.Limit(3).Find(&users)
//// SELECT * FROM users LIMIT 3;

// -1 取消 Limit 条件
db.Limit(10).Find(&users1).Limit(-1).Find(&users2)
//// SELECT * FROM users LIMIT 10; (users1)
//// SELECT * FROM users; (users2)
```

#### 2.9.5 Offset 偏移
**Offset，指定开始返回记录前要跳过的记录数。**
```go
db.Offset(3).Find(&users)
//// SELECT * FROM users OFFSET 3;

// -1 取消 Offset 条件
db.Offset(10).Find(&users1).Offset(-1).Find(&users2)
//// SELECT * FROM users OFFSET 10; (users1)
//// SELECT * FROM users; (users2)
```

#### 2.9.6 Count 总数
Count，该 model 能获取的记录总数。
**Notice Count 必须是链式查询的最后一个操作 ，因为它会覆盖前面的 SELECT，但如果里面使用了 count 时不会覆盖**
```go
db.Where("name = ?", "jinzhu").Or("name = ?", "jinzhu 2").Find(&users).Count(&count)
//// SELECT * from USERS WHERE name = 'jinzhu' OR name = 'jinzhu 2'; (users)
//// SELECT count(*) FROM users WHERE name = 'jinzhu' OR name = 'jinzhu 2'; (count)

db.Model(&User{}).Where("name = ?", "jinzhu").Count(&count)
//// SELECT count(*) FROM users WHERE name = 'jinzhu'; (count)

db.Table("deleted_users").Count(&count)
//// SELECT count(*) FROM deleted_users;

db.Table("deleted_users").Select("count(distinct(name))").Count(&count)
//// SELECT count( distinct(name) ) FROM deleted_users; (count)

```

#### 2.9.7 Group & Having
```go
rows, err := db.Table("orders").Select("date(created_at) as date, sum(amount) as toal").Group("date(created_at)").Rows()
for rows.Next() {
    ...
}

// 使用Scan将多条结果扫描进事先准备好的结构体切片中
type Result struct {
    Date time.time
    Total int
}
var rets []Result
db.Table("users").Select("date(created_at) as date, sum(age) as total").Group("date(created_at)").Scan(&rets)

rows, err := db.Table("orders").Select("date(created_at) as date, sum(amount) as total").Group("date(created_at)").Having("sum(amount) > ?", 100).Rows()
for rows.Next() {
  ...
}

type Result struct {
  Date  time.Time
  Total int64
}
db.Table("orders").Select("date(created_at) as date, sum(amount) as total").Group("date(created_at)").Having("sum(amount) > ?", 100).Scan(&results)
```

#### 2.9.8 Joins 连接
```go
rows, err := db.Table("users").Select("users.name, emails.email").Joins("left join emails on emails.user_id = users.id").Rows()
for rows.Next() {
  ...
}

db.Table("users").Select("users.name, emails.email").Joins("left join emails on emails.user_id = users.id").Scan(&results)

// 多连接及参数
db.Joins("JOIN emails ON emails.user_id = users.id AND emails.email = ?", "jinzhu@example.org").Joins("JOIN credit_cards ON credit_cards.user_id = users.id").Where("credit_cards.number = ?", "411111111111").Find(&user)

```
#### 2.9.9 Pluck
Pluck, 查询model中一个列作为切片。若想查询多个列，应该使用 Select + Find
```go
var ages []int64
db.Find(&users).Pluck("age", &ages)

var names []string
db.Model(&Users{}).Pluck("name", &names)

db.Table("deleted_users").Pluck("name", &names)

// 查询多个字段时
db.Select("name, age").Find(&users)
```

#### 2.9.10 Scan
Scan 的扫描结果最少是一个struct
```go
type Result struct {
  Name string
  Age int
}

var result Result
db.Table("users").Select("name, age").Where("name = ?", "Antonio").Scan(&result)

var result []Result
db.Table("users").Select("name, age").Where("id > ?", 0).Scan(&results)

// 原生 SQL
db.Raw("SELECT name, age FROM users WHERE name = ?", "Antonio").Scan(&result)
```

## 3 Method Chaining 
```go
// create a query
tx := db.Where("name = ?", "jinzhu")

// add more conditions
if someCondition {
  tx = tx.Where("age = ?", 20)
} else {
  tx = tx.Where("age = ?", 30)
}

if yetAnotherCondition {
  tx = tx.Where("active = ?", 1)
}
```
在调用立即执行方法前不会生成Query语句，借助这个特性可以创建一个函数来处理一些通用逻辑。

### 3.1 Immediate methods 
立即执行方法是指那些会立即生成SQL语句并发送到数据库的方法，一般是CRUD方法：
**Create、 First、 Find、 Take、 Save、 UpdateXXX、 Delete、 Scan、 Row、 Rows...**
基于上面的链式方法代码立刻执行方法的例子：
```go
tx.Find(&user)
// generated SQL
SELECT * FROM users Where name = 'jinzhu' AND age = 30 AND acive = 1;
```

### 3.2 Scopes 范围
Scope是建立在链式操作的基础之上的。，可以用它写出更多可重用的函数库
```go
func AmountGreaterThan1000(db *gorm.DB) *gorm.DB {
  return db.Where("amount > ?", 1000)
}

func PaidWithCreditCard(db * gorm.DB) * gorm.DB {
  return db.Where("pay_mode_sigh = ?", "C")
}

func PaidWithCod(db *gorm.DB) *gorm.DB {
  return db.Where("pay_mode_sign = ?", "c")
}

func OrderStatus(status []string) func (db *gorm.DB) *gorm.DB {
  return func (db *gorm.DB) *gorm.DB {
    return db.Scopes(AmountGreaterThan1000).Where("status IN (?)",status)
  }
}

db.Scopes(AmountGreaterThan1000, PaidWithCreditCard).Find(&orders)


db.Scopes(AmountGreaterThan1000, orderStatus([]string{"paid", "shipped"})).Find(&orders)
```

### 3.3 Multiple Immedia Methods
**Multiple Immediate Methods，在 GORM 中使用多个立即执行方法时，后一个立即执行方法会复用前一个立即执行方法的条件 (不包括内联条件) 。**
```go
db.Where("name LIKE ?", "jinzhu%").Find(&users, "id IN (?)", []int{1, 2, 3}).Count(&count)

// generate SQL
SELECT * FROM users WHERE name LIKE 'jinzhu%' AND id IN (1, 2, 3)

SELECT count(*) FROM users WHERE name LIKE 'jinzhu%'
```

## 4 Update

### 4.1 Update all Fields
Save()默认会更新该对象的所有字段，即使你没有赋值。
```go
db.First(&user)

user.Name = "七米"
user.Age = 99
db.Save(&user)
```

### 4.2 Update modified fields 
如果只更新指定字段，可以使用Update或者Updates
```go
db.Model(&user).Update("name", "hello")

db.Model(&user).Where("active = ?", true).Update("name", "hello")

db.Model(&user).Updates(map[string]interface{}{"name":"hello", "age":18, "active":false})

// 使用 struct 更新多个属性，只会更新其中有变化且为非零值的字段
db.Model(&user).Updates(User{Name: "hello", Age: 18})
// 对于下面的操作，不会发生任何更新，"", 0, false 都是其类型的零值
db.Model(&user).Updates(User{Name: "", Age: 0, Active: false})
```

### 4.3 Updated chosen fields
如果想更新或忽略某些字段，你可以使用 Select，Omit
```go
db.Model(&user).Select("name").Updates(map[string]interface{}{"name": "hello", "age": 18, "active": false})
//// UPDATE users SET name='hello', updated_at='2013-11-17 21:34:10' WHERE id=111;

db.Model(&user).Omit("name").Updates(map[string]interface{}{"name": "hello", "age": 18, "active": false})
//// UPDATE users SET age=18, active=false, updated_at='2013-11-17 21:34:10' WHERE id=111;
```

### 4.4 No hooks Update
上面的更新操作会自动运行 model 的 BeforeUpdate, AfterUpdate 方法，更新 UpdatedAt 时间戳, 在更新时保存其 Associations, 如果不想调用这些方法，可以使用 UpdateColumn， UpdateColumns
```go
// 更新单个属性，类似于 `Update`
db.Model(&user).UpdateColumn("name", "hello")
//// UPDATE users SET name='hello' WHERE id = 111;

// 更新多个属性，类似于 `Updates`
db.Model(&user).UpdateColumns(User{Name: "hello", Age: 18})
//// UPDATE users SET name='hello', age=18 WHERE id = 111;
```

### 4.5 Batch update
**批量更新时Hooks（钩子函数）不会运行.**
```go
db.Table("users").Where("id IN (?)", []int{10, 11}).Updates(map[string]interface{}{"name": "hello", "age": 18})
//// UPDATE users SET name='hello', age=18 WHERE id IN (10, 11);

// 使用 struct 更新时，只会更新非零值字段，若想更新所有字段，请使用map[string]interface{}
db.Model(User{}).Updates(User{Name: "hello", Age: 18})
//// UPDATE users SET name='hello', age=18;

// 使用 `RowsAffected` 获取更新记录总数
db.Model(User{}).Updates(User{Name: "hello", Age: 18}).RowsAffected
```

### 4.6 Use SQL expression to update
```go
var user User
db.First(&user)

db.Model(&user).Update("age", gorm.Expr("age * ? + ?", 2, 100))
//// UPDATE `users` SET `age` = age * 2 + 100, `updated_at` = '2020-02-16 13:10:20'  WHERE `users`.`id` = 1;

db.Model(&user).Updates(map[string]interface{}{"age": gorm.Expr("age * ? + ?", 2, 100)})
//// UPDATE "users" SET "age" = age * '2' + '100', "updated_at" = '2020-02-16 13:05:51' WHERE `users`.`id` = 1;

db.Model(&user).UpdateColumn("age", gorm.Expr("age - ?", 1))
//// UPDATE "users" SET "age" = age - 1 WHERE "id" = '1';

db.Model(&user).Where("age > 10").UpdateColumn("age", gorm.Expr("age - ?", 1))
//// UPDATE "users" SET "age" = age - 1 WHERE "id" = '1' AND quantity > 10;
```

### 4.7 Update hooks' value
修改 BeforeUpdate, BeforeSave 等 Hooks 中更新的值，可以用tx.Statement.SetColumn
```go
func (user *User) BeforeSave(tx *gorm.DB) error {
    if pw, err := bcrypt.GenerateFromPassword(user.Password, bcrypt.DefaultCost); err == nil {
        tx.Statement.SetColumn("EncryptedPassword", pw)
    }
    return nil
}
```

## 5 Delete

### 5.1 Delete record
**warm:**删除记录时，请确保主键字段有值，GORM 会通过主键去删除记录，如果主键为空，GORM 会删除该 model 的所有记录。
```go
// 删除现有记录
db.Delete(&email)
//// DELETE from emails where id=10;

// 为删除 SQL 添加额外的 SQL 操作
db.Set("gorm:delete_option", "OPTION (OPTIMIZE FOR UNKNOWN)").Delete(&email)
//// DELETE from emails where id=10 OPTION (OPTIMIZE FOR UNKNOWN);
```

### 5.2 Batch delete
删除全部匹配的记录.
```go
db.Where("email LIKE ?", "%jinzhu%").Delete(Email{})
//// DELETE from emails where email LIKE "%jinzhu%";

db.Delete(Email{}, "email LIKE ?", "%jinzhu%")
//// DELETE from emails where email LIKE "%jinzhu%";
```

### 5.3 Soft Delete
如果一个 model 有 DeletedAt 字段，他将自动获得软删除的功能！ 当调用 Delete 方法时， 记录不会真正的从数据库中被删除， 只会将DeletedAt 字段的值会被设置为当前时间。
```go
db.Delete(&user)
//// UPDATE users SET deleted_at="2013-10-29 10:23" WHERE id = 111;

// 批量删除
db.Where("age = ?", 20).Delete(&User{})
//// UPDATE users SET deleted_at="2013-10-29 10:23" WHERE age = 20;

// 查询记录时会忽略被软删除的记录
db.Where("age = 20").Find(&user)
//// SELECT * FROM users WHERE age = 20 AND deleted_at IS NULL;

// Unscoped 方法可以查询被软删除的记录
db.Unscoped().Where("age = 20").Find(&users)
//// SELECT * FROM users WHERE age = 20;
```

### 5.4 Psysical delete
```go
// Unscoped 方法可以物理删除记录
db.Unscoped().Delete(&order)
//// DELETE FROM orders WHERE id=10;
