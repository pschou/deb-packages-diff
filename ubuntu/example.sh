#!/bin/bash

../../deb-get-repomd/deb-get-repomd -mirrors mirrorlist-ubuntu.txt -keyring keys -output reponew -tree  -insecure  -repo dists/xenial/universe/binary-amd64

../../deb-packages-diff/deb-package-diff  -repo 'dists/xenial/universe/binary-amd64'  -new 'reponew/dists/xenial/universe/binary-amd64/'  -output ADDED.txt  -showAdded  -old ''
