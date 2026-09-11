package inventory

import (
	"os"
	"path/filepath"
	"reflect"
	"testing"
)

func TestParseFile(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "hosts")
	content := `
# inventory example

[web]
192.168.1.10
192.168.1.11 user=admin
192.168.1.12 user=test port=2222 password=123456

[db]
192.168.1.20 user=root port=2200
`
	
	if err := os.WriteFile(path, []byte(content), 0600); err != nil {
		t.Fatal(err)
	}
	host, err := ParseFile(path)
	if err != nil {
		t.Fatalf("ParseFile() error: = %v", err)
	}
	if len(host) != 4 {
		t.Fatalf("expected 4 hosts got %d", len(host))
	}
	
	tests := []struct {
		name    string
		host    Host
		address string
		user    string
		port    int
		passwd  string
		group   string
	}{
		{
			name:    "default",
			host:    host[0],
			address: "192.168.1.10",
			user:    "root",
			port:    22,
			group:   "web",
		},
		{
			name:    "custom user",
			host:    host[1],
			address: "192.168.1.11",
			user:    "admin",
			port:    22,
			group:   "web",
		},
		{
			name:    "custom ssh config",
			host:    host[2],
			address: "192.168.1.12",
			user:    "test",
			port:    2222,
			passwd:  "123456",
			group:   "web",
		},
		{
			name:    "db",
			host:    host[3],
			address: "192.168.1.20",
			user:    "root",
			port:    2200,
			group:   "db",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.host.Address != tt.address {
				t.Errorf("Address %s, want %s", tt.host.Address, tt.address)
			}
			if tt.host.User != tt.user {
				t.Errorf("User %s, want %s", tt.host.User, tt.user)
			}
			if tt.host.Port != tt.port {
				t.Errorf("Port %d, want %d", tt.host.Port, tt.port)
			}
			if tt.host.Passwd != tt.passwd {
				t.Errorf("Passwd %s, want %s", tt.host.Passwd, tt.passwd)
			}
			if tt.host.Group != tt.group {
				t.Errorf("Group %s, want %s", tt.host.Group, tt.group)
			}
		})
	}
}

func TestParseFile_InvalidParamenter(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "hosts")
	content := `
[web]
192.168.1.10 invalid
`
	if err := os.WriteFile(path, []byte(content), 0600); err != nil {
	}
	_, err := ParseFile(path)
	if err == nil {
		t.Fatal("expected error,got nil")
	}
}

func TestParseFile_InvalidPort(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "hosts")
	content := `
[web]
192.168.1.10 port=777777
`
	if err := os.WriteFile(path, []byte(content), 0600); err != nil {
		t.Fatal(err)
	}
	_, err := ParseFile(path)
	if err == nil {
		t.Fatal("expected error, got nil")
	}
}
func TestInventoryResolveGroup(t *testing.T) {
	inv := &Inventory{
		hosts: []Host{
			{
				Address: "192.168.1.10",
				User:    "root",
				Port:    22,
				Group:   "web",
			},
			{
				Address: "192.168.1.11",
				User:    "admin",
				Port:    2200,
				Group:   "db",
			},
			{
				Address: "192.168.1.12",
				User:    "test",
				Port:    22,
				Group:   "web",
			},
		},
	}
	hosts, err := inv.Resolve([]string{"web"})
	if err != nil {
		t.Fatalf("Resolve() error: = %v", err)
	}
	if len(hosts) != 2 {
		t.Fatalf("expected 2 hosts got %d", len(hosts))
	}
	if hosts[0].Address != "192.168.1.10" {
		t.Errorf("hosts[0].Address = %s, want %s", hosts[0].Address, "192.168.1.10")
	}
}

func TestInventoryResolveHost(t *testing.T) {
	inv := &Inventory{
		hosts: []Host{
			{
				Address: "10.0.0.1",
				User:    "root",
				Port:    22,
				Group:   "web",
			},
			{
				Address: "10.0.0.2",
				User:    "root",
				Port:    22,
				Group:   "db",
			},
		},
	}
	hosts, err := inv.Resolve([]string{"10.0.0.2"})
	if err != nil {
		t.Fatalf("Resolve() error: = %v", err)
	}
	want := []Host{
		{
			Address: "10.0.0.2",
			User:    "root",
			Port:    22,
			Group:   "db",
		},
	}
	if !reflect.DeepEqual(hosts, want) {
		t.Fatalf("ResolveHost() hosts = %v, want %v", hosts, want)
	}
}
func TestInventoryResolveAll(t *testing.T) {
	inv := &Inventory{
		hosts: []Host{
			{
				Address: "192.2.2.1",
				User:    "root",
				Port:    22,
			},
			{
				Address: "192.2.2.2",
				User:    "root",
				Port:    22,
			},
		},
	}
	hosts, err := inv.Resolve([]string{"all"})
	if err != nil {
		t.Fatalf("Resolve() error: = %v", err)
	}
	if len(hosts) != 2 {
		t.Fatalf("expected 2 hosts got %d", len(hosts))
	}
}
func TestInventoryResolveAndDeduplicate(t *testing.T) {
	inv := &Inventory{
		hosts: []Host{
			{
				Address: "192.2.2.1",
				User:    "root",
				Port:    22,
				Group:   "web",
			},
			{
				Address: "192.2.2.2",
				User:    "root",
				Port:    22,
				Group:   "web",
			},
		},
	}
	hosts, err := inv.Resolve([]string{"web", "192.2.2.1"})
	if err != nil {
		t.Fatalf("Resolve() error: = %v", err)
	}
	if len(hosts) != 2 {
		t.Fatalf("expected 2 hosts got %d", len(hosts))
	}
}
func TestDeduplicateDifferentSSHUsers(t *testing.T) {
	hosts := []Host{
		{
			Address: "10.0.0.1",
			User:    "root",
			Port:    22,
		},
		{
			Address: "10.0.0.1",
			User:    "admin",
			Port:    22,
		},
		{
			Address: "10.0.0.1",
			User:    "root",
			Port:    22,
		},
	}
	
	result := Deduplicate(hosts)
	
	if len(result) != 2 {
		t.Fatalf(
			"expected 2 hosts, got %d",
			len(result),
		)
	}
	
	if result[0].User != "root" {
		t.Errorf(
			"result[0].User = %q",
			result[0].User,
		)
	}
	
	if result[1].User != "admin" {
		t.Errorf(
			"result[1].User = %q",
			result[1].User,
		)
	}
}

func TestInventoryResolveUnknownTarget(t *testing.T) {
	inv := &Inventory{
		hosts: []Host{
			{
				Address: "10.0.0.1",
				User:    "root",
				Port:    22,
				Group:   "web",
			},
		},
	}
	
	_, err := inv.Resolve([]string{"unknown"})
	if err == nil {
		t.Fatal("expected error, got nil")
	}
}

func TestInventoryResolveEmptyTargets(t *testing.T) {
	inv := &Inventory{
		hosts: []Host{
			{
				Address: "10.0.0.1",
				User:    "root",
				Port:    22,
			},
		},
	}
	
	_, err := inv.Resolve(nil)
	if err == nil {
		t.Fatal("expected error, got nil")
	}
}

func TestInventoryHostsReturnsCopy(t *testing.T) {
	inv := &Inventory{
		hosts: []Host{
			{
				Address: "10.0.0.1",
				User:    "root",
				Port:    22,
			},
		},
	}
	
	hosts := inv.Hosts()
	
	hosts[0].Address = "changed"
	
	if inv.hosts[0].Address != "10.0.0.1" {
		t.Fatal("Hosts() returned internal slice")
	}
}
