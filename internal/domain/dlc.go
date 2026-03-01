package domain

type DLCType string

const (
	TypeExpansionPack DLCType = "EP"
	TypeFreePack      DLCType = "FP"
	TypeGamePack      DLCType = "GP"
	TypeStuffPack     DLCType = "SP"
	TypeKit           DLCType = "Kit"
)

type DLC struct {
	Code string
	Name string
	Type DLCType
}
