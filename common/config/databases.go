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

package config

import (
	"fmt"
	"log"

	"github.com/pydio/cells/common"
	"github.com/pydio/cells/common/utils/std"
	"github.com/spf13/cast"
)

func getDatabaseFromRef(name string) (string, string, error) {
	var databases map[string]*std.Database
	err := Values("databases").Scan(&databases)
	if err != nil {
		log.Fatal(err)
	}

	if v, ok := databases[name]; ok {
		return v.Driver, v.DSN, nil
	}

	return "", "", fmt.Errorf("not found")
}

// GetDatabase retrieves the database data from the config
func GetDatabase(conf common.ConfigValues) (string, string) {
	fmt.Println(conf.Get())
	switch v := conf.Get().(type) {
	case string:
		drv, dsn, err := getDatabaseFromRef(v)
		if err != nil {
			break
		}

		return drv, dsn
	default:
		m, err := cast.ToStringMapStringE(v)
		if err != nil {
			break
		}

		return m["drv"], m["dsn"]
	}

	defaultDBKey := Values("defaults").String("database", "")
	drv, dsn, err := getDatabaseFromRef(defaultDBKey)
	if err != nil {
		log.Fatal("[FATAL] Could not find default database! Please make sure that databases are correctly configured and started.")
	}

	return drv, dsn
}
