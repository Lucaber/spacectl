package spacefile

import (
	"fmt"

	"github.com/hashicorp/go-multierror"
)

type StageDef struct {
	Name         string          `hcl:",key"`
	Inherit      string          `hcl:"inherit"`
	Applications SoftwareDefList `hcl:"application"`
	Databases    SoftwareDefList `hcl:"database"`
	Cronjobs     CronjobDefList  `hcl:"cron"`

	inheritStage *StageDef
}

func (d *StageDef) Validate() error {
	var err *multierror.Error

	if len(d.Applications) > 1 {
		err = multierror.Append(err, fmt.Errorf("Stage '%s' shoud not contain more than one application", d.Name))
	}

	return err.ErrorOrNil()
}

// Application returns a reference to the one application defined for this stage
func (d *StageDef) Application() *SoftwareDef {
	for i := range d.Applications {
		app := d.Applications[i]
		return &app
	}

	return nil
}

func (d *StageDef) resolveUserData() error {
	var mErr *multierror.Error
	var err error

	for i := range d.Applications {
		d.Applications[i].UserData, err = unfuckHCL(d.Applications[i].UserData, "")
		mErr = multierror.Append(mErr, err)
	}

	return mErr.ErrorOrNil()
}

func (d *StageDef) resolveInheritance(level int) error {
	if level > 4 {
		return fmt.Errorf("Could not resolve stage dependencies after %d levels. Please check that there is no cyclic inheritance", level)
	}

	if d.inheritStage == nil {
		return nil
	}

	err := d.inheritStage.resolveInheritance(level + 1)
	if err != nil {
		return err
	}

	originalName := d.Name

	d.Applications, err = d.Applications.Merge(d.inheritStage.Applications)
	if err != nil {
		return err
	}

	d.Databases, err = d.Databases.Merge(d.inheritStage.Databases)
	if err != nil {
		return err
	}

	d.Name = originalName
	d.inheritStage = nil

	return nil
}