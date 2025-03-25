package main

import (
	"fmt"

	"github.com/jinzhu/gorm"
	_ "github.com/jinzhu/gorm/dialects/mysql"
)

type UserInfo struct {
	ID     uint
	Name   string
	Gender string
	Hobby  string
}

func main() {
	db, err := gorm.Open("mysql", "root:cmx1014@tcp(127.0.0.1:13306)/db1?charset=utf8mb4&parseTime=True&loc=Local")
	if err != nil {
		panic("conneted error :" + err.Error())
	}
	defer db.Close()

	db.AutoMigrate(&UserInfo{})

	u1 := UserInfo{1, "Amy", "male", "basketball"}
	u2 := UserInfo{2, "kitty", "female", "swim"}

	// create
	db.Create(&u1)
	db.Create(&u2)

	// query
	var u = new(UserInfo)
	db.First(u)
	fmt.Printf("%#v\n", u)

	var uu = new(UserInfo)
	db.Find(&uu, "hobby=?", "swim")
	fmt.Printf("%#v\n", uu)

	// update
	db.Model(&u).Update("hobby", "run")

	// delete
	db.Delete(&u)

}
