package math

import "testing"

func TestCartesianToIso(t *testing.T) {
	type args struct {
		x float64
		y float64
	}
	tests := []struct {
		name  string
		args  args
		wantX float64
		wantY float64
	}{
		{
			name:  "origin_point",
			args:  args{x: 0, y: 0},
			wantX: 0,
			wantY: 16,  // Has +16 offset for bottom-center anchoring
		},
		{
			name:  "unit_positive",
			args:  args{x: 1, y: 1},
			wantX: 0,
			wantY: 48,  // 32 + 16 offset
		},
		{
			name:  "unit_negative",
			args:  args{x: -1, y: -1},
			wantX: 0,
			wantY: -16,  // -32 + 16 offset
		},
		{
			name:  "diagonal_right",
			args:  args{x: 2, y: 0},
			wantX: 64,
			wantY: 48,  // 32 + 16 offset
		},
		{
			name:  "diagonal_left",
			args:  args{x: 0, y: 2},
			wantX: -64,
			wantY: 48,  // 32 + 16 offset
		},
		{
			name:  "arbitrary_point",
			args:  args{x: 3.5, y: 2.5},
			wantX: 32,
			wantY: 112,  // 96 + 16 offset
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotX, gotY := CartesianToIso(tt.args.x, tt.args.y)
			if gotX != tt.wantX {
				t.Errorf("CartesianToIso() gotX = %v, want %v", gotX, tt.wantX)
			}
			if gotY != tt.wantY {
				t.Errorf("CartesianToIso() gotY = %v, want %v", gotY, tt.wantY)
			}
		})
	}
}

func TestScreenToIso(t *testing.T) {
	type args struct {
		screenX      float64
		screenY      float64
		camX         float64
		camY         float64
		zoomLevel    float64
		screenWidth  float64
		screenHeight float64
	}
	tests := []struct {
		name  string
		args  args
		wantX int
		wantY int
	}{
		{
			name: "center_screen",
			args: args{
				screenX:      512,
				screenY:      384,
				camX:         0,
				camY:         0,
				zoomLevel:    1.0,
				screenWidth:  1024,
				screenHeight: 768,
			},
			wantX: 0,
			wantY: 0,
		},
		{
			name: "offset_position",
			args: args{
				screenX:      544,
				screenY:      400,
				camX:         0,
				camY:         0,
				zoomLevel:    1.0,
				screenWidth:  1024,
				screenHeight: 768,
			},
			wantX: 1,
			wantY: 0,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotX, gotY := ScreenToIso(tt.args.screenX, tt.args.screenY, tt.args.camX, tt.args.camY, tt.args.zoomLevel, tt.args.screenWidth, tt.args.screenHeight)
			if gotX != tt.wantX {
				t.Errorf("ScreenToIso() gotX = %v, want %v", gotX, tt.wantX)
			}
			if gotY != tt.wantY {
				t.Errorf("ScreenToIso() gotY = %v, want %v", gotY, tt.wantY)
			}
		})
	}
}

func TestIsPointInFloorArea(t *testing.T) {
	type args struct {
		gridX int
		gridY int
	}
	tests := []struct {
		name string
		args args
		want bool
	}{
		{
			name: "valid_center",
			args: args{gridX: 5, gridY: 5},
			want: true,
		},
		{
			name: "valid_corner",
			args: args{gridX: 0, gridY: 0},
			want: true,
		},
		{
			name: "valid_max",
			args: args{gridX: 10, gridY: 10},
			want: true,
		},
		{
			name: "invalid_negative",
			args: args{gridX: -1, gridY: 5},
			want: false,
		},
		{
			name: "invalid_exceed",
			args: args{gridX: 11, gridY: 5},
			want: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := IsPointInFloorArea(tt.args.gridX, tt.args.gridY); got != tt.want {
				t.Errorf("IsPointInFloorArea() = %v, want %v", got, tt.want)
			}
		})
	}
}

func BenchmarkCartesianToIso(b *testing.B) {
	for i := 0; i < b.N; i++ {
		CartesianToIso(float64(i%100), float64(i%100))
	}
}

func BenchmarkScreenToIso(b *testing.B) {
	for i := 0; i < b.N; i++ {
		ScreenToIso(float64(i%100)*32, float64(i%100)*16, 0, 0, 1.0, 1024, 768)
	}
}