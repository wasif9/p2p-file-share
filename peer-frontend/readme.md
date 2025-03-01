# Front-end Using Qt binding to Golang


## Dependency
[MIQT - mappu](https://github.com/mappu/miqt)


## Requirements
1. [Qt6](https://www.qt.io/)
2. [Go](https://go.dev/dl/)
3. Qt toolchain for Golang
    - [(Windowns) GCC or Clang toolchain](https://build-qt.fsu0413.me/6.2-series/index.html)
    - [pkg-config](https://sourceforge.net/projects/pkgconfiglite/files/0.28-1/)


## (Windows) Go Environment Variables
```sh
$CGO_ENABLED=1 
$CC=gcc
$CXX=g++
$PKG_CONFIG=\Path\to\pgk-config
```

## (Windows) System Environment Variables
Add ``C:\Qt\6.x.x\mingw_64\lib\pkgconfig`` to ``PKG_CONFIG_PATH``

Add ``C:\Qt\6.8.2\mingw_64\bin`` to ``PATH``


## (Windows) Build
1. Run ``go build -ldflags "-s -w -H windowsgui`` to dynamically build (need dependencies installed) an executable binary
2. Run ``windeployqt  app.exe`` to statically build (all dependencies are put in the folder)