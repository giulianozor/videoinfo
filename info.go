package main

import (
	"fmt"
	"strings"
)

// streamType identifies the kind of media stream.
type streamType int

// unknown marks a metadata value that the container did not provide, so the
// report never prints misleading zeroed numbers (0 kb/s, 0.00 fps, 0x0, …).
const unknown = "unknown"

const (
	streamVideo streamType = iota
	streamAudio
	streamSubtitle
	streamOther
)

func (t streamType) String() string {
	switch t {
	case streamVideo:
		return "video"
	case streamAudio:
		return "audio"
	case streamSubtitle:
		return "subtitle"
	default:
		return "other"
	}
}

// streamInfo holds the metadata gathered for a single stream.
type streamInfo struct {
	index       int
	typ         streamType
	codec       string
	resolution  string
	fps         float64
	bitrateKbps int64
	sampleRate  int
	channels    int
	language    string
}

// chapter represents one chapter within the container.
type chapter struct {
	index int
	title string
	start float64
	end   float64
}

// fileInfo is the fully gathered metadata for an opened file.
type fileInfo struct {
	name        string
	duration    float64
	hasDuration bool
	streams     []streamInfo
	chapters    []chapter
}

// status returns a human-readable bitrate, marking unknown values as such
// instead of printing a misleading "0 kb/s".
func (s streamInfo) status() string {
	if s.bitrateKbps <= 0 {
		return unknown
	}
	return fmt.Sprintf("%d kb/s", s.bitrateKbps)
}

// fpsString renders the frame rate, marking unknown values as such.
func (s streamInfo) fpsString() string {
	if s.fps <= 0 {
		return unknown
	}
	return fmt.Sprintf("%.2f fps", s.fps)
}

// sampleRateString renders the sample rate, marking unknown values as such.
func (s streamInfo) sampleRateString() string {
	if s.sampleRate <= 0 {
		return unknown
	}
	return fmt.Sprintf("%d Hz", s.sampleRate)
}

// format renders the stream as a block of output text.
func (s streamInfo) format() string {
	switch s.typ {
	case streamVideo:
		resolution := s.resolution
		if resolution == "" || resolution == "0x0" {
			resolution = unknown
		}
		return fmt.Sprintf("Stream #%d — video\n  codec:      %s\n  resolution: %s\n  frame rate: %s\n  bitrate:    %s\n",
			s.index, s.codec, resolution, s.fpsString(), s.status())
	case streamAudio:
		return fmt.Sprintf("Stream #%d — audio\n  codec:       %s\n  sample rate: %s\n  channels:    %d\n  bitrate:     %s\n  language:    %s\n",
			s.index, s.codec, s.sampleRateString(), s.channels, s.status(), s.language)
	case streamSubtitle:
		return fmt.Sprintf("Stream #%d — subtitle\n  codec:    %s\n  language: %s\n",
			s.index, s.codec, s.language)
	default:
		return fmt.Sprintf("Stream #%d — other\n  codec: %s\n", s.index, s.codec)
	}
}

// format renders the full file report.
func (f fileInfo) format() string {
	var b strings.Builder
	fmt.Fprintf(&b, "File:     %s\n", f.name)
	if f.hasDuration {
		fmt.Fprintf(&b, "Duration: %.2fs\n", f.duration)
	} else {
		b.WriteString("Duration: " + unknown + "\n")
	}
	fmt.Fprintf(&b, "Streams:  %d\n\n", len(f.streams))

	for _, s := range f.streams {
		b.WriteString(s.format())
		b.WriteString("\n")
	}

	if len(f.chapters) > 0 {
		fmt.Fprintf(&b, "Chapters (%d):\n", len(f.chapters))
		for _, ch := range f.chapters {
			fmt.Fprintf(&b, "  #%d %-20s %.2fs - %.2fs\n", ch.index, ch.title, ch.start, ch.end)
		}
	}
	return b.String()
}
