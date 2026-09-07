package controller

import (
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"

	pdmock "github.com/openshift/configuration-anomaly-detection/pkg/pagerduty/mock"
)

// TestTrackingPDClient guards against CAD double-escalating PagerDuty
// incidents (ROSAENG-66516): every code path that might escalate an incident
// (an investigation's direct call, an investigation's action, or the
// controller's own generic fallback) goes through the same trackingPDClient,
// so HasEscalated() is the single source of truth for "has this incident
// already been escalated".
func TestTrackingPDClient(t *testing.T) {
	t.Run("starts unescalated", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		tracked := newTrackingPDClient(pdmock.NewMockClient(ctrl))

		assert.False(t, tracked.HasEscalated())
	})

	t.Run("EscalateIncident marks it escalated", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		mockClient := pdmock.NewMockClient(ctrl)
		mockClient.EXPECT().EscalateIncident().Return(nil)

		tracked := newTrackingPDClient(mockClient)
		require.NoError(t, tracked.EscalateIncident())
		assert.True(t, tracked.HasEscalated())
	})

	t.Run("EscalateIncidentWithNote marks it escalated", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		mockClient := pdmock.NewMockClient(ctrl)
		mockClient.EXPECT().EscalateIncidentWithNote("reason").Return(nil)

		tracked := newTrackingPDClient(mockClient)
		require.NoError(t, tracked.EscalateIncidentWithNote("reason"))
		assert.True(t, tracked.HasEscalated())
	})

	t.Run("a failed escalation is not tracked", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		mockClient := pdmock.NewMockClient(ctrl)
		mockClient.EXPECT().EscalateIncident().Return(errors.New("pagerduty unavailable"))

		tracked := newTrackingPDClient(mockClient)
		assert.Error(t, tracked.EscalateIncident())
		assert.False(t, tracked.HasEscalated())
	})

	t.Run("second escalation attempt is still visible as already escalated", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		mockClient := pdmock.NewMockClient(ctrl)
		mockClient.EXPECT().EscalateIncident().Return(nil)

		tracked := newTrackingPDClient(mockClient)
		require.NoError(t, tracked.EscalateIncident())

		// Simulates pagerduty.go's fallback check: a caller that consults
		// HasEscalated() first must not issue a second EscalateIncident call.
		assert.True(t, tracked.HasEscalated())
	})
}
