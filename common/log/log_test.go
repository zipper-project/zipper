// Copyright (C) 2017, Zipper Team.  All rights reserved.
//
// This file is part of zipper
//
// The zipper is free software: you can use, copy, modify,
// and distribute this software for any purpose with or
// without fee is hereby granted, provided that the above
// copyright notice and this permission notice appear in all copies.
//
// The zipper is distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
// ISC License for more details.
//
// You should have received a copy of the ISC License
// along with this program.  If not, see <https://opensource.org/licenses/isc>.

package log

import (
	"testing"
)

func TestLogToFile(t *testing.T) {
	log, err := New()
	if err != nil {
		t.Errorf("New error %s", err)
	}
	log.Info("test")

	log2, err := New("/tmp/ss.log")
	if err != nil {
		t.Errorf("New with filename error %s", err)
	}

	log2.Debug("test")
	log2.Info("info")
}
