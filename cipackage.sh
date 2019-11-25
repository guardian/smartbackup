#!/bin/bash -e

cd outputs
zip ../smartbackup-${BUILD_NUM}.zip *
cd ..

aws s3 cp smartbackup-${BUILD_NUM}.zip ${OUTPUT_URL}/${BUILD_NUM}/smartbackup-${BUILD_NUM}.zip --acl-public

if [ "${BUILD_BRANCH}" == "master" ]; then
    aws s3 cp smartbackup-${BUILD_NUM}.zip ${OUTPUT_URL}/latest/smartbackup.zip --acl-public
fi