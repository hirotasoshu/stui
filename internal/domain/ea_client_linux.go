//go:build linux

package domain

type WinePrefixSource string

const (
	WinePrefixSourceWine    WinePrefixSource = "wine"
	WinePrefixSourceSteam   WinePrefixSource = "steam"
	WinePrefixSourceLutris  WinePrefixSource = "lutris"
	WinePrefixSourceBottles WinePrefixSource = "bottles"
)

type EAClientInfo struct {
	WinePrefix   string
	PrefixSource WinePrefixSource
}
