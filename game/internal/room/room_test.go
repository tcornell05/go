package room

import (
	"reflect"
	"testing"

	"github.com/hajimehoshi/ebiten/v2"
)

func TestNewRoom(t *testing.T) {
	type args struct {
		name   string
		width  int
		height int
	}
	tests := []struct {
		name string
		args args
		want *Room
	}{
		{
			name: "standard_room",
			args: args{
				name:   "Test Room",
				width:  10,
				height: 10,
			},
			want: &Room{
				Name:         "Test Room",
				Width:        10,
				Height:       10,
				Tiles:        make(map[string]TileData),
				LoadedImages: make(map[string]*ebiten.Image),
			},
		},
		{
			name: "large_room",
			args: args{
				name:   "Large Room",
				width:  20,
				height: 20,
			},
			want: &Room{
				Name:         "Large Room",
				Width:        20,
				Height:       20,
				Tiles:        make(map[string]TileData),
				LoadedImages: make(map[string]*ebiten.Image),
			},
		},
		{
			name: "empty_name",
			args: args{
				name:   "",
				width:  5,
				height: 5,
			},
			want: &Room{
				Name:         "",
				Width:        5,
				Height:       5,
				Tiles:        make(map[string]TileData),
				LoadedImages: make(map[string]*ebiten.Image),
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := NewRoom(tt.args.name, tt.args.width, tt.args.height)
			if got.Name != tt.want.Name {
				t.Errorf("NewRoom() Name = %v, want %v", got.Name, tt.want.Name)
			}
			if got.Width != tt.want.Width {
				t.Errorf("NewRoom() Width = %v, want %v", got.Width, tt.want.Width)
			}
			if got.Height != tt.want.Height {
				t.Errorf("NewRoom() Height = %v, want %v", got.Height, tt.want.Height)
			}
			if got.Tiles == nil {
				t.Error("NewRoom() Tiles map is nil")
			}
			if got.LoadedImages == nil {
				t.Error("NewRoom() LoadedImages map is nil")
			}
		})
	}
}

func TestGetTileKey(t *testing.T) {
	type args struct {
		layer int
		x     int
		y     int
	}
	tests := []struct {
		name string
		args args
		want string
	}{
		{
			name: "floor_tile",
			args: args{layer: 0, x: 5, y: 5},
			want: "0:5:5",
		},
		{
			name: "wall_tile",
			args: args{layer: 1, x: 0, y: 0},
			want: "1:0:0",
		},
		{
			name: "furniture_tile",
			args: args{layer: 2, x: 10, y: 10},
			want: "2:10:10",
		},
		{
			name: "negative_coords",
			args: args{layer: 0, x: -1, y: -1},
			want: "0:-1:-1",
		},
		{
			name: "large_layer",
			args: args{layer: 99, x: 0, y: 0},
			want: "99:0:0",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := GetTileKey(tt.args.layer, tt.args.x, tt.args.y); got != tt.want {
				t.Errorf("GetTileKey() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestRoom_PlaceTile(t *testing.T) {
	type fields struct {
		Name   string
		Width  int
		Height int
	}
	type args struct {
		tile TileData
	}
	tests := []struct {
		name   string
		fields fields
		args   args
		verify func(*Room, *testing.T)
	}{
		{
			name: "place_floor_tile",
			fields: fields{
				Name:   "Test Room",
				Width:  10,
				Height: 10,
			},
			args: args{
				tile: TileData{
					AssetPath: "floor.png",
					X:         5,
					Y:         5,
					Layer:     0,
					Category:  "floor",
				},
			},
			verify: func(r *Room, t *testing.T) {
				key := GetTileKey(0, 5, 5)
				if _, exists := r.Tiles[key]; !exists {
					t.Error("Tile was not placed")
				} else if r.Tiles[key].AssetPath != "floor.png" {
					t.Errorf("Tile AssetPath = %v, want floor.png", r.Tiles[key].AssetPath)
				}
			},
		},
		{
			name: "replace_existing_tile",
			fields: fields{
				Name:   "Test Room",
				Width:  10,
				Height: 10,
			},
			args: args{
				tile: TileData{
					AssetPath: "new_floor.png",
					X:         5,
					Y:         5,
					Layer:     0,
					Category:  "floor",
				},
			},
			verify: func(r *Room, t *testing.T) {
				// First place a tile
				r.PlaceTile(TileData{
					AssetPath: "old_floor.png",
					X:         5,
					Y:         5,
					Layer:     0,
				})
				// Then replace it
				r.PlaceTile(TileData{
					AssetPath: "new_floor.png",
					X:         5,
					Y:         5,
					Layer:     0,
				})
				key := GetTileKey(0, 5, 5)
				if _, exists := r.Tiles[key]; !exists {
					t.Error("Tile was not placed")
				} else if r.Tiles[key].AssetPath != "new_floor.png" {
					t.Errorf("Tile AssetPath = %v, want new_floor.png", r.Tiles[key].AssetPath)
				}
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r := NewRoom(tt.fields.Name, tt.fields.Width, tt.fields.Height)
			r.PlaceTile(tt.args.tile)
			if tt.verify != nil {
				tt.verify(r, t)
			}
		})
	}
}

func TestRoom_GetTile(t *testing.T) {
	type fields struct {
		Name   string
		Width  int
		Height int
		Tiles  map[string]TileData
	}
	type args struct {
		x     int
		y     int
		layer int
	}
	tests := []struct {
		name   string
		fields fields
		args   args
		want   *TileData
	}{
		{
			name: "existing_tile",
			fields: fields{
				Name:   "Test Room",
				Width:  10,
				Height: 10,
				Tiles: map[string]TileData{
					"0:5:5": {
						AssetPath: "floor.png",
						X:         5,
						Y:         5,
						Layer:     0,
					},
				},
			},
			args: args{x: 5, y: 5, layer: 0},
			want: &TileData{
				AssetPath: "floor.png",
				X:         5,
				Y:         5,
				Layer:     0,
			},
		},
		{
			name: "non_existing_tile",
			fields: fields{
				Name:   "Test Room",
				Width:  10,
				Height: 10,
				Tiles:  map[string]TileData{},
			},
			args: args{x: 5, y: 5, layer: 0},
			want: nil,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r := &Room{
				Name:         tt.fields.Name,
				Width:        tt.fields.Width,
				Height:       tt.fields.Height,
				Tiles:        tt.fields.Tiles,
				LoadedImages: make(map[string]*ebiten.Image),
			}
			gotTile, gotExists := r.GetTile(tt.args.layer, tt.args.x, tt.args.y)
			var got *TileData
			if gotExists {
				got = &gotTile
			}
			if tt.want == nil && got != nil {
				t.Errorf("GetTile() = %v, want nil", got)
			} else if tt.want != nil && got == nil {
				t.Error("GetTile() = nil, want non-nil")
			} else if tt.want != nil && got != nil {
				if !reflect.DeepEqual(*got, *tt.want) {
					t.Errorf("GetTile() = %v, want %v", *got, *tt.want)
				}
			}
		})
	}
}

func TestRoom_RemoveTile(t *testing.T) {
	type fields struct {
		Name   string
		Width  int
		Height int
	}
	type args struct {
		x     int
		y     int
		layer int
	}
	tests := []struct {
		name   string
		fields fields
		args   args
		setup  func(*Room)
		verify func(*Room, *testing.T)
	}{
		{
			name: "remove_existing_tile",
			fields: fields{
				Name:   "Test Room",
				Width:  10,
				Height: 10,
			},
			args: args{x: 5, y: 5, layer: 0},
			setup: func(r *Room) {
				r.PlaceTile(TileData{
					AssetPath: "floor.png",
					X:         5,
					Y:         5,
					Layer:     0,
				})
			},
			verify: func(r *Room, t *testing.T) {
				if _, exists := r.GetTile(0, 5, 5); exists {
					t.Error("Tile was not removed")
				}
			},
		},
		{
			name: "remove_non_existing_tile",
			fields: fields{
				Name:   "Test Room",
				Width:  10,
				Height: 10,
			},
			args: args{x: 5, y: 5, layer: 0},
			setup: func(r *Room) {
				// No setup, tile doesn't exist
			},
			verify: func(r *Room, t *testing.T) {
				// Should not crash, just do nothing
				if len(r.Tiles) != 0 {
					t.Error("Tiles map should remain empty")
				}
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r := NewRoom(tt.fields.Name, tt.fields.Width, tt.fields.Height)
			if tt.setup != nil {
				tt.setup(r)
			}
			r.RemoveTile(tt.args.layer, tt.args.x, tt.args.y)
			if tt.verify != nil {
				tt.verify(r, t)
			}
		})
	}
}