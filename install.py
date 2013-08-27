#!/usr/bin/env python
# This script provides a standard installation of sbp-linux-config through
# the following steps:
#   1. It copies everything from ./src to ./bin.
#   2. It makes several symlinks in standard places (such as ~) that point
#      to the appropriate files in ./bin.
#   3. If arguments are provided, each is interpreted as a directory which
#      may contain zero or more subdirectories corresponding to the
#      subdirectories of ./src. Each file in each of these directires is
#      read in and appended to the corresponding file in ./bin. If no such
#      file exists yet in ./bin, it is created with the appended contents.
#      This provides a simple mechanism for adding per-machine customizations.

import os
import os.path as p
import shutil
import sys
import subprocess

HOME = p.expanduser('~')
SBP_LINUX_CONFIG = p.join(HOME, 'sbp-linux-config')
SRC = p.join(SBP_LINUX_CONFIG, 'src')
BIN = p.join(SBP_LINUX_CONFIG, 'bin')
DOTFILES_BIN = p.join(BIN, 'dotfiles')
SCRIPTS_BIN = p.join(BIN, 'scripts')
I3STATUS_CONF = p.join(BIN, 'dotfiles/i3status.conf')

# Some utility methods for other install scripts to use for manipulating the
# output of this script:

def readFile(name):
  with open(name) as f:
    return f.read()

def writeFile(name, text):
  with open(name, 'w') as f:
    f.write(text)

def insertBefore(text, afterLine, newLine):
  """ Inserts newLine into text, right before afterLine. """
  lines = text.splitlines()
  lineNum = lines.index(afterLine)
  lines.insert(lineNum, newLine)
  return '\n'.join(lines)

# Helper function.
def forceLink(target, linkName):
  """ Forces a symlink, even if the linkName already exists. """
  if p.islink(linkName) or p.isfile(linkName):
    print 'Deleting existing file %s ...' % linkName
    os.remove(linkName)

  # Don't handle the case where linkName is a directory--it's too easy to
  # blow away existing config folders that way.

  if 'fjiji3' in target:
    pass
  else:
    print 'Linking %s as %s ...' % (target, linkName)
    os.symlink(target, linkName)

# Recursive helper for linking over individual files in the tree rooted at
# dotfiles.
def linkDotfiles(targetDir, linkDir, addDot):
  if not p.exists(linkDir):
    print 'Creating %s ...' % linkDir
    os.mkdir(linkDir)

  for childName in os.listdir(targetDir):
    targetChild = p.join(targetDir, childName)
 
    linkChildName = '.' + childName if addDot else childName
    linkChild = p.join(linkDir, linkChildName)

    if p.isfile(targetChild):
      forceLink(targetChild, linkChild)
    elif p.isdir(targetChild):
      # Recurse, and don't add any more dots.
      linkDotfiles(targetChild, linkChild, False)


def standard(appendDirs):
  """ Invokes the standard install procedure. """

  # Clean out any existing bin stuff.
  if p.isdir(BIN):
    shutil.rmtree(BIN)

  # Perform the copy.
  shutil.copytree(SRC, BIN)

  # Process arguments to see if they contain append-files.
  for appendDir in appendDirs:
    if not p.isdir(appendDir):
      print 'ERROR: %s is not a directory.' % appendDir
    else:
      # Look at every file in the appendDir.
      for root, dirs, files in os.walk(appendDir):
        # Make root relative to the appendDir, since we'll want to use it both in
        # the appendDir and in BIN.
        root = p.relpath(root, appendDir)
        for fil in files:
          # Compute the full path from the appendDir to the file.
          fil = p.join(root, fil)

          appendSource = p.join(appendDir, fil)
          appendDest = p.join(BIN, fil)

          if p.exists(appendDest):
            print 'Appending %s to %s ...' % (appendSource, appendDest)
            with open(appendDest) as f:
              text = f.read()
            while not text.endswith('\n\n'):
              text += '\n'
            with open(appendSource) as f:
              text += f.read()
            with open(appendDest, 'w') as f:
              f.write(text)
          else:
            print 'Copying %s to %s ...' % (appendSource, appendDest)
            shutil.copy(appendSource, appendDest)

  # Link over dotfiles.
  linkDotfiles(DOTFILES_BIN, HOME, True)

  # Prevent GNOME's nautilus from leaving behind those weird "Desktop" windows.
  # Link in all the other scripts that should be on the path.
  forceLink(SCRIPTS_BIN, p.join(HOME, 'bin'))

  # This may print some errors if there is no X session; suppress those errors.
  with open('/dev/null', 'w') as sink:
    subprocess.call(['gsettings', 'set', 'org.gnome.desktop.background',
        'show-desktop-icons', 'false'], stderr=sink)


def standardLaptop():
  """ Meant to be invoked after standard() for laptops. Adds some useful
  configuration settings for laptops.
  """
  text = readFile(I3STATUS_CONF)

  print 'Inserting Battery entry into i3status.conf ...'
  text = insertBefore(text, 'order += "cpu_usage"', 'order += "battery 0"')

  print 'Inserting Wi-Fi entry into i3status.conf ...'
  text = insertBefore(text,
      'order += "ethernet eth0"', 'order += "wireless wlan0"')

  writeFile(I3STATUS_CONF, text)


if __name__ == '__main__':
  standard(sys.argv[1:])
