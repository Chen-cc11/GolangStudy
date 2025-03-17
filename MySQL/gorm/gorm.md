# DEFINITION： GORM MODEL
GORM 为了方便模型定义，专门内置了一个 gorm.Model 结构体。
```go
type Model struct {
    ID unit `gorm:"primary_key"`
    CreateAt time.time
    UpdateAt time.time
    DeleteAt time.time
} 
```
我们可以把它嵌入到自己的模型中，也可以完全自定义模型。
## example : model definition
```go
type User struct {
  gorm.Model
  Name         string
  Age          sql.NullInt64
  Birthday     *time.Time
  Email        string  `gorm:"type:varchar(100);unique_index"`
  Role         string  `gorm:"size:255"` // 设置字段大小为255
  MemberNumber *string `gorm:"unique;not null"` // 设置会员号（member number）唯一并不为空
  Num          int     `gorm:"AUTO_INCREMENT"` // 设置 num 为自增类型
  Address      string  `gorm:"index:addr"` // 给address字段创建名为addr的索引
  IgnoreMe     int     `gorm:"-"` // 忽略本字段
}
```
## struct tags
使用结构体声明模型时，标记（tags）是可选的。gorm 支持 下面的标记：
#### single tags
Column	指定列名
Type	指定列数据类型
Size	指定列大小, 默认值255
PRIMARY_KEY	将列指定为主键
UNIQUE	将列指定为唯一
DEFAULT	指定列默认值
PRECISION	指定列精度
NOT NULL	将列指定为非 NULL
AUTO_INCREMENT	指定列是否为自增类型
INDEX	创建具有或不带名称的索引, 如果多个索引同名则创建复合索引
UNIQUE_INDEX	和 INDEX 类似，只不过创建的是唯一索引
EMBEDDED	将结构设置为嵌入
EMBEDDED_PREFIX	设置嵌入结构的前缀
-	忽略此字段

#### Associate related tags
MANY2MANY	指定连接表
FOREIGNKEY	设置外键
ASSOCIATION_FOREIGNKEY	设置关联外键
POLYMORPHIC	指定多态类型
POLYMORPHIC_VALUE	指定多态值
JOINTABLE_FOREIGNKEY	指定连接表的外键
ASSOCIATION_JOINTABLE_FOREIGNKEY	指定连接表的关联外键
SAVE_ASSOCIATIONS	是否自动完成 save 的相关操作
ASSOCIATION_AUTOUPDATE	是否自动完成 update 的相关操作
ASSOCIATION_AUTOCREATE	是否自动完成 create 的相关操作
ASSOCIATION_SAVE_REFERENCE	是否自动完成引用的 save 的相关操作
PRELOAD	是否自动完成预加载的相关操作

# Convention ： primary key、table name、 column name
## Primary key
GORM 默认会使用名为ID的字段作为表的主键。
type User struct {
    ID string // default : primary
    Name string
}

## Table name
表名默认是就是结构体名称的复数。
```go
type User struct {} // default : table name users

// set the table name of User to 'profiles'
func (User) TableName() string {
  return "profiles"
}

func (u User) TableName() string {
  if u.Role == "admin" {
    return "admin_users"
  } else {
    return "users"
  }
}

// 禁用默认表名的复数形式，如果置为 true，则 `User` 的默认表名是 `user`
db.SingularTable(true)

GORM 还支持更改默认表名称规则：
gorm.DefaultTableNameHandler = func (db *gorm.DB , defaultTableName string) stirng {
  return "prefix_" + defaultTableName;
}
```

## Column name

```go
type User struct {
  ID        uint      // column name is `id`
  Name      string    // column name is `name`
  Birthday  time.Time // column name is `birthday`
  CreatedAt time.Time // column name is `created_at`
}
```
还可以通过结构体tag指定列名：
```go
type Animal struct {
  AnimalId    int64     `gorm:"column:beast_id"`         // set column name to `beast_id`
  Birthday    time.Time `gorm:"column:day_of_the_beast"` // set column name to `day_of_the_beast`
  Age         int64     `gorm:"column:age_of_the_beast"` // set column name to `age_of_the_beast`
}
```

## Timestamp tracking
### CreateAt
```go
db.Create(&user) // `CreatedAt`将会是当前时间
// 可以使用`Update`方法来改变`CreateAt`的值
db.Model(&user).Update("CreatedAt", time.Now())
```
### UpdateAt
```go
db.Save(&user) // `UpdatedAt`将会是当前时间
db.Model(&user).Update("name", "jinzhu") // `UpdatedAt`将会是当前时间
```
### DeleteAt
如果模型有DeletedAt字段，调用Delete删除该记录时，将会设置DeletedAt字段为当前时间，而不是直接将记录从数据库中删除。