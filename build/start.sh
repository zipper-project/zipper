#!/bin/bash

# kill all zipper and vm
ps x | grep zipper | awk '{print $1}' | xargs kill >null 2>&1
rm -f null

# start zipper
for i in 1 2 3 4 # 5
do
	mkdir -p nohup
 	./bin/zipper --config=$i.yaml > nohup/$i.file 2>&1 &
	#./bin/zipper --config=l0-ca-handshake/000${i}_abc/000${i}_abc.yaml &
done
