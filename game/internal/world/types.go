package world

// TileType represents the type of world tile
type TileType string

const (
	TileTypeRoom     TileType = "room"     // Private rentable room
	TileTypeHallway  TileType = "hallway"  // Connecting corridor
	TileTypeLobby    TileType = "lobby"    // Public gathering space
	TileTypeShop     TileType = "shop"     // Commercial space
	TileTypePublic   TileType = "public"   // Other public spaces
	TileTypeElevator TileType = "elevator" // Vertical transport
	TileTypeStairs   TileType = "stairs"   // Stairs between floors
	TileTypeOutdoor  TileType = "outdoor"  // Outdoor areas
	TileTypeEmpty    TileType = "empty"    // Unassigned/empty tile
)

// AccessLevel defines who can enter a world tile
type AccessLevel string

const (
	AccessPublic     AccessLevel = "public"     // Anyone can enter
	AccessPrivate    AccessLevel = "private"    // Owner/invited only
	AccessRestricted AccessLevel = "restricted" // Special permissions
	AccessStaff      AccessLevel = "staff"      // Staff only
)

// Direction represents cardinal and ordinal directions for tile connections
type Direction string

const (
	DirectionNorth     Direction = "N"
	DirectionNorthEast Direction = "NE"
	DirectionEast      Direction = "E"
	DirectionSouthEast Direction = "SE"
	DirectionSouth     Direction = "S"
	DirectionSouthWest Direction = "SW"
	DirectionWest      Direction = "W"
	DirectionNorthWest Direction = "NW"
	DirectionUp        Direction = "UP"   // For multi-floor
	DirectionDown      Direction = "DOWN" // For multi-floor
)

// RentalStatus represents the rental state of a tile
type RentalStatus string

const (
	RentalAvailable   RentalStatus = "available"
	RentalOccupied    RentalStatus = "occupied"
	RentalReserved    RentalStatus = "reserved"
	RentalMaintenance RentalStatus = "maintenance"
	RentalNotRentable RentalStatus = "not_rentable"
)