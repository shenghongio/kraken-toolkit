package kubeadm

import "testing"

func TestRegisterInitPhases(t *testing.T) {
	reg := NewPhaseRegistry()
	RegisterInitPhases(reg)
	if len(reg.List()) != 12 {
		t.Fatalf("RegisterInitPhases failed, expect 12 phases, got %d", len(reg.List()))
	}
	sorted, err := reg.Sorted()
	if err != nil {
		t.Fatalf("sorted failed: %v", err)
	}
	if len(sorted) != 12 {
		t.Fatalf("expected 12 phases, got %d", len(sorted))
	}
	if sorted[0].Name() != "preflight" {
		t.Fatalf("expected phase 'preflight', got %s", sorted[0].Name())
	}
	if sorted[11].Name() != "addon" {
		t.Fatalf("expected phase 'addon', got %s", sorted[11].Name())
	}
	
}
func TestInitPhaseCommands(t *testing.T) {
	phaseConfig := PhaseConfig{
		KubeadmBinary:     "/usr/bin/kubeadm",
		KubeadmConfigPath: "./kubeadm-config.yaml",
	}
	tests := []struct {
		phase    Phase
		expected []string
	}{
		{&PreflightPhase{}, []string{"/usr/bin/kubeadm", "init", "phase", "preflight", "--config", "./kubeadm-config.yaml"}},
		{&CertsPhase{}, []string{"/usr/bin/kubeadm", "init", "phase", "certs", "all", "--config", "./kubeadm-config.yaml"}},
		{&AddonPhase{}, []string{"/usr/bin/kubeadm", "init", "phase", "addon", "all", "all", "--config", "./kubeadm-config.yaml"}},
	}
	for _, tt := range tests {
		command := tt.phase.Command(phaseConfig)
		if len(command) != len(tt.expected) {
			t.Errorf("%v: expected %v  got %v", tt.phase.Name(), tt.expected, command)
			continue
		}
		for i := range command {
			if command[i] != tt.expected[i] {
				t.Errorf("%s[%d]: expected %s, got %s", tt.phase.Name(), i, tt.expected[i], command[i])
				break
			}
		}
	}
}
