#!/bin/bash -e
set -x

cid=`docker run \
	--name 'tests-container' \
	--rm \
	-d \
	patrol-tools \
	tail -f /dev/null`

docker cp tests.sh ${cid}:/tmp/tests.sh
docker exec -it tests-container /tmp/tests.sh
docker rm -f tests-container
