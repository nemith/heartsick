# heartsick :broken_heart:

[![CircleCI](https://img.shields.io/circleci/project/github/nemith/heartsick/master.svg)](https://circleci.com/gh/nemith/heartsick)
[![codecov](https://img.shields.io/codecov/c/gh/nemith/heartsick.svg)](https://codecov.io/gh/nemith/heartsick)

**NOTE: This isn't yet complete and probably shouldn't be used unless you are brave.**

An direct replacement of [homesick](github.com/technicalpickles/homesick) but written in [Go](http://golang.org).  This means zero runtime requirement, just download a binary, run and go on your way.

## Chages from homesick
 * Some aliases were added that I always type:
     * path = showpath
     * shell = cd
     * symlink = link
 * `rc` command isn't implemented yet


 ## TODO
 - [ ] Implement `rc` that reads shebangs for multiple language support (fallback to executing with ruby if shebang missing)
 - [ ] More unit tests
 - [ ] Add flags to overwrite destination castle director