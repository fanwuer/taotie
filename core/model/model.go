package model

import (
	"fmt"
	"taotie/core/util/rdb"
)

var Rdb *rdb.MyDb

func InitRdb(dbConfig rdb.MyDbConfig) error {
	if Rdb != nil {
		return nil
	}

	db, err := rdb.NewDb(dbConfig)
	if err != nil {
		return err
	}

	Rdb = db
	return nil
}

func CreateTable(tables []interface{}) {
	for _, table := range tables {
		ok, err := Rdb.IsTableExist(table)
		if err != nil {
			fmt.Println(err.Error())
			continue
		}
		if !ok {
			sess := Rdb.Client.NewSession()
			sess.Charset("utf8mb4")
			err = sess.CreateTable(table)
			if err != nil {
				sess.Close()
				fmt.Println(err.Error())
				continue
			}

			sess.Close()
		} else {
			sess := Rdb.Client.NewSession()
			err = sess.Sync2(table)
			if err != nil {
				sess.Close()
				fmt.Println(err.Error())
			}
			sess.Close()
		}

		err = Rdb.Client.CreateIndexes(table)
		if err != nil {
			fmt.Println(err.Error())
		}
		err = Rdb.Client.CreateUniques(table)
		if err != nil {
			fmt.Println(err.Error())
			continue
		}
	}

	u := new(User)
	u.Name = "admin"
	u.Email = "admin@admin"
	u.NickName = "admin"
	u.Password = "123456789"
	u.Status = 1
	u.Vip = 1
	u.Describe = "super man"
	u.InsertOne()
}
