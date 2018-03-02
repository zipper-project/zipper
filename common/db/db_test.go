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
	"testing"

	"upper.io/db.v3"
	"upper.io/db.v3/mongo"
	"upper.io/db.v3/mssql"
	"upper.io/db.v3/mysql"
	"upper.io/db.v3/postgresql"
	"upper.io/db.v3/ql"
	"upper.io/db.v3/sqlite"
)

var wrappers = []string{
	mongo.Adapter,
	mssql.Adapter,
	mysql.Adapter,
	postgresql.Adapter,
	ql.Adapter,
	sqlite.Adapter,
}

func TestSetup(t *testing.T) {
	var err error
	for _, wrapper := range wrappers {
		t.Logf("Testing wrapper: %q TestSetup", wrapper)

		var sess db.Database
		sess, err = Open(wrapper)
		if err != nil {
			t.Fatalf(`Test for wrapper %s failed: %q`, wrapper, err)
		}

		err = Setup(wrapper, sess)
		if err != nil {
			t.Fatalf(`Test for wrapper %s failed: %q`, wrapper, err)
		}

		err = sess.Close()
		if err != nil {
			t.Fatalf(`Test for wrapper %s failed: %q`, wrapper, err)
		}
	}
}
