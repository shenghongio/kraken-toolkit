package kubeadm

import "testing"

func TestRegisterJoinPhases(t *testing.T) {
	r := NewPhaseRegistry()
	RegisterJoinPhases(r)
	
	if len(r.List()) != 5 {
		t.Fatalf("expected 5 join phases, got %d", len(r.List()))
	}
	
	sorted, err := r.Sorted()
	if err != nil {
		t.Fatalf("sorted failed: %v", err)
	}
	if len(sorted) != 5 {
		t.Fatalf("expected 5 sorted, got %d", len(sorted))
	}
	if sorted[0].Name() != "preflight" {
		t.Errorf("expected first=preflight, got %s", sorted[0].Name())
	}
}

func TestJoinPhaseCommands(t *testing.T) {
	cfg := PhaseConfig{
		KubeadmBinary:     "/usr/bin/kubeadm",
		KubeadmConfigPath: "/tmp/kubeadm.yaml",
	}
	
	tests := []struct {
		phase    Phase
		expected []string
	}{
		{&JoinPreflightPhase{}, []string{"/usr/bin/kubeadm", "join", "phase", "preflight", "--config", "/tmp/kubeadm.yaml"}},
		{&ControlPlanePreparePhase{}, []string{"/usr/bin/kubeadm", "join", "phase", "control-plane-prepare", "all", "--config", "/tmp/kubeadm.yaml"}},
		{&JoinKubeletStartPhase{}, []string{"/usr/bin/kubeadm", "join", "phase", "kubelet-start", "--config", "/tmp/kubeadm.yaml"}},
		{&ControlPlaneJoinPhase{}, []string{"/usr/bin/kubeadm", "join", "phase", "control-plane-join", "all", "--config", "/tmp/kubeadm.yaml"}},
		{&JoinKubeconfigPhase{}, []string{"/usr/bin/kubeadm", "join", "phase", "kubeconfig", "--config", "/tmp/kubeadm.yaml"}},
	}
	
	for _, tt := range tests {
		got := tt.phase.Command(cfg)
		if len(got) != len(tt.expected) {
			t.Errorf("%s: expected %v, got %v", tt.phase.Name(), tt.expected, got)
			continue
		}
		for i := range got {
			if got[i] != tt.expected[i] {
				t.Errorf("%s[%d]: expected %q, got %q", tt.phase.Name(), i, tt.expected[i], got[i])
				break
			}
		}
	}
}
