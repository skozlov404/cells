/*
 * Copyright (c) 2018. Abstrium SAS <team (at) pydio.com>
 * This file is part of Pydio Cells.
 *
 * Pydio Cells is free software: you can redistribute it and/or modify
 * it under the terms of the GNU Affero General Public License as published by
 * the Free Software Foundation, either version 3 of the License, or
 * (at your option) any later version.
 *
 * Pydio Cells is distributed in the hope that it will be useful,
 * but WITHOUT ANY WARRANTY; without even the implied warranty of
 * MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
 * GNU Affero General Public License for more details.
 *
 * You should have received a copy of the GNU Affero General Public License
 * along with Pydio Cells.  If not, see <http://www.gnu.org/licenses/>.
 *
 * The latest code can be found at <https://pydio.com>.
 */

package service

import (
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"

	"github.com/hashicorp/go-version"
	"go.uber.org/zap"

	"github.com/pydio/cells/common"
	"github.com/pydio/cells/common/config"
	"github.com/pydio/cells/common/log"
	"github.com/pydio/cells/common/utils/migrations"
)

// ValidVersion creates a version.NewVersion ignoring the error.
func ValidVersion(v string) *version.Version {
	obj, _ := version.NewVersion(v)
	return obj
}

// FirstRun returns version "zero".
func FirstRun() *version.Version {
	obj, _ := version.NewVersion("0.0.0")
	return obj
}

// Latest retrieves current common Cells version.
func Latest() *version.Version {
	return common.Version()
}

// LastKnownVersion looks on this server if there was a previous version of this service
func LastKnownVersion(serviceName string) (v *version.Version, e error) {

	serviceDir, e := config.ServiceDataDir(serviceName)
	if e != nil {
		return nil, e
	}
	versionFile := filepath.Join(serviceDir, "version")

	data, err := ioutil.ReadFile(versionFile)
	if err != nil {
		if os.IsNotExist(err) {
			fake, _ := version.NewVersion("0.0.0")
			return fake, nil
		}
		return nil, err
	}
	return version.NewVersion(strings.TrimSpace(string(data)))

}

// UpdateVersion writes the version string to file
func UpdateVersion(serviceName string, v *version.Version) error {

	dir, err := config.ServiceDataDir(serviceName)
	if err != nil {
		return err
	}
	versionFile := filepath.Join(dir, "version")
	return ioutil.WriteFile(versionFile, []byte(v.String()), 0755)
}

// UpdateServiceVersion applies migration(s) if necessary and stores new current version for future use.
func UpdateServiceVersion(s Service) error {
	options := s.Options()
	newVersion, _ := version.NewVersion(options.Version)

	lastVersion, e := LastKnownVersion(options.Name)
	if e != nil {
		return e
	}

	writeVersion, err := migrations.Apply(options.Context, lastVersion, newVersion, options.Migrations)
	if writeVersion != nil {
		if e := UpdateVersion(options.Name, writeVersion); e != nil {
			log.Logger(options.Context).Error("could not write version file", zap.Error(e))
		}
	}
	return err

}
