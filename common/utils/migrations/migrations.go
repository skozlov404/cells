package migrations

import (
	"context"

	"github.com/hashicorp/go-version"
)

// Migration defines a target version and functions to upgrade and/or downgrade.
type Migration struct {
	TargetVersion *version.Version
	Up            func(ctx context.Context) error
	Down          func(ctx context.Context) error
}

// FirstRun returns version "zero".
func FirstRun() *version.Version {
	obj, _ := version.NewVersion("0.0.0")
	return obj
}

// ApplyMigrations browse migrations upward on downward and apply them sequentially. It returns a version to be
// saved as the current valid version of the service, or nil if no changes were necessary. In specific case where
// current version is 0.0.0 (first run), it only applies first run migration (if any) and returns target version.
func Apply(ctx context.Context, current *version.Version, target *version.Version, migrations []*Migration) (*version.Version, error) {

	if target.Equal(current) {
		return nil, nil
	}

	if migrations == nil {
		return target, nil
	}

	// corner case of the fresh install, returns the current target version to be stored
	// if current.Equal(FirstRun()) {
	// 	m := migrations[0]
	// 	fmt.Println(m.TargetVersion.Equal(FirstRun()))

	// 	// Double check to insure we really only perform FirstRun initialisation
	// 	if !m.TargetVersion.Equal(FirstRun()) {

	// 		// no first run init, doing nothing
	// 		return target, nil
	// 	}

	// 	//log.Logger(ctx).Debug(fmt.Sprintf("About to initialise service at version %s", target.String()))
	// 	err := m.Up(ctx)
	// 	if err != nil {
	// 		//log.Logger(ctx).Error(fmt.Sprintf("could not initialise service at version %s", target.String()), zap.Error(err))
	// 		return current, err
	// 	}
	// 	return target, nil
	// }

	//log.Logger(ctx).Debug(fmt.Sprintf("About to perform migration from %s to %s", current.String(), target.String()))

	if target.GreaterThan(current) {
		var successVersion *version.Version
		for _, migration := range migrations {
			v := migration.TargetVersion
			if migration.Up != nil && (current.String() == "0.0.0" || v.GreaterThan(current)) && (v.LessThan(target) || v.Equal(target)) {
				if err := migration.Up(ctx); err != nil {
					return successVersion, err
				}
				successVersion, _ = version.NewVersion(v.String())
			}
		}
	}

	if target.LessThan(current) {

		var successVersion *version.Version
		for i := len(migrations) - 1; i >= 0; i-- {
			migration := migrations[i]
			v := migration.TargetVersion
			if migration.Down != nil && v.GreaterThan(target) && (v.LessThan(current) || v.Equal(current)) {
				if err := migration.Down(ctx); err != nil {
					return successVersion, err
				}
				successVersion, _ = version.NewVersion(v.String())
			}
		}

	}

	return target, nil
}
