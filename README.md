# videoinfo

A small command-line tool that prints technical metadata (codec, resolution,
frame rate, bitrate, language, chapters, …) for a media file using FFmpeg's
libavformat/libavcodec via cgo.

## Requirements

- Go 1.26+
- FFmpeg development libraries (`libavformat`, `libavcodec`, `libavutil`) and
  `pkg-config`. On macOS with Homebrew: `brew install ffmpeg`. On Debian/Ubuntu:
  `sudo apt install libavformat-dev libavcodec-dev libavutil-dev pkg-config`.

> cgo locates the FFmpeg libraries through `pkg-config`. If they are not found,
> set `PKG_CONFIG_PATH` to the directory containing the `.pc` files (e.g.
> `/opt/homebrew/lib/pkgconfig`).

## Build

```sh
make build        # produces ./videoinfo
make test         # run unit tests
make vet          # static analysis
```

## Usage

```sh
./videoinfo <file>
```

Example output:

```
File:     sample.mp4
Duration: 60.00s
Streams:  2

Stream #0 — video
  codec:      h264
  resolution: 1920x1080
  frame rate: 30.00 fps
  bitrate:    4500 kb/s

Stream #1 — audio
  codec:       aac
  sample rate: 48000 Hz
  channels:    2
  bitrate:     128 kb/s
  language:    eng
```

## Install

```sh
make install                # installs to /usr/local/bin
make install PREFIX=~/.local   # custom prefix
```

## Uninstall

```sh
make uninstall
```
