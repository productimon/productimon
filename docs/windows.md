# Windows Env Setup

1. Install MSYS2 https://www.msys2.org/ to C:\msys64
2. In MSYS2 MinGW, `pacman -S mingw-w64-x86_64-gcc vim git zip unzip patch diffutils sqlite3 sqlite-vfslog`
3. Append the following to `~/.bashrc` in MinGW:

```
export PATH=/c/Python27:/c/Python27/Scripts:/c/msys64/mingw64/bin:$PATH
export BAZEL_VS="C:/VS/2019/BuildTools"
export JAVA_HOME="C:/Program Files/Java/jdk1.8.0_251"
export MSYS2_ARG_CONV_EXCL="*"
```

4. regedit: set `HKLM\SYSTEM\CurrentControlSet\Control\FileSystem LongPathsEnabled` to 1 and reboot windows
5. Install VS Redistributable `https://www.microsoft.com/en-us/download/details.aspx?id=48145`
6. Install Build Tools for Visual Studio 2019 - `https://visualstudio.microsoft.com/downloads/#build-tools-for-visual-studio-2019`, select C++ build tools, install to `C:\VS\2019\BuildTools` to avoid spaces in path (we use this to build bazel, and use mingw-gcc to build productimon)
7. Install JDK8 (MUST BE 8. 11 DOESN'T WORK ON WINDOWS) (`https://www.oracle.com/au/java/technologies/javase/javase-jdk8-downloads.html` - you need to create an oracle account to download because they are evil)
8. Install native Python (don't use pacman in mingw) - `https://www.python.org/downloads/release/python-2718/` (in my experience bazel works much better with Python 2, though 3 should be supported now)
9. Build bazel from HEAD - we can't use stable release until `https://github.com/bazelbuild/bazel/commit/8350681fc900e9ee22ba62db2c5fdeb1e239321a` is cherry-picked

```
wget https://github.com/bazelbuild/bazel/archive/8ea8a3a831aa5deff1f17a75523b87e08ce5dd62.tar.gz
tar zxvf 8ea8a3a831aa5deff1f17a75523b87e08ce5dd62.tar.gz
cd bazel-8ea8a3a831aa5deff1f17a75523b87e08ce5dd62
wget https://github.com/bazelbuild/bazel/releases/download/3.3.1/bazel-3.3.1-windows-x86_64.exe
./bazel-3.3.1-windows-x86_64.exe build //src:bazel-dev.exe"
cp bazel-bin/src/bazel-dev.exe /bin/bazel
```

10. Create `~/.bazelrc` in MinGW with the following content (you need to do this after you compile bazel because we need MSVC compiler to compile bazel itself):

```
build --compiler=mingw-gcc
run --compiler=mingw-gcc
test --compiler=mingw-gcc
```

11. Clone and build productimon!

```
git clone https://git.yiad.am/productimon (and checkout to your desired snapshot)
bazel build //aggregator
bazel build //reporter:reporter_cli
```
