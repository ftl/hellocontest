package entry

import "github.com/ftl/hellocontest/core"

// SerialClaims holds per-VFO serial-number reservations and provides
// collision-avoidance between VFOs. A claim transitions through three states:
//   - unclaimed (claimed[vfo] == 0): no serial reserved
//   - claimed (claimed[vfo] != 0, committed[vfo] == false): serial reserved, recyclable on release
//   - committed (claimed[vfo] != 0, committed[vfo] == true): serial sent over air (UI-only distinction)
//
// Both claimed and committed serials are recyclable on release. The committed
// flag is used only for UI visualization (bold serial label).
type SerialClaims struct {
	claimed   []core.QSONumber // per-VFO reserved serial; 0 = unclaimed
	snapshot  []core.QSONumber // logbook.NextQSONumber() value at claim time
	committed []bool           // per-VFO: serial was sent over air
}

func newSerialClaims() SerialClaims {
	return SerialClaims{
		claimed:   make([]core.QSONumber, core.VFOCount),
		snapshot:  make([]core.QSONumber, core.VFOCount),
		committed: make([]bool, core.VFOCount),
	}
}

// nextUnclaimed returns the first serial for forVFO not already claimed by the other VFO.
func (s *SerialClaims) nextUnclaimed(forVFO core.VFOID, base core.QSONumber) core.QSONumber {
	otherVFO := core.VFO1
	if forVFO == core.VFO1 {
		otherVFO = core.VFO2
	}
	if s.claimed[otherVFO] != 0 && s.claimed[otherVFO] == base {
		return base + 1
	}
	return base
}

// DisplayedSerial returns the serial currently visible to the operator on vfo's row:
// the claimed value if any, otherwise the next unclaimed preview.
func (s *SerialClaims) DisplayedSerial(vfo core.VFOID, base core.QSONumber) core.QSONumber {
	if s.claimed[vfo] != 0 {
		return s.claimed[vfo]
	}
	return s.nextUnclaimed(vfo, base)
}

// AllDisplayed returns the displayed serial for every VFO, in VFO index order.
func (s *SerialClaims) AllDisplayed(base core.QSONumber) []core.QSONumber {
	out := make([]core.QSONumber, len(s.claimed))
	for vfo := range len(s.claimed) {
		out[vfo] = s.DisplayedSerial(core.VFOID(vfo), base)
	}
	return out
}

// ClaimNext reserves the next unclaimed serial for vfo using base as the logbook reference.
// Sticky: subsequent calls while a claim exists are no-ops.
func (s *SerialClaims) ClaimNext(vfo core.VFOID, base core.QSONumber) {
	if s.claimed[vfo] != 0 {
		return
	}
	s.claimed[vfo] = s.nextUnclaimed(vfo, base)
	s.snapshot[vfo] = base
	s.committed[vfo] = false
}

// Commit marks the claim for vfo as committed (serial was sent over air).
// No-op if vfo has no claim.
func (s *SerialClaims) Commit(vfo core.VFOID) {
	if s.claimed[vfo] == 0 {
		return
	}
	s.committed[vfo] = true
}

// Release frees the claim slot for vfo.
func (s *SerialClaims) Release(vfo core.VFOID) {
	s.claimed[vfo] = 0
	s.snapshot[vfo] = 0
	s.committed[vfo] = false
}
