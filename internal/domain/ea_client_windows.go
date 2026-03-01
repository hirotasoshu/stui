//go:build windows

package domain

type EAClientType string

const (
	EAClientTypeEAApp  EAClientType = "ea_app"
	EAClientTypeOrigin EAClientType = "origin"
)

type EAClientInfo struct {
	Path       string
	ClientType EAClientType
}
