package main

import (
	"bytes"
	"strings"
	"testing"
)

func TestRunNonExistentFile(t *testing.T) {
	var buf bytes.Buffer
	err := run("/nonexistent/video-that-does-not-exist.mkv", &buf)
	if err == nil {
		t.Fatal("expected an error for a nonexistent file")
	}
	if !strings.Contains(err.Error(), "could not open") {
		t.Errorf("unexpected error: %v", err)
	}
}

func TestStreamInfoStatus(t *testing.T) {
	tests := []struct {
		name string
		kbps int64
		want string
	}{
		{"positive", 4500, "4500 kb/s"},
		{"zero is unknown", 0, "unknown"},
		{"negative is unknown", -1, "unknown"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := streamInfo{bitrateKbps: tt.kbps}
			if got := s.status(); got != tt.want {
				t.Errorf("status() = %q, want %q", got, tt.want)
			}
		})
	}
}

func TestStreamInfoFormat(t *testing.T) {
	tests := []struct {
		name string
		in   streamInfo
		want []string // expected substrings
	}{
		{
			name: "video",
			in: streamInfo{
				index: 0, typ: streamVideo, codec: "h264",
				resolution: "1920x1080", fps: 29.97, bitrateKbps: 4500,
			},
			want: []string{"Stream #0 — video", "h264", "1920x1080", "29.97 fps", "4500 kb/s"},
		},
		{
			name: "audio",
			in: streamInfo{
				index: 1, typ: streamAudio, codec: "aac",
				sampleRate: 48000, channels: 2, bitrateKbps: 128, language: "eng",
			},
			want: []string{"Stream #1 — audio", "aac", "48000 Hz", "2", "128 kb/s", "eng"},
		},
		{
			name: "subtitle",
			in:   streamInfo{index: 2, typ: streamSubtitle, codec: "subrip", language: "fre"},
			want: []string{"Stream #2 — subtitle", "subrip", "fre"},
		},
		{
			name: "other",
			in:   streamInfo{index: 3, typ: streamOther, codec: "attachment"},
			want: []string{"Stream #3 — other", "attachment"},
		},
		{
			name: "video with unknown bitrate",
			in:   streamInfo{index: 0, typ: streamVideo, codec: "h264", resolution: "640x480"},
			want: []string{"bitrate:    unknown"},
		},
		{
			name: "video with unknown fps",
			in:   streamInfo{index: 0, typ: streamVideo, codec: "h264", resolution: "640x480", bitrateKbps: 100},
			want: []string{"frame rate: unknown"},
		},
		{
			name: "video with unknown resolution",
			in:   streamInfo{index: 0, typ: streamVideo, codec: "h264", fps: 30, bitrateKbps: 100},
			want: []string{"resolution: unknown"},
		},
		{
			name: "audio with unknown sample rate",
			in:   streamInfo{index: 1, typ: streamAudio, codec: "aac"},
			want: []string{"sample rate: unknown"},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			out := tt.in.format()
			for _, sub := range tt.want {
				if !strings.Contains(out, sub) {
					t.Errorf("format() = %q; missing substring %q", out, sub)
				}
			}
		})
	}
}

func TestFileInfoFormat(t *testing.T) {
	f := fileInfo{
		name:        "sample.mp4",
		duration:    60,
		hasDuration: true,
		streams: []streamInfo{
			{index: 0, typ: streamVideo, codec: "h264", resolution: "1920x1080", fps: 30, bitrateKbps: 4500},
		},
		chapters: []chapter{
			{index: 0, title: "Intro", start: 0, end: 5},
		},
	}
	out := f.format()

	for _, sub := range []string{
		"File:     sample.mp4",
		"Duration: 60.00s",
		"Streams:  1",
		"Stream #0 — video",
		"Chapters (1):",
		"#0 Intro                0.00s - 5.00s",
	} {
		if !strings.Contains(out, sub) {
			t.Errorf("format() missing %q\nFull output:\n%s", sub, out)
		}
	}
}

func TestFileInfoFormatNoChapters(t *testing.T) {
	f := fileInfo{name: "a.mp4", duration: 1,
		streams: []streamInfo{{index: 0, typ: streamVideo, codec: "h264", resolution: "1x1"}}}
	if out := f.format(); strings.Contains(out, "Chapters") {
		t.Errorf("unexpected chapters section in output:\n%s", out)
	}
}

func TestFileInfoFormatUnknownDuration(t *testing.T) {
	f := fileInfo{name: "live.mp4",
		streams: []streamInfo{{index: 0, typ: streamVideo, codec: "h264", resolution: "1x1"}}}
	if out := f.format(); !strings.Contains(out, "Duration: unknown") {
		t.Errorf("expected unknown duration, got:\n%s", out)
	}
}

func TestStreamInfoFPSString(t *testing.T) {
	tests := []struct {
		name string
		fps  float64
		want string
	}{
		{"valid", 30, "30.00 fps"},
		{"zero is unknown", 0, "unknown"},
		{"negative is unknown", -1, "unknown"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := streamInfo{fps: tt.fps}
			if got := s.fpsString(); got != tt.want {
				t.Errorf("fpsString() = %q, want %q", got, tt.want)
			}
		})
	}
}

func TestStreamInfoSampleRateString(t *testing.T) {
	tests := []struct {
		name string
		rate int
		want string
	}{
		{"valid", 48000, "48000 Hz"},
		{"zero is unknown", 0, "unknown"},
		{"negative is unknown", -1, "unknown"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := streamInfo{sampleRate: tt.rate}
			if got := s.sampleRateString(); got != tt.want {
				t.Errorf("sampleRateString() = %q, want %q", got, tt.want)
			}
		})
	}
}

func TestStreamTypeString(t *testing.T) {
	tests := []struct {
		typ  streamType
		want string
	}{
		{streamVideo, "video"},
		{streamAudio, "audio"},
		{streamSubtitle, "subtitle"},
		{streamOther, "other"},
	}
	for _, tt := range tests {
		if got := tt.typ.String(); got != tt.want {
			t.Errorf("String() = %q, want %q", got, tt.want)
		}
	}
}
