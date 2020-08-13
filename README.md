# Gorm Bulk Insert/Upsert

`Gorm Bulk Insert` is a library to implement bulk insert/upsert using [gorm](https://github.com/jinzhu/gorm).

## Purpose

Save bulk records

## Installation

`$ go get github.com/sunary/gorm-bulk-insert`

This library depends on gorm, following command is also necessary unless you've installed gorm.

`$ go get github.com/jinzhu/gorm`

## Usage

### BulkInsert

```go
bulk.BulkInsert(db, bulkData)
// or
bulk.BulkInsertWithTableName(db, tableName, bulkData)
```

### BulkUpsert

```go
bulk.BulkUpsert(db, bulkData, uniqueKeys)
// or
bulk.BulkUpsertWithTableName(db, tableName, bulkData, uniqueKeys)
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

	err = bulk.BulkInsert(db, bulkData)
	// or err = bulk.BulkInsertWithTableName(db, User{}.TableName(), bulkData)
	if err != nil {
		log.Fatal(err)
	}

	var bulkUpsertData []interface{}
	for i := 0; i < 100; i++ {
		bulkUpsertData = append(bulkUpsertData,
			User{
				UserName: "sunary",
				Age:      22,
				Hobby:    "soccer",
			},
		)
	}

	err = bulk.BulkUpsert(db, bulkUpsertData, []string{"name"})
	// or err = bulk.BulkUpsertWithTableName(db, User{}.TableName(), bulkData, []string{"name"})
	if err != nil {
		log.Fatal(err)
	}
}

```

## License

This project is under Apache 2.0 License