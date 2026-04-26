package detection

import (
	"slices"

	"github.com/chewxy/math32"
	"github.com/oomph-ac/oomph/game"
	"github.com/oomph-ac/oomph/player"
	"github.com/sandertv/gophertunnel/minecraft/protocol/packet"
)

type AimA struct {
	mPlayer  *player.Player
	metadata *player.DetectionMetadata

	rotations     []float32
	rotationCount int
}

func New_AimA(p *player.Player) *AimA {
	return &AimA{
		mPlayer: p,
		metadata: &player.DetectionMetadata{
			FailBuffer:    5,
			MaxBuffer:     5,
			MaxViolations: 10,
			TrustDuration: -1,
		},
		rotations: make([]float32, 100),
	}
}

func (*AimA) Type() string {
	return TypeAim
}

func (*AimA) SubType() string {
	return "A"
}

func (*AimA) Description() string {
	return "Checks for an inconsistent difference between player rotations."
}

func (*AimA) Punishable() bool {
	return true
}

func (d *AimA) Metadata() *player.DetectionMetadata {
	return d.metadata
}

func (d *AimA) Detect(pk packet.Packet) {
	input, ok := pk.(*packet.PlayerAuthInput)
	if !ok {
		return
	}

	// This check will only apply to players rotating their camera with a mouse.
	if input.InputMode != packet.InputModeMouse {
		return
	}

	if d.mPlayer.Movement().XCollision() || d.mPlayer.Movement().ZCollision() || math32.Abs(d.mPlayer.Movement().Rotation().X()) >= 89 { // why does this always false ROTATION checks??!!!
		return
	}

	if d.mPlayer.Movement().TicksSinceTeleport() <= 1 {
		d.rotationCount = 0
		return
	}

	yawDelta := game.Round32(math32.Abs(d.mPlayer.Movement().RotationDelta().Z()), 5)
	if yawDelta < 1e-4 || yawDelta >= 180 {
		return
	}

	for _, r := range d.rotations {
		if math32.Abs(r-yawDelta) <= 1e-4 {
			return
		}
	}
	d.rotations[d.rotationCount] = yawDelta
	d.rotationCount++

	if d.rotationCount == 100 {
		var rotations = make([]float32, len(d.rotations))
		copy(rotations, d.rotations)
		slices.Sort(rotations)

		bSlope, matchAmt := d.determineBestSlope(rotations)
		d.mPlayer.Dbg.Notify(player.DebugModeRotations, true, "bestSlope=%f matchAmt=%d", bSlope, matchAmt)

		if matchAmt <= 5 {
			d.mPlayer.FailDetection(d, "bSl", game.Round32(bSlope, 5), "amt", matchAmt)
		} else {
			d.metadata.Buffer = 0
		}

		for i := 1; i < len(d.rotations); i++ {
			d.rotations[i-1] = d.rotations[i]
		}
		d.rotationCount--
	}
}

func (d *AimA) determineBestSlope(rotations []float32) (float32, int) {
	slopes := make([]float32, len(rotations)-2)
	for i := 0; i < len(rotations)-2; i++ {
		slopes[i] = rotations[i+1] - rotations[i]
	}
	slices.Sort(slopes)

	var (
		bestSlope    float32
		currentSlope float32 = math32.MaxFloat32 - 1

		bestCount, currentCount int
	)

	for _, slope := range slopes {
		if math32.Abs(slope-currentSlope) <= 1e-4 {
			currentCount++
		} else {
			currentSlope = slope
			currentCount = 1
		}

		if currentCount > bestCount {
			bestCount = currentCount
			bestSlope = currentSlope
		}
	}

	return bestSlope, bestCount
}
