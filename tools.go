//go:build tools
// +build tools

package main

import (
	_ "github.com/pion/webrtc/v4/pkg/media"
	_ "github.com/pion/webrtc/v4/pkg/rtcerr"
)
