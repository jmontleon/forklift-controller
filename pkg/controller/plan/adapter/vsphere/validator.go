package vsphere

import (
	liberr "github.com/konveyor/controller/pkg/error"
	api "github.com/konveyor/forklift-controller/pkg/apis/forklift/v1beta1"
	"github.com/konveyor/forklift-controller/pkg/apis/forklift/v1beta1/ref"
	"github.com/konveyor/forklift-controller/pkg/controller/provider/web"
	model "github.com/konveyor/forklift-controller/pkg/controller/provider/web/vsphere"
)

//
// vSphere validator.
type Validator struct {
	plan      *api.Plan
	inventory web.Client
}

//
// Load.
func (r *Validator) Load() (err error) {
	r.inventory, err = web.NewClient(r.plan.Referenced.Provider.Source)
	return
}

//
// Validate whether warm migration is supported from this provider type.
func (r *Validator) WarmMigration() (ok bool) {
	ok = true
	return
}

//
// Validate that a VM's networks have been mapped.
func (r *Validator) NetworksMapped(vmRef ref.Ref) (ok bool, err error) {
	if r.plan.Referenced.Map.Network == nil {
		return
	}
	vm := &model.VM{}
	err = r.inventory.Find(vm, vmRef)
	if err != nil {
		err = liberr.Wrap(
			err,
			"VM not found in inventory.",
			"vm",
			vmRef.String())
		return
	}

	for _, net := range vm.Networks {
		if !r.plan.Referenced.Map.Network.Status.Refs.Find(ref.Ref{ID: net.ID}) {
			return
		}
	}
	ok = true
	return
}

//
// Validate that a VM's disk backing storage has been mapped.
func (r *Validator) StorageMapped(vmRef ref.Ref) (ok bool, err error) {
	if r.plan.Referenced.Map.Storage == nil {
		return
	}
	vm := &model.VM{}
	err = r.inventory.Find(vm, vmRef)
	if err != nil {
		err = liberr.Wrap(
			err,
			"VM not found in inventory.",
			"vm",
			vmRef.String())
		return
	}

	for _, disk := range vm.Disks {
		if !r.plan.Referenced.Map.Storage.Status.Refs.Find(ref.Ref{ID: disk.Datastore.ID}) {
			return
		}
	}
	ok = true
	return
}

//
// Validate that a VM's Host isn't in maintenance mode.
func (r *Validator) MaintenanceMode(vmRef ref.Ref) (ok bool, err error) {
	vm := &model.VM{}
	err = r.inventory.Find(vm, vmRef)
	if err != nil {
		err = liberr.Wrap(
			err,
			"VM not found in inventory.",
			"vm",
			vmRef.String())
		return
	}

	host := &model.Host{}
	hostRef := ref.Ref{ID: vm.Host}
	err = r.inventory.Find(host, hostRef)
	if err != nil {
		err = liberr.Wrap(
			err,
			"Host not found in inventory.",
			"vm",
			vmRef.String(),
			"host",
			hostRef.String())
		return
	}

	ok = !host.InMaintenanceMode
	return
}
