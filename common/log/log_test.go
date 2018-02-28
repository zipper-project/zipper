// Copyright (C) 2017, Zipper Team Technology Co.,Ltd.  All rights reserved.
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
