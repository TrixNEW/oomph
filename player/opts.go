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
	//
	// The zero value (LocalCombatOpts{}) is NOT equivalent to the in-tree
	// defaults and is unsafe for production use: BBoxExpansion=0 collapses the
	// raycast bbox to exact size (stricter — eats more legit hits), and
	// EntitySearchRadius=0 DISABLES misprediction-entity resolution, which is
	// a real anti-cheat regression (a killaura masking attacks as air-swings
	// would not be matched to its target). All Disable* flags and
	// RawDistanceFallback default to false, so the boolean side of the zero
	// value is safe; the numeric side is not.
	//
	// player.New() initialises this field via DefaultLocalCombatOpts(); any
	// caller constructing player.Opts{} directly (tests, embedding
	// integrations) MUST seed LocalCombat from DefaultLocalCombatOpts() and
	// then tweak. See the per-field docs for which sentinel triggers a
	// fallback to the original constant on each numeric knob.
	LocalCombat LocalCombatOpts
}

// LocalCombatOpts controls the server-authoritative combat component's
// hit-validation behaviour.
//
// The zero value is unsafe for production: see Opts.LocalCombat for the full
// rationale. Construct via DefaultLocalCombatOpts() and tweak from there;
// per-field docs document which sentinel falls back to the original constant
// (some fields treat 0 as a legitimate explicit value, not a fallback).
type LocalCombatOpts struct {
	// DisableFullAuthoritative removes server-side gating of attack packets.
	// When true, attacks always forward to the remote server and detections
	// (reach/hitbox/killaura on the client tracker) keep flagging — they just
	// no longer prevent the hit from registering. Default: false (gating on).
	DisableFullAuthoritative bool
	// DisableBlockOcclusionCheck disables the wall-occlusion check that
	// cancels hits whose ray passes through a solid block. Reach/angle gating
	// still applies. Useful when client/server block-state desync (e.g. a
	// recently-broken block) eats legit hits. Default: false (check on).
	DisableBlockOcclusionCheck bool
	// RawDistanceFallback also accepts the closest-point-on-bbox distance
	// (not just a successful raycast) for non-touch input modes when no
	// raycast lands, gated by MaximumReach and MaximumAttackAngle. Touch
	// always gets this fallback regardless of this flag. Default: false.
	//
	// NOTE: this is a meaningful weakening of reach detection — it accepts
	// any in-cone, in-reach hit even when no raycast intersects the bbox.
	// It is not just a recovery path for narrow lerp aliasing.
	RawDistanceFallback bool

	// BBoxExpansion is the amount the targeted entity's bounding box is grown
	// by when raycasting. Larger values are more lenient toward edge-of-hitbox
	// hits. Set to a negative value to keep the default of 0.1; 0 is honoured
	// as "no growth, exact bbox".
	BBoxExpansion float32
	// MaximumReach is the maximum allowed distance for a valid survival hit.
	// Defaults to 2.9 when unset (<= 0). 0 is treated as unset because a
	// zero reach makes no hits ever land.
	MaximumReach float32
	// ReachLeniency is added to MaximumReach for the raycast pass only. It
	// absorbs ~1 frame of network jitter without weakening the raw-distance
	// reach detection. Default: 0 (no extra leniency).
	ReachLeniency float32
	// LerpSteps is the number of partial-tick samples taken between the
	// previous and current attack position when validating hits. Higher
	// values are more accurate at the cost of CPU. Defaults to 10 when unset
	// (<= 0). 0 is treated as unset because a zero step count is divide-by-zero.
	//
	// NOTE: this value is captured once at construction (player creation) so
	// that pre-allocated result slices stay correctly sized. Runtime changes
	// take effect on the next player session.
	LerpSteps int
	// EntitySearchRadius is the radius the misprediction search uses when
	// the client swung in air but may have actually hit something. Set to a
	// negative value to keep the default of 6.0; 0 is honoured as "do not
	// search" (all swings in air will be treated as mispredictions resolved
	// to no entity).
	EntitySearchRadius float32
}

// DefaultLocalCombatOpts returns oomph's original strict combat tuning as an
// explicit starting point.
//
// The zero value of LocalCombatOpts preserves anti-cheat strictness (all
// Disable* flags off, RawDistanceFallback off) but is NOT numerically
// equivalent to DefaultLocalCombatOpts(): several numeric fields treat 0 as a
// legitimate explicit value rather than a "use default" sentinel — see the
// per-field docs. In particular, the zero value yields BBoxExpansion=0 (exact
// bbox, no growth) and EntitySearchRadius=0 (misprediction search disabled).
// Callers constructing player.Opts{} directly should start from this helper
// to match in-tree defaults.
func DefaultLocalCombatOpts() LocalCombatOpts {
	return LocalCombatOpts{
		BBoxExpansion:      0.1,
		MaximumReach:       2.9,
		ReachLeniency:      0.0,
		LerpSteps:          10,
		EntitySearchRadius: 6.0,
	}
}

func (p *Player) Opts() *Opts {
	return p.opts
}
