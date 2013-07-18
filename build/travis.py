#! /usr/bin/env python

import os, os.path, boto.s3.connection

access_key = os.getenv("AWS_ACCESS_KEY")
secret_key = os.getenv("AWS_SECRET_KEY")
bucket = os.getenv("BUCKET")
version = os.getenv("VERSION")
goos = os.getenv("GOOS")

s3 = boto.s3.connection.S3Connection(access_key, secret_key)
bucket = s3.get_bucket(bucket)

for envpath in ["NGROK", "NGROKD"]:
	file_path = os.getenv(envpath)
	if goos == "windows":
		file_path += ".exe"
	dir_path, name = os.path.split(file_path)
	_, platform = os.path.split(dir_path)
	key_name = "%s/%s/%s" % (platform, version, name)
	key = bucket.new_key(key_name)
	key.set_contents_from_filename(file_path)
