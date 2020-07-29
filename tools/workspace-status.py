#!/usr/bin/env python

# This script will be run by bazel when the build process starts to
# generate key-value information that represents the status of the
# workspace. The output should be like
#
# KEY1 VALUE1
# KEY2 VALUE2
#
# If the script exits with non-zero code, it's considered as a failure
# and the output will be discarded.

from __future__ import print_function
import os
import subprocess
import getpass
import socket
import sys

ROOT = os.path.abspath(__file__)
while not os.path.exists(os.path.join(ROOT, 'WORKSPACE')):
  ROOT = os.path.dirname(ROOT)

os.chdir(ROOT)

commit = subprocess.check_output(['git', 'rev-parse', 'HEAD']).strip().decode('utf-8')
version = subprocess.check_output(['git', 'describe', '--always', '--match', 'v[0-9].*', '--dirty']).strip().decode('utf-8')
builder = '%s@%s:%s' % (getpass.getuser(), socket.gethostname(), ROOT)

print("STABLE_GIT_COMMIT %s" % commit)
print("STABLE_GIT_VERSION %s" % version)
print("STABLE_BUILDER %s" % builder)

