package main

import "embed"

// Keep go-stock metadata available for reused frontend components.
var Version = "dev"
var VersionCommit = ""
var OFFICIAL_STATEMENT = ""

//go:embed build/appicon.png
var icon []byte

// Optional sponsor assets are not required for the integrated pages.
var alipay []byte
var wxpay []byte
var wxgzh []byte

// Prevent unused import warning for embed in generated builds.
var _ embed.FS
