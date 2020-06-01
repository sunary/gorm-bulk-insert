# Gorm Bulk Insert

`Gorm Bulk Insert` is a library to implement bulk insert using [gorm](https://github.com/jinzhu/gorm). Execute bulk insert just by passing a slice of struct, as if you were using a gorm regularly.

## Purpose

When saving a large number of records in database, inserting at once - instead of inserting one by one - leads to significant performance improvement. This is widely known as bulk insert.

Gorm is one of the most popular ORM and contains very developer-friendly features, but bulk insert is not provided.

This library is aimed to solve the bulk insert problem.

## Installation

`$ go get github.com/sunary/gorm-bulk-insert`

This library depends on gorm, following command is also necessary unless you've installed gorm.

`$ go get github.com/jinzhu/gorm`

## Usage

```go
bulk.Insert(db, tableName, bulkData)
```

## Example

```go
package main

import (
	"log"
	"time"

	"github.com/jinzhu/gorm"
	_ "github.com/jinzhu/gorm/dialects/mysql"
	bulk "github.com/sunary/gorm-bulk-insert"
)

type User struct {
	ID        int
	UserName  string `gorm:"column:name"`
	Age       int
	Hobby     string `gorm:"-"`
	CreatedAt time.Time
	UpdatedAt time.Time
}

func (u User) TableName() string {
	return "user"
}

func main() {
	db, err := gorm.Open("mysql", "root:password@tcp(localhost:3306)/db_name")
	if err != nil {
		log.Fatal(err)
	}

	var bulkData []interface{}
	for i := 0; i < 10000; i++ {
		bulkData = append(bulkData,
			User{
				UserName: "sunary",
				Age:      22,
				Hobby:    "dance",
			},
		)
	}

	err = bulk.Insert(db, User{}.TableName(), bulkData)
	if err != nil {
		log.Fatal(err)
	}
}

```

## License

This project is under Apache 2.0 License