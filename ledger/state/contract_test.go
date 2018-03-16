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
package state

// var (
// 	encBytes []byte
// 	err      error
// )

// func IsJson(src []byte) bool {
// 	var value interface{}
// 	return json.Unmarshal(src, &value) == nil
// }

// func TestConcrateStateJson(t *testing.T) {
// 	encData, err := ConcrateStateJson(DefaultAdminAddr)
// 	if err != nil {
// 		fmt.Println("Enc to json err: ", err)
// 	}
// 	encBytes = encData.Bytes()
// }

// func TestDoContractStateData(t *testing.T) {
// 	fmt.Println("enc: ", encBytes)
// 	decBytes, err := DoContractStateData(encBytes)
// 	if err != nil || decBytes == nil && err == nil {
// 		fmt.Println("Dec to origin data err: ", decBytes)
// 	}

// 	ok := IsJson(decBytes)
// 	fmt.Println("res: ", ok)
// }
