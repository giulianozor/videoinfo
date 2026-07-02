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
	"os"
	"unsafe"
)

func main() {
	if len(os.Args) < 2 {
		fmt.Fprintln(os.Stderr, "usage: mp4info <file.mp4>")
		os.Exit(1)
	}
	filename := os.Args[1]

	cFilename := C.CString(filename)
	defer C.free(unsafe.Pointer(cFilename))

	var fmtCtx *C.AVFormatContext

	if ret := C.avformat_open_input(&fmtCtx, cFilename, nil, nil); ret < 0 {
		fmt.Fprintf(os.Stderr, "could not open %q (code %d)\n", filename, int(ret))
		os.Exit(1)
	}
	defer C.avformat_close_input(&fmtCtx)

	if ret := C.avformat_find_stream_info(fmtCtx, nil); ret < 0 {
		fmt.Fprintf(os.Stderr, "could not read stream info (code %d)\n", int(ret))
		os.Exit(1)
	}

	fmt.Printf("File:     %s\n", filename)
	fmt.Printf("Duration: %.2fs\n", float64(fmtCtx.duration)/float64(C.AV_TIME_BASE))
	fmt.Printf("Streams:  %d\n\n", int(fmtCtx.nb_streams))

	streams := unsafe.Slice(fmtCtx.streams, int(fmtCtx.nb_streams))

	for i, st := range streams {
		params := st.codecpar
		codecName := "unknown"
		if dec := C.avcodec_find_decoder(params.codec_id); dec != nil {
			codecName = C.GoString(dec.name)
		}

		switch params.codec_type {
		case C.AVMEDIA_TYPE_VIDEO:
			fps := 0.0
			if st.avg_frame_rate.den != 0 {
				fps = float64(st.avg_frame_rate.num) / float64(st.avg_frame_rate.den)
			}
			fmt.Printf("Stream #%d — video\n", i)
			fmt.Printf("  codec:      %s\n", codecName)
			fmt.Printf("  resolution: %dx%d\n", int(params.width), int(params.height))
			fmt.Printf("  frame rate: %.2f fps\n", fps)
			fmt.Printf("  bitrate:    %d kb/s\n\n", int64(params.bit_rate)/1000)

		case C.AVMEDIA_TYPE_AUDIO:
			fmt.Printf("Stream #%d — audio\n", i)
			fmt.Printf("  codec:       %s\n", codecName)
			fmt.Printf("  sample rate: %d Hz\n", int(params.sample_rate))
			fmt.Printf("  channels:    %d\n", int(params.ch_layout.nb_channels))
			fmt.Printf("  bitrate:     %d kb/s\n", int64(params.bit_rate)/1000)
			fmt.Printf("  language:    %s\n\n", metadataValue(st.metadata, "language"))

		case C.AVMEDIA_TYPE_SUBTITLE:
			fmt.Printf("Stream #%d — subtitle\n", i)
			fmt.Printf("  codec:    %s\n", codecName)
			fmt.Printf("  language: %s\n\n", metadataValue(st.metadata, "language"))
		}
	}

	if n := int(fmtCtx.nb_chapters); n > 0 {
		fmt.Printf("Chapters (%d):\n", n)
		chapters := unsafe.Slice(fmtCtx.chapters, n)
		for i, ch := range chapters {
			tb := float64(ch.time_base.num) / float64(ch.time_base.den)
			start := float64(ch.start) * tb
			end := float64(ch.end) * tb
			title := metadataValue(ch.metadata, "title")
			fmt.Printf("  #%d %-20s %.2fs - %.2fs\n", i, title, start, end)
		}
	}
}

// metadataValue reads a key (e.g. "language", "title") out of an AVDictionary.
func metadataValue(dict *C.AVDictionary, key string) string {
	cKey := C.CString(key)
	defer C.free(unsafe.Pointer(cKey))

	entry := C.av_dict_get(dict, cKey, nil, 0)
	if entry == nil {
		return "unknown"
	}
	return C.GoString(entry.value)
}
