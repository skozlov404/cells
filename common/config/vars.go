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

// Package config provides tools for managing configurations
package config

import (
	"github.com/pydio/cells/x/filex"
)

var (
	// pydioConfigDir is the cached location for the configuration
	pydioConfigDir string

	// PydioConfigFile is the default file name for the configuration
	PydioConfigFile = "pydio.json"

	// VersionsStore is the default Version Store for the configuration
	VersionsStore filex.VersionsStore
)

// DefaultOAuthClientID set the default client id to use
const DefaultOAuthClientID = "cells-frontend"
