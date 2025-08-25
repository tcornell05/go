package entity

import (
	"testing"
)

func TestNewPlayer(t *testing.T) {
	type args struct {
		startX float64
		startY float64
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
		verify  func(*Player, *testing.T)
	}{
		{
			name: "create_player_at_origin",
			args: args{
				startX: 0,
				startY: 0,
			},
			wantErr: false,
			verify: func(p *Player, t *testing.T) {
				if p.X != 0 || p.Y != 0 {
					t.Errorf("Player position = (%v, %v), want (0, 0)", p.X, p.Y)
				}
				if p.Speed != 0.1 {
					t.Errorf("Player speed = %v, want 0.1", p.Speed)
				}
				if p.Facing != 0 {
					t.Errorf("Player facing = %v, want 0 (South)", p.Facing)
				}
			},
		},
		{
			name: "create_player_at_position",
			args: args{
				startX: 5.5,
				startY: 7.5,
			},
			wantErr: false,
			verify: func(p *Player, t *testing.T) {
				if p.X != 5.5 || p.Y != 7.5 {
					t.Errorf("Player position = (%v, %v), want (5.5, 7.5)", p.X, p.Y)
				}
			},
		},
		{
			name: "negative_position",
			args: args{
				startX: -10,
				startY: -10,
			},
			wantErr: false,
			verify: func(p *Player, t *testing.T) {
				if p.X != -10 || p.Y != -10 {
					t.Errorf("Player position = (%v, %v), want (-10, -10)", p.X, p.Y)
				}
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := NewPlayer(tt.args.startX, tt.args.startY)
			if (err != nil) != tt.wantErr {
				t.Errorf("NewPlayer() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if tt.verify != nil && got != nil {
				tt.verify(got, t)
			}
		})
	}
}

func TestPlayer_Rotate(t *testing.T) {
	tests := []struct {
		name         string
		startFacing  int
		rotations    int
		wantFacing   int
	}{
		{
			name:        "rotate_once_from_south",
			startFacing: 0,
			rotations:   1,
			wantFacing:  1,
		},
		{
			name:        "rotate_full_circle",
			startFacing: 0,
			rotations:   8,
			wantFacing:  0,
		},
		{
			name:        "rotate_from_southeast",
			startFacing: 7,
			rotations:   1,
			wantFacing:  0,
		},
		{
			name:        "multiple_rotations",
			startFacing: 2,
			rotations:   3,
			wantFacing:  5,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p := &Player{
				Facing: tt.startFacing,
			}
			for i := 0; i < tt.rotations; i++ {
				p.Rotate()
			}
			if p.Facing != tt.wantFacing {
				t.Errorf("Player.Rotate() facing = %v, want %v", p.Facing, tt.wantFacing)
			}
		})
	}
}

func TestPlayer_UpdateAnimation(t *testing.T) {
	tests := []struct {
		name           string
		animating      bool
		initialTimer   float64
		initialFrame   int
		expectedTimer  float64
		expectedFrame  int
	}{
		{
			name:          "animation_timer_advances",
			animating:     true,
			initialTimer:  0,
			initialFrame:  0,
			expectedTimer: 1.0/60.0,
			expectedFrame: 0, // Frame doesn't change immediately
		},
		{
			name:          "not_animating_resets",
			animating:     false,
			initialTimer:  5,
			initialFrame:  2,
			expectedTimer: 0,
			expectedFrame: 0,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p := &Player{
				Animating:     tt.animating,
				WalkAnimTimer: tt.initialTimer,
				WalkAnimFrame: tt.initialFrame,
				X:             0,
				Y:             0,
				TargetX:       5, // Set target far enough to keep animating
				TargetY:       5,
				AnimationSpeed: 0.1, // Set a reasonable animation speed
			}
			p.UpdateAnimation()
			if p.WalkAnimTimer != tt.expectedTimer {
				t.Errorf("UpdateAnimation() timer = %v, want %v", p.WalkAnimTimer, tt.expectedTimer)
			}
			if p.WalkAnimFrame != tt.expectedFrame {
				t.Errorf("UpdateAnimation() frame = %v, want %v", p.WalkAnimFrame, tt.expectedFrame)
			}
		})
	}
}

func TestPlayer_KeepInBounds(t *testing.T) {
	type fields struct {
		X float64
		Y float64
	}
	type args struct {
		gridMovement bool
	}
	tests := []struct {
		name   string
		fields fields
		args   args
		wantX  float64
		wantY  float64
	}{
		{
			name:   "within_bounds_no_change",
			fields: fields{X: 5, Y: 5},
			args:   args{gridMovement: false},
			wantX:  5,
			wantY:  5,
		},
		{
			name:   "exceed_max_x",
			fields: fields{X: 15, Y: 5},
			args:   args{gridMovement: false},
			wantX:  11,
			wantY:  5,
		},
		{
			name:   "exceed_max_y",
			fields: fields{X: 5, Y: 15},
			args:   args{gridMovement: false},
			wantX:  5,
			wantY:  11,
		},
		{
			name:   "below_min_x",
			fields: fields{X: 0.5, Y: 5},
			args:   args{gridMovement: false},
			wantX:  1,
			wantY:  5,
		},
		{
			name:   "below_min_y",
			fields: fields{X: 5, Y: 0.5},
			args:   args{gridMovement: false},
			wantX:  5,
			wantY:  1,
		},
		{
			name:   "grid_movement_rounds",
			fields: fields{X: 5.7, Y: 5.3},
			args:   args{gridMovement: true},
			wantX:  6,
			wantY:  5,
		},
		{
			name:   "grid_movement_at_boundary",
			fields: fields{X: 11.4, Y: 11.6},
			args:   args{gridMovement: true},
			wantX:  11,
			wantY:  11,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p := &Player{
				X: tt.fields.X,
				Y: tt.fields.Y,
			}
			p.KeepInBounds(tt.args.gridMovement, nil)
			if p.X != tt.wantX {
				t.Errorf("KeepInBounds() X = %v, want %v", p.X, tt.wantX)
			}
			if p.Y != tt.wantY {
				t.Errorf("KeepInBounds() Y = %v, want %v", p.Y, tt.wantY)
			}
		})
	}
}

