#!/bin/bash -e

cd outputs
cp ../example-config.yaml .
zip ../smartbackup-${BUILD_NUM}.zip *
cd ..

echo AWS access key is ${AWS_ACCESS_KEY_ID}
echo AWS profile is ${AWS_PROFILE}
echo Build branch is ${BUILD_BRANCH}

aws s3 cp smartbackup-${BUILD_NUM}.zip ${OUTPUT_URL}/${BUILD_NUM}/smartbackup-${BUILD_NUM}.zip --acl public-read

if [ "${BUILD_BRANCH}" == "master" ]; then
    aws s3 cp smartbackup-${BUILD_NUM}.zip ${OUTPUT_URL}/latest/smartbackup.zip --acl public-read
fi