package entry

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/ftl/hellocontest/core"
)

func newTestClaims(vfo1Claimed, vfo2Claimed core.QSONumber) SerialClaims {
	s := newSerialClaims()
	s.claimed[core.VFO1] = vfo1Claimed
	s.claimed[core.VFO2] = vfo2Claimed
	return s
}

// --- nextUnclaimed ---

func TestSerialClaims_nextUnclaimed_noConflict(t *testing.T) {
	s := newSerialClaims()
	assert.Equal(t, core.QSONumber(5), s.nextUnclaimed(core.VFO1, 5))
	assert.Equal(t, core.QSONumber(5), s.nextUnclaimed(core.VFO2, 5))
}

func TestSerialClaims_nextUnclaimed_otherVFOClaimedSameBase_skips(t *testing.T) {
	tt := []struct {
		name    string
		forVFO  core.VFOID
		other   core.VFOID
	}{
		{"VFO1 skips VFO2", core.VFO1, core.VFO2},
		{"VFO2 skips VFO1", core.VFO2, core.VFO1},
	}
	for _, tc := range tt {
		t.Run(tc.name, func(t *testing.T) {
			s := newSerialClaims()
			s.claimed[tc.other] = 5
			assert.Equal(t, core.QSONumber(6), s.nextUnclaimed(tc.forVFO, 5))
		})
	}
}

func TestSerialClaims_nextUnclaimed_otherVFOClaimedDifferentBase_noSkip(t *testing.T) {
	s := newTestClaims(0, 3) // VFO2 claimed 3, base is 5
	assert.Equal(t, core.QSONumber(5), s.nextUnclaimed(core.VFO1, 5))
}

func TestSerialClaims_nextUnclaimed_otherVFOUnclaimed_noSkip(t *testing.T) {
	s := newSerialClaims() // both unclaimed
	assert.Equal(t, core.QSONumber(5), s.nextUnclaimed(core.VFO1, 5))
	assert.Equal(t, core.QSONumber(5), s.nextUnclaimed(core.VFO2, 5))
}

// --- DisplayedSerial ---

func TestSerialClaims_DisplayedSerial_noClaim_returnsBase(t *testing.T) {
	s := newSerialClaims()
	assert.Equal(t, core.QSONumber(7), s.DisplayedSerial(core.VFO1, 7))
	assert.Equal(t, core.QSONumber(7), s.DisplayedSerial(core.VFO2, 7))
}

func TestSerialClaims_DisplayedSerial_noClaim_otherConflicts_skips(t *testing.T) {
	s := newTestClaims(0, 7) // VFO2 holds 7
	// VFO1 has no claim; VFO2 claimed base → VFO1 must skip
	assert.Equal(t, core.QSONumber(8), s.DisplayedSerial(core.VFO1, 7))
}

func TestSerialClaims_DisplayedSerial_hasClaim_returnsClaimed(t *testing.T) {
	s := newTestClaims(42, 0)
	// VFO1's claimed value is returned regardless of base or other VFO
	assert.Equal(t, core.QSONumber(42), s.DisplayedSerial(core.VFO1, 99))
}

func TestSerialClaims_DisplayedSerial_bothClaimed_eachReturnsOwn(t *testing.T) {
	s := newTestClaims(10, 11)
	assert.Equal(t, core.QSONumber(10), s.DisplayedSerial(core.VFO1, 10))
	assert.Equal(t, core.QSONumber(11), s.DisplayedSerial(core.VFO2, 10))
}

// --- AllDisplayed ---

func TestSerialClaims_AllDisplayed_length(t *testing.T) {
	s := newSerialClaims()
	result := s.AllDisplayed(1)
	assert.Len(t, result, int(core.VFOCount))
}

func TestSerialClaims_AllDisplayed_noClaims_bothReturnBase(t *testing.T) {
	s := newSerialClaims()
	result := s.AllDisplayed(3)
	assert.Equal(t, core.QSONumber(3), result[core.VFO1])
	assert.Equal(t, core.QSONumber(3), result[core.VFO2])
}

func TestSerialClaims_AllDisplayed_VFO1Claimed_VFO2SkipsIfConflict(t *testing.T) {
	s := newTestClaims(5, 0) // VFO1 holds 5
	result := s.AllDisplayed(5)
	assert.Equal(t, core.QSONumber(5), result[core.VFO1]) // own claim
	assert.Equal(t, core.QSONumber(6), result[core.VFO2]) // bumped
}

func TestSerialClaims_AllDisplayed_VFO2Claimed_VFO1SkipsIfConflict(t *testing.T) {
	s := newTestClaims(0, 5) // VFO2 holds 5
	result := s.AllDisplayed(5)
	assert.Equal(t, core.QSONumber(6), result[core.VFO1]) // bumped
	assert.Equal(t, core.QSONumber(5), result[core.VFO2]) // own claim
}

func TestSerialClaims_AllDisplayed_bothClaimed_noInteraction(t *testing.T) {
	s := newTestClaims(5, 6)
	result := s.AllDisplayed(5)
	assert.Equal(t, core.QSONumber(5), result[core.VFO1])
	assert.Equal(t, core.QSONumber(6), result[core.VFO2])
}

// --- ClaimNext ---

func TestSerialClaims_ClaimNext_setsClaimedAndSnapshot(t *testing.T) {
	s := newSerialClaims()
	s.ClaimNext(core.VFO1, 5)
	assert.Equal(t, core.QSONumber(5), s.claimed[core.VFO1])
	assert.Equal(t, core.QSONumber(5), s.snapshot[core.VFO1])
}

func TestSerialClaims_ClaimNext_otherVFOConflict_bumps(t *testing.T) {
	s := newTestClaims(0, 5) // VFO2 holds 5
	s.ClaimNext(core.VFO1, 5)
	assert.Equal(t, core.QSONumber(6), s.claimed[core.VFO1])
	assert.Equal(t, core.QSONumber(5), s.snapshot[core.VFO1])
}

func TestSerialClaims_ClaimNext_sticky_secondCallNoOp(t *testing.T) {
	s := newSerialClaims()
	s.ClaimNext(core.VFO1, 5)
	s.ClaimNext(core.VFO1, 9) // base advanced; must not change claim
	assert.Equal(t, core.QSONumber(5), s.claimed[core.VFO1])
	assert.Equal(t, core.QSONumber(5), s.snapshot[core.VFO1])
}

func TestSerialClaims_ClaimNext_bothVFOs_collisionAvoided(t *testing.T) {
	s := newSerialClaims()
	s.ClaimNext(core.VFO1, 5) // VFO1 gets 5
	s.ClaimNext(core.VFO2, 5) // VFO2 must bump to 6
	assert.Equal(t, core.QSONumber(5), s.claimed[core.VFO1])
	assert.Equal(t, core.QSONumber(6), s.claimed[core.VFO2])
}

func TestSerialClaims_ClaimNext_doesNotAffectOtherVFO(t *testing.T) {
	s := newSerialClaims()
	s.ClaimNext(core.VFO1, 5)
	assert.Equal(t, core.QSONumber(0), s.claimed[core.VFO2])
	assert.Equal(t, core.QSONumber(0), s.snapshot[core.VFO2])
}

// --- Release ---

func TestSerialClaims_Release_clearsClaim(t *testing.T) {
	s := newTestClaims(5, 0)
	s.snapshot[core.VFO1] = 5
	s.Release(core.VFO1)
	assert.Equal(t, core.QSONumber(0), s.claimed[core.VFO1])
	assert.Equal(t, core.QSONumber(0), s.snapshot[core.VFO1])
}

func TestSerialClaims_Release_doesNotAffectOtherVFO(t *testing.T) {
	s := newTestClaims(5, 7)
	s.snapshot[core.VFO2] = 7
	s.Release(core.VFO1)
	assert.Equal(t, core.QSONumber(7), s.claimed[core.VFO2])
	assert.Equal(t, core.QSONumber(7), s.snapshot[core.VFO2])
}

func TestSerialClaims_Release_allowsReClaim(t *testing.T) {
	s := newSerialClaims()
	s.ClaimNext(core.VFO1, 5)
	s.Release(core.VFO1)
	s.ClaimNext(core.VFO1, 9) // now free; should claim 9
	assert.Equal(t, core.QSONumber(9), s.claimed[core.VFO1])
	assert.Equal(t, core.QSONumber(9), s.snapshot[core.VFO1])
}
