#!/bin/python

import datetime
import hashlib
import os
import sys

from collections import defaultdict
from operator import itemgetter

def hashfile(hfile, hasher, blocksize=65536):
    with open(hfile, 'rb') as afile:
        buf = afile.read(blocksize)
        while len(buf) > 0:
            hasher.update(buf)
            buf = afile.read(blocksize)
    return hasher.hexdigest()

def walk(path, hashfiles):
    for root, dirs, files in os.walk(path):
        for name in files:
            filepath = os.path.join(root, name)
            hexdigest = hashfile(filepath, hashlib.md5())
            hashfiles[hexdigest].append((filepath, os.path.getmtime(filepath)))

def check_dupes(hf):
    for k, v in hf.items():
        if len(v) > 1:
            print "\n", k
            v.sort(key=itemgetter(1))
            for filepath, mtime in v:
                dt = datetime.datetime.fromtimestamp(mtime).strftime('%Y-%m-%d %H:%M:%S')
                print dt, " .. ", filepath

if __name__ == '__main__':
    path = sys.argv[1]
    hashfiles = defaultdict(list)
    if path == None:
        pass
    else:
        walk(path, hashfiles)

    check_dupes(hashfiles)
