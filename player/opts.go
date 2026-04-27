package player

import "github.com/oomph-ac/oconfig"

type Opts struct {
	Combat   oconfig.CombatOpts
	Movement oconfig.MovementOpts
	Network  oconfig.NetworkOpts

	// LocalCombat carries combat tuning that is not (yet) part of oconfig.
	// These knobs let operators trade detection strictness for hit-registration
	// leniency, which is useful when players complain about hits being eaten by
	// the full server-authoritative combat gating.
	LocalCombat LocalCombatOpts
}

// LocalCombatOpts controls the server-authoritative combat component's
// hit-validation behaviour. All fields can be tuned at runtime; defaults
// from DefaultLocalCombatOpts preserve oomph's original behaviour.
type LocalCombatOpts struct {
	// FullAuthoritative gates attack packets behind the server-side validator.
	// When false, attacks always forward to the remote server and detections
	// (reach/hitbox/killaura on the client tracker) keep flagging — they just
	// no longer prevent the hit from registering. Default: true.
	FullAuthoritative bool
	// BBoxExpansion is the amount the targeted entity's bounding box is grown
	// by when raycasting. Larger values are more lenient toward edge-of-hitbox
	// hits. Default: 0.1.
	BBoxExpansion float32
	// MaximumReach is the maximum allowed distance for a valid survival hit.
	// Default: 2.9.
	MaximumReach float32
	// ReachLeniency is added to MaximumReach for the raycast pass only. It
	// absorbs ~1 frame of network jitter without weakening the raw-distance
	// reach detection. Default: 0.
	ReachLeniency float32
	// LerpSteps is the number of partial-tick samples taken between the
	// previous and current attack position when validating hits. Higher
	// values are more accurate at the cost of CPU. Default: 10.
	LerpSteps int
	// EntitySearchRadius is the radius the misprediction search uses when
	// the client swung in air but may have actually hit something. Default: 6.0.
	EntitySearchRadius float32
	// RawDistanceFallback accepts hits using the closest point on the entity's
	// bounding box (not just a successful raycast) for non-touch input modes,
	// gated by MaximumReach and MaximumAttackAngle. Touch always gets this
	// fallback. Default: false.
	RawDistanceFallback bool
	// BlockedByBlockCancel cancels hits whose ray passes through a solid
	// block. Set false to suppress this surgically — reach/angle gating
	// still applies. Useful when client/server block-state desync (e.g. a
	// recently-broken block) eats legit hits. Default: true.
	BlockedByBlockCancel bool
}

// DefaultLocalCombatOpts returns oomph's original combat tuning.
func DefaultLocalCombatOpts() LocalCombatOpts {
	return LocalCombatOpts{
		FullAuthoritative:    true,
		BBoxExpansion:        0.1,
		MaximumReach:         2.9,
		ReachLeniency:        0.0,
		LerpSteps:            10,
		EntitySearchRadius:   6.0,
		RawDistanceFallback:  false,
		BlockedByBlockCancel: true,
	}
}

func (p *Player) Opts() *Opts {
	return p.opts
}
