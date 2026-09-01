package main

/*
#cgo pkg-config: libavformat libavcodec libavutil
#include <libavformat/avformat.h>
#include <libavcodec/avcodec.h>
#include <libavutil/dict.h>
*/
import "C"

import (
	"fmt"
	"io"
	"os"
	"unsafe"
)

func main() {
	if len(os.Args) < 2 {
		fmt.Fprintln(os.Stderr, "usage: videoinfo <file>")
		os.Exit(1)
	}
	if err := run(os.Args[1], os.Stdout); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

// run opens filename and writes its metadata report to w.
func run(filename string, w io.Writer) error {
	cFilename := C.CString(filename)
	defer C.free(unsafe.Pointer(cFilename))

	var fmtCtx *C.AVFormatContext

	if ret := C.avformat_open_input(&fmtCtx, cFilename, nil, nil); ret < 0 {
		return fmt.Errorf("could not open %q (code %d)", filename, int(ret))
	}
	defer C.avformat_close_input(&fmtCtx)

	if ret := C.avformat_find_stream_info(fmtCtx, nil); ret < 0 {
		return fmt.Errorf("could not read stream info (code %d)", int(ret))
	}

	info := fileInfo{
		name:        filename,
		duration:    float64(fmtCtx.duration) / float64(C.AV_TIME_BASE),
		hasDuration: fmtCtx.duration != C.AV_NOPTS_VALUE,
	}

	streams := unsafe.Slice(fmtCtx.streams, int(fmtCtx.nb_streams))
	for i, st := range streams {
		info.streams = append(info.streams, gatherStream(i, st))
	}

	if n := int(fmtCtx.nb_chapters); n > 0 {
		chapters := unsafe.Slice(fmtCtx.chapters, n)
		for i, ch := range chapters {
			tb := 1.0
			if den := ch.time_base.den; den != 0 {
				tb = float64(ch.time_base.num) / float64(den)
			}
			info.chapters = append(info.chapters, chapter{
				index: i,
				title: metadataValue(ch.metadata, "title"),
				start: float64(ch.start) * tb,
				end:   float64(ch.end) * tb,
			})
		}
	}

	_, err := io.WriteString(w, info.format())
	return err
}

func gatherStream(i int, st *C.AVStream) streamInfo {
	params := st.codecpar

	codecName := unknown
	if dec := C.avcodec_find_decoder(params.codec_id); dec != nil {
		codecName = C.GoString(dec.name)
	}

	s := streamInfo{
		index:       i,
		codec:       codecName,
		language:    metadataValue(st.metadata, "language"),
		bitrateKbps: int64(params.bit_rate) / 1000,
	}

	switch params.codec_type {
	case C.AVMEDIA_TYPE_VIDEO:
		fps := 0.0
		if st.avg_frame_rate.den != 0 {
			fps = float64(st.avg_frame_rate.num) / float64(st.avg_frame_rate.den)
		}
		s.typ = streamVideo
		s.fps = fps
		s.resolution = fmt.Sprintf("%dx%d", int(params.width), int(params.height))

	case C.AVMEDIA_TYPE_AUDIO:
		s.typ = streamAudio
		s.sampleRate = int(params.sample_rate)
		s.channels = int(params.ch_layout.nb_channels)

	case C.AVMEDIA_TYPE_SUBTITLE:
		s.typ = streamSubtitle

	default:
		s.typ = streamOther
	}

	return s
}

// metadataValue reads a key (e.g. "language", "title") out of an AVDictionary.
func metadataValue(dict *C.AVDictionary, key string) string {
	cKey := C.CString(key)
	defer C.free(unsafe.Pointer(cKey))

	entry := C.av_dict_get(dict, cKey, nil, 0)
	if entry == nil {
		return unknown
	}
	return C.GoString(entry.value)
}
