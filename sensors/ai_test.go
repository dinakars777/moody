package sensors

import "testing"

func TestAIStopIsIdempotent(t *testing.T) {
	ai := NewAI(false)
	ai.Stop()
	ai.Stop()
}
