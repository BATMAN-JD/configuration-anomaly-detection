package controller

import (
	"sync/atomic"

	"github.com/openshift/configuration-anomaly-detection/pkg/pagerduty"
)

// trackingPDClient wraps a pagerduty.Client and records whether it has
// already escalated the incident. A single instance is shared by the
// investigation pipeline (via incidentNotifier.AttachToBuilder), the action
// executor, and the controller's own fallback logic, so all three agree on
// whether the incident still needs a generic escalation.
type trackingPDClient struct {
	pagerduty.Client
	escalated atomic.Bool
}

func newTrackingPDClient(client pagerduty.Client) *trackingPDClient {
	return &trackingPDClient{Client: client}
}

func (t *trackingPDClient) EscalateIncident() error {
	err := t.Client.EscalateIncident()
	if err == nil {
		t.escalated.Store(true)
	}
	return err
}

func (t *trackingPDClient) EscalateIncidentWithNote(note string) error {
	err := t.Client.EscalateIncidentWithNote(note)
	if err == nil {
		t.escalated.Store(true)
	}
	return err
}

// HasEscalated reports whether this incident has already been escalated,
// through any path (investigation action, direct call, or note-attached
// escalation).
func (t *trackingPDClient) HasEscalated() bool {
	return t.escalated.Load()
}
