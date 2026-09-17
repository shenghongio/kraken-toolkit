package runtime

import (
	"context"
	"fmt"
	"github.com/kraken-pedestal/internal/basic/executor"
	"github.com/kraken-pedestal/internal/basic/inventory"
)

type Runtime struct {
	Inventory *inventory.Inventory
	Executor  *executor.Executor
}

func New(inv *inventory.Inventory, exec *executor.Executor) *Runtime {
	return &Runtime{
		Inventory: inv,
		Executor:  exec,
	}
}

func (r *Runtime) Shell(ctx context.Context, command string) []executor.Result {
	if r == nil {
		return []executor.Result{
			{
				Error: fmt.Errorf("runtime is nil"),
			},
		}
	}
	if r.Inventory == nil {
		return []executor.Result{
			{
				Error: fmt.Errorf("inventory is nil"),
			},
		}
	}
	if r.Executor == nil {
		return []executor.Result{
			{
				Error: fmt.Errorf("executor is nil"),
			},
		}
	}

	hosts := r.Inventory.Hosts()
	return r.Executor.Shell(ctx, hosts, command)
}
