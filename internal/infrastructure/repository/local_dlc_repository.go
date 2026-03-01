package repository

import (
	"os"
	"path/filepath"

	"stui/internal/domain"
)

type LocalDlcRepository struct{}

func NewLocalDlcRepository() *LocalDlcRepository {
	return &LocalDlcRepository{}
}

var allExpansionPacks = []domain.DLC{
	{Code: "EP01", Name: "Get to Work", Type: domain.TypeExpansionPack},
	{Code: "EP02", Name: "Get Together", Type: domain.TypeExpansionPack},
	{Code: "EP03", Name: "City Living", Type: domain.TypeExpansionPack},
	{Code: "EP04", Name: "Cats & Dogs", Type: domain.TypeExpansionPack},
	{Code: "EP05", Name: "Seasons", Type: domain.TypeExpansionPack},
	{Code: "EP06", Name: "Get Famous", Type: domain.TypeExpansionPack},
	{Code: "EP07", Name: "Island Living", Type: domain.TypeExpansionPack},
	{Code: "EP08", Name: "Discover University", Type: domain.TypeExpansionPack},
	{Code: "EP09", Name: "Eco Lifestyle", Type: domain.TypeExpansionPack},
	{Code: "EP10", Name: "Snowy Escape", Type: domain.TypeExpansionPack},
	{Code: "EP11", Name: "Cottage Living", Type: domain.TypeExpansionPack},
	{Code: "EP12", Name: "High School Years", Type: domain.TypeExpansionPack},
	{Code: "EP13", Name: "Growing Together", Type: domain.TypeExpansionPack},
	{Code: "EP14", Name: "Horse Ranch", Type: domain.TypeExpansionPack},
	{Code: "EP15", Name: "For Rent", Type: domain.TypeExpansionPack},
	{Code: "EP16", Name: "Lovestruck", Type: domain.TypeExpansionPack},
	{Code: "EP17", Name: "Life & Death", Type: domain.TypeExpansionPack},
	{Code: "EP18", Name: "Businesses & Hobbies", Type: domain.TypeExpansionPack},
	{Code: "EP19", Name: "Enchanted by Nature", Type: domain.TypeExpansionPack},
	{Code: "EP20", Name: "Adventure Awaits", Type: domain.TypeExpansionPack},
	{Code: "EP21", Name: "Royalty and Legacy", Type: domain.TypeExpansionPack},
}

var allFreePacks = []domain.DLC{
	{Code: "FP01", Name: "Holiday Celebration Pack", Type: domain.TypeFreePack},
}

var allGamePacks = []domain.DLC{
	{Code: "GP01", Name: "Outdoor Retreat", Type: domain.TypeGamePack},
	{Code: "GP02", Name: "Spa Day", Type: domain.TypeGamePack},
	{Code: "GP03", Name: "Dine Out", Type: domain.TypeGamePack},
	{Code: "GP04", Name: "Vampires", Type: domain.TypeGamePack},
	{Code: "GP05", Name: "Parenthood", Type: domain.TypeGamePack},
	{Code: "GP06", Name: "Jungle Adventure", Type: domain.TypeGamePack},
	{Code: "GP07", Name: "StrangerVille", Type: domain.TypeGamePack},
	{Code: "GP08", Name: "Realm of Magic", Type: domain.TypeGamePack},
	{Code: "GP09", Name: "Star Wars: Journey to Batuu", Type: domain.TypeGamePack},
	{Code: "GP10", Name: "Dream Home Decorator", Type: domain.TypeGamePack},
	{Code: "GP11", Name: "My Wedding Stories", Type: domain.TypeGamePack},
	{Code: "GP12", Name: "Werewolves", Type: domain.TypeGamePack},
}

var allStuffPacks = []domain.DLC{
	{Code: "SP01", Name: "Luxury Party", Type: domain.TypeStuffPack},
	{Code: "SP02", Name: "Perfect Patio", Type: domain.TypeStuffPack},
	{Code: "SP03", Name: "Cool Kitchen", Type: domain.TypeStuffPack},
	{Code: "SP04", Name: "Spooky", Type: domain.TypeStuffPack},
	{Code: "SP05", Name: "Movie Hangout", Type: domain.TypeStuffPack},
	{Code: "SP06", Name: "Romantic Garden", Type: domain.TypeStuffPack},
	{Code: "SP07", Name: "Kids Room", Type: domain.TypeStuffPack},
	{Code: "SP08", Name: "Backyard", Type: domain.TypeStuffPack},
	{Code: "SP09", Name: "Vintage Glamour", Type: domain.TypeStuffPack},
	{Code: "SP10", Name: "Bowling Night", Type: domain.TypeStuffPack},
	{Code: "SP11", Name: "Fitness", Type: domain.TypeStuffPack},
	{Code: "SP12", Name: "Toddler", Type: domain.TypeStuffPack},
	{Code: "SP13", Name: "Laundry Day", Type: domain.TypeStuffPack},
	{Code: "SP14", Name: "My First Pet", Type: domain.TypeStuffPack},
	{Code: "SP15", Name: "Moschino", Type: domain.TypeStuffPack},
	{Code: "SP16", Name: "Tiny Living", Type: domain.TypeStuffPack},
	{Code: "SP17", Name: "Nifty Knitting", Type: domain.TypeStuffPack},
	{Code: "SP18", Name: "Paranormal", Type: domain.TypeStuffPack},
	{Code: "SP46", Name: "Home Chef Hustle Stuff Pack", Type: domain.TypeStuffPack},
}

var allKits = []domain.DLC{
	{Code: "SP20", Name: "Throwback Fit", Type: domain.TypeKit},
	{Code: "SP21", Name: "Country Kitchen", Type: domain.TypeKit},
	{Code: "SP22", Name: "Bust the Dust", Type: domain.TypeKit},
	{Code: "SP23", Name: "Courtyard Oasis", Type: domain.TypeKit},
	{Code: "SP24", Name: "Fashion Street", Type: domain.TypeKit},
	{Code: "SP25", Name: "Industrial Loft", Type: domain.TypeKit},
	{Code: "SP26", Name: "Incheon Arrivals", Type: domain.TypeKit},
	{Code: "SP28", Name: "Modern Menswear Kit", Type: domain.TypeKit},
	{Code: "SP29", Name: "Blooming Rooms", Type: domain.TypeKit},
	{Code: "SP30", Name: "Carnaval Streetwear Kit", Type: domain.TypeKit},
	{Code: "SP31", Name: "Décor to the Max", Type: domain.TypeKit},
	{Code: "SP32", Name: "Moonlight Chic", Type: domain.TypeKit},
	{Code: "SP33", Name: "Little Campers", Type: domain.TypeKit},
	{Code: "SP35", Name: "Desert Luxe", Type: domain.TypeKit},
	{Code: "SP36", Name: "Pastel Pop Kit", Type: domain.TypeKit},
	{Code: "SP37", Name: "Everyday Clutter Kit", Type: domain.TypeKit},
	{Code: "SP38", Name: "Simtimates Collection Kit", Type: domain.TypeKit},
	{Code: "SP39", Name: "Bathroom Clutter Kit", Type: domain.TypeKit},
	{Code: "SP40", Name: "Greenhouse Haven Kit", Type: domain.TypeKit},
	{Code: "SP41", Name: "Basement Treasures Kit", Type: domain.TypeKit},
	{Code: "SP42", Name: "Grunge Revival Kit", Type: domain.TypeKit},
	{Code: "SP43", Name: "Book Nook Kit", Type: domain.TypeKit},
	{Code: "SP44", Name: "Poolside Splash Kit", Type: domain.TypeKit},
	{Code: "SP45", Name: "Modern Luxe Kit", Type: domain.TypeKit},
	{Code: "SP47", Name: "Castle Estate Kit", Type: domain.TypeKit},
	{Code: "SP48", Name: "Goth Galore Kit", Type: domain.TypeKit},
	{Code: "SP49", Name: "Crystal Creations Stuff Pack", Type: domain.TypeKit},
	{Code: "SP50", Name: "Urban Homage Kit", Type: domain.TypeKit},
	{Code: "SP51", Name: "Party Essentials Kit", Type: domain.TypeKit},
	{Code: "SP52", Name: "Riviera Retreat Kit", Type: domain.TypeKit},
	{Code: "SP53", Name: "Cozy Bistro Kit", Type: domain.TypeKit},
	{Code: "SP54", Name: "Artist Studio Kit", Type: domain.TypeKit},
	{Code: "SP55", Name: "Storybook Nursery Kit", Type: domain.TypeKit},
	{Code: "SP56", Name: "Sweet Slumber Party Kit", Type: domain.TypeKit},
	{Code: "SP57", Name: "Cozy Kitsch Kit", Type: domain.TypeKit},
	{Code: "SP58", Name: "Comfy Gamer Kit", Type: domain.TypeKit},
	{Code: "SP59", Name: "Secret Sanctuary Kit", Type: domain.TypeKit},
	{Code: "SP60", Name: "Casanova Cave Kit", Type: domain.TypeKit},
	{Code: "SP61", Name: "Refined Living Room Kit", Type: domain.TypeKit},
	{Code: "SP62", Name: "Business Chic Kit", Type: domain.TypeKit},
	{Code: "SP63", Name: "Sleek Bathroom Kit", Type: domain.TypeKit},
	{Code: "SP64", Name: "Sweet Allure Kit", Type: domain.TypeKit},
	{Code: "SP65", Name: "Restoration Workshop Kit", Type: domain.TypeKit},
	{Code: "SP66", Name: "Golden Years Kit", Type: domain.TypeKit},
	{Code: "SP67", Name: "Kitchen Clutter Kit", Type: domain.TypeKit},
	{Code: "SP68", Name: "SpongeBob's House Kit", Type: domain.TypeKit},
	{Code: "SP69", Name: "Autumn Apparel Kit", Type: domain.TypeKit},
	{Code: "SP70", Name: "SpongeBob Kid's Room Kit", Type: domain.TypeKit},
	{Code: "SP71", Name: "Grange Mudroom Kit", Type: domain.TypeKit},
	{Code: "SP72", Name: "Essential Glam Kit", Type: domain.TypeKit},
	{Code: "SP73", Name: "Modern Retreat Kit", Type: domain.TypeKit},
	{Code: "SP74", Name: "Garden to Table Kit", Type: domain.TypeKit},
	{Code: "SP75", Name: "Wonderland Playroom Kit", Type: domain.TypeKit},
	{Code: "SP76", Name: "Silver Screen Style Kit", Type: domain.TypeKit},
	{Code: "SP77", Name: "Tea Time Solarium Kit", Type: domain.TypeKit},
	{Code: "SP81", Name: "Prairie Dreams Kit", Type: domain.TypeKit},
	{Code: "SP82", Name: "Yard Charm Kit", Type: domain.TypeKit},
}

func (r *LocalDlcRepository) GetExpansionPacks() []domain.DLC {
	return allExpansionPacks
}

func (r *LocalDlcRepository) GetFreePacks() []domain.DLC {
	return allFreePacks
}

func (r *LocalDlcRepository) GetGamePacks() []domain.DLC {
	return allGamePacks
}

func (r *LocalDlcRepository) GetStuffPacks() []domain.DLC {
	return allStuffPacks
}

func (r *LocalDlcRepository) GetKits() []domain.DLC {
	return allKits
}

func (r *LocalDlcRepository) GetInstalledExpansionPacks(gamePath string) []domain.DLC {
	return filterInstalled(allExpansionPacks, gamePath)
}

func (r *LocalDlcRepository) GetInstalledFreePacks(gamePath string) []domain.DLC {
	return filterInstalled(allFreePacks, gamePath)
}

func (r *LocalDlcRepository) GetInstalledGamePacks(gamePath string) []domain.DLC {
	return filterInstalled(allGamePacks, gamePath)
}

func (r *LocalDlcRepository) GetInstalledStuffPacks(gamePath string) []domain.DLC {
	return filterInstalled(allStuffPacks, gamePath)
}

func (r *LocalDlcRepository) GetInstalledKits(gamePath string) []domain.DLC {
	return filterInstalled(allKits, gamePath)
}

func filterInstalled(dlcs []domain.DLC, gamePath string) []domain.DLC {
	var installed []domain.DLC
	for _, dlc := range dlcs {
		dlcPath := filepath.Join(gamePath, dlc.Code)
		if _, err := os.Stat(dlcPath); err == nil {
			installed = append(installed, dlc)
		}
	}
	return installed
}
