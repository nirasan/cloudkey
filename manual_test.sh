#!/bin/sh

if [ ! -f config.toml ]; then
    echo "Please create config file named 'config.toml'"
    echo ""
    echo "go run *.go create gcp [g|s] -c config.toml"
    echo "go run *.go create aws [st|sh] -c config.toml"
    exit 1
fi

###################
# prepare
###################

tmpdir=$(mktemp -d test.XXXXXXXXXX)
cd $tmpdir
echo "-- test dir is " $tmpdir

cp ../*.go .
cp ../config.toml .

echo "-- create file1"
cat <<EOL >file1
aaa
bbb
ccc
EOL

echo "-- cp file1 file1.orig"
cp file1 file1.orig

###################
# test encrypt and decrypt using gcp
###################

echo ""
echo "-- encrypt file1 using gcp"
go run *.go encrypt gcp -c config.toml -e .gcpencrypted file1

if [ ! -f file1.gcpencrypted ]; then
    echo "-- ERROR: encrypted file is not exists"
    rm file1
fi

echo "-- decrypt file1 using gcp"
go run *.go decrypt gcp -c config.toml -e .gcpencrypted file1.gcpencrypted

result=$(diff file1 file1.orig)
if [ "$result" != "" ]; then
    echo "-- ERROR: original file and decrypted file is not same"
fi

###################
# test encrypt and decrypt using aws
###################

echo ""
echo "-- encrypt file1 using aws"
go run *.go encrypt aws -c config.toml -e .awsencrypted file1

if [ ! -f file1.awsencrypted ]; then
    echo "-- ERROR: encrypted file is not exists"
    rm file1
fi

echo "-- decrypt file1 using aws"
go run *.go decrypt aws -c config.toml -e .awsencrypted file1.awsencrypted

result=$(diff file1 file1.orig)
if [ "$result" != "" ]; then
    echo "-- ERROR: original file and decrypted file is not same"
fi



###################
# test re-encrypt and re-decrypt using gcp
###################

echo ""
echo "-- prepare re-encrypt and re-decrypt using gcp"
mkdir -p dir1/dir2

cat <<EOL > dir1/file2
111, 222, 333
EOL

cat <<EOL > dir1/dir2/file3
xxx, xxx, xxx
EOL

file2=dir1/file2
file3=dir1/dir2/file3

go run *.go encrypt gcp -c config.toml -e .gcpencrypted $file2
go run *.go encrypt gcp -c config.toml -e .gcpencrypted $file3

echo ""
echo "-- re-encrypt files using gcp"

cp ${file2}.gcpencrypted ${file2}.gcpencrypted.orig
cp ${file3}.gcpencrypted ${file3}.gcpencrypted.orig

echo aaa >> dir1/file2
echo aaa >> dir1/dir2/file3

go run *.go re-encrypt gcp -c config.toml -e .gcpencrypted dir1

result2=$(diff ${file2}.gcpencrypted ${file2}.gcpencrypted.orig)
if [ "$result2" = "" ]; then
    echo "-- ERROR: re-encrypt file2"
fi

result3=$(diff ${file3}.gcpencrypted ${file3}.gcpencrypted.orig)
if [ "$result3" = "" ]; then
    echo "-- ERROR: re-encrypt file3"
fi

echo ""
echo "-- re-decrypt files using gcp"

mv ${file2} ${file2}.orig
mv ${file3} ${file3}.orig

go run *.go re-decrypt gcp -c config.toml -e .gcpencrypted dir1

result2=$(diff ${file2} ${file2}.orig)
if [ "$result2" != "" ]; then
    echo "-- ERROR: re-decrypt file2"
fi

result3=$(diff ${file3} ${file3}.orig)
if [ "$result3" != "" ]; then
    echo "-- ERROR: re-decrypt file3"
fi



###################
# test re-encrypt and re-decrypt using aws
###################

echo ""
echo "-- prepare re-encrypt and re-decrypt using aws"
mkdir -p dirA/dirB

file2=dirA/file2
file3=dirA/dirB/file3

cat <<EOL > $file2
111, 222, 333
EOL

cat <<EOL > $file3
xxx, xxx, xxx
EOL

go run *.go encrypt aws -c config.toml -e .awsencrypted $file2
go run *.go encrypt aws -c config.toml -e .awsencrypted $file3

echo ""
echo "-- re-encrypt files using aws"

cp ${file2}.awsencrypted ${file2}.awsencrypted.orig
cp ${file3}.awsencrypted ${file3}.awsencrypted.orig

echo aaa >> $file2
echo aaa >> $file3

go run *.go re-encrypt aws -c config.toml -e .awsencrypted dirA

result2=$(diff ${file2}.awsencrypted ${file2}.awsencrypted.orig)
if [ "$result2" = "" ]; then
    echo "-- ERROR: re-encrypt file2"
fi

result3=$(diff ${file3}.awsencrypted ${file3}.awsencrypted.orig)
if [ "$result3" = "" ]; then
    echo "-- ERROR: re-encrypt file3"
fi

echo ""
echo "-- re-decrypt files using aws"

mv ${file2} ${file2}.orig
mv ${file3} ${file3}.orig

go run *.go re-decrypt aws -c config.toml -e .awsencrypted dirA

result2=$(diff ${file2} ${file2}.orig)
if [ "$result2" != "" ]; then
    echo "-- ERROR: re-decrypt file2"
fi

result3=$(diff ${file3} ${file3}.orig)
if [ "$result3" != "" ]; then
    echo "-- ERROR: re-decrypt file3"
fi

cd ../
rm -rf $tmpdir
