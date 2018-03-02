// Copyright (C) 2017, Zipper Team.  All rights reserved.
//
// This file is part of zipper
//
// The zipper is free software: you can redistribute it and/or modify
// it under the terms of the GNU General Public License as published by
// the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version.
//
// The zipper is distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
//
// GNU General Public License for more details.
//
// You should have received a copy of the GNU General Public License
// along with this program.  If not, see <http://www.gnu.org/licenses/>.

package db

import (
	"database/sql"
	"fmt"

	"gopkg.in/mgo.v2"
	db "upper.io/db.v3"
	"upper.io/db.v3/mongo"
	"upper.io/db.v3/mssql"
	"upper.io/db.v3/mysql"
	"upper.io/db.v3/postgresql"
	"upper.io/db.v3/ql"
	"upper.io/db.v3/sqlite"
)

var host = "localhost"

var settings = map[string]db.ConnectionURL{
	mongo.Adapter: &mongo.ConnectionURL{
		Database: `upperio_tests`,
		Host:     host,
		User:     `upperio_tests`,
		Password: `upperio_secret`,
	},
	postgresql.Adapter: &postgresql.ConnectionURL{
		Database: `upperio_tests`,
		Host:     host,
		User:     `upperio_tests`,
		Password: `upperio_secret`,
		Options: map[string]string{
			"timezone": "UTC",
		},
	},
	sqlite.Adapter: &sqlite.ConnectionURL{
		Database: `sqlite3-test.db`,
	},
	mysql.Adapter: &mysql.ConnectionURL{
		Database: `upperio_tests`,
		Host:     host,
		User:     `upperio_tests`,
		Password: `upperio_secret`,
		Options: map[string]string{
			"parseTime": "true",
		},
	},
	mssql.Adapter: &mssql.ConnectionURL{
		Database: `upperio_tests`,
		Host:     host,
		User:     `upperio_tests`,
		Password: `upperio_Secre3t`,
	},
	ql.Adapter: &ql.ConnectionURL{
		Database: `ql-test.db`,
	},
}

var setupFn = map[string]func(driver interface{}) error{
	mongo.Adapter: func(driver interface{}) error {
		if mgod, ok := driver.(*mgo.Session); ok {
			var col *mgo.Collection
			col = mgod.DB("upperio_tests").C("birthdays")
			col.DropCollection()

			return nil
		}
		return fmt.Errorf("Driver error: Expecting *mgo.Session got %T (%#v)", driver, driver)
	},
	postgresql.Adapter: func(driver interface{}) error {
		if sqld, ok := driver.(*sql.DB); ok {
			var err error
			_, err = sqld.Exec(`DROP TABLE IF EXISTS "birthdays"`)
			if err != nil {
				return err
			}
			_, err = sqld.Exec(`CREATE TABLE "birthdays" (
						"id" serial primary key,
						"name" CHARACTER VARYING(50),
						"born" TIMESTAMP WITH TIME ZONE,
						"born_ut" INT
				)`)
			if err != nil {
				return err
			}

			return nil
		}
		return fmt.Errorf("Driver error: Expecting *sql.DB got %T (%#v)", driver, driver)
	},
	sqlite.Adapter: func(driver interface{}) error {
		if sqld, ok := driver.(*sql.DB); ok {
			var err error
			_, err = sqld.Exec(`DROP TABLE IF EXISTS "birthdays"`)
			if err != nil {
				return err
			}
			_, err = sqld.Exec(`CREATE TABLE "birthdays" (
					"id" INTEGER PRIMARY KEY,
					"name" VARCHAR(50) DEFAULT NULL,
					"born" DATETIME DEFAULT NULL,
					"born_ut" INTEGER
				)`)
			if err != nil {
				return err
			}

			return nil
		}
		return fmt.Errorf("Driver error: Expecting *sql.DB got %T (%#v)", driver, driver)
	},
	mysql.Adapter: func(driver interface{}) error {
		if sqld, ok := driver.(*sql.DB); ok {
			var err error
			_, err = sqld.Exec(`DROP TABLE IF EXISTS ` + "`" + `birthdays` + "`" + ``)
			if err != nil {
				return err
			}
			_, err = sqld.Exec(`CREATE TABLE ` + "`" + `birthdays` + "`" + ` (
					id BIGINT(20) UNSIGNED NOT NULL AUTO_INCREMENT, PRIMARY KEY(id),
					name VARCHAR(50),
					born DATE,
					born_ut BIGINT(20) SIGNED
				) CHARSET=utf8`)
			if err != nil {
				return err
			}

			return nil
		}
		return fmt.Errorf("Driver error: Expecting *sql.DB got %T (%#v)", driver, driver)
	},
	mssql.Adapter: func(driver interface{}) error {
		if sqld, ok := driver.(*sql.DB); ok {
			var err error
			_, err = sqld.Exec(`DROP TABLE IF EXISTS [birthdays]`)
			if err != nil {
				return err
			}
			_, err = sqld.Exec(`CREATE TABLE [birthdays] (
					id BIGINT PRIMARY KEY NOT NULL IDENTITY(1,1),
					name NVARCHAR(50),
					born DATETIME,
					born_ut BIGINT
				)`)
			if err != nil {
				return err
			}

			return nil
		}
		return fmt.Errorf("Driver error: Expecting *sql.DB got %T (%#v)", driver, driver)
	},
	ql.Adapter: func(driver interface{}) error {
		if sqld, ok := driver.(*sql.DB); ok {
			var err error
			var tx *sql.Tx
			if tx, err = sqld.Begin(); err != nil {
				return err
			}
			_, err = tx.Exec(`DROP TABLE IF EXISTS birthdays`)
			if err != nil {
				return err

			}
			_, err = tx.Exec(`CREATE TABLE birthdays (
					name string,
					born time,
					born_ut int
				)`)
			if err != nil {
				return err
			}
			if err = tx.Commit(); err != nil {
				return err
			}

			return nil
		}
		return fmt.Errorf("Driver error: Expecting *sql.DB got %T (%#v)", driver, driver)
	},
}

func Open(wrapper string) (db.Database, error) {
	var err error
	var sess db.Database
	if settings[wrapper] == nil {
		err = fmt.Errorf(`No such settings entry for wrapper %s`, wrapper)
	} else {
		sess, err = db.Open(wrapper, settings[wrapper])
	}
	return sess, err
}

func Setup(wrapper string, sess db.Database) error {
	var err error
	if setupFn[wrapper] == nil {
		err = fmt.Errorf(`No such settings entry for wrapper %s`, wrapper)
	} else if err = setupFn[wrapper](sess.Driver()); err != nil {
		err = fmt.Errorf(`Failed to setup wrapper %s: %q`, wrapper, err)
	}
	return err
}
