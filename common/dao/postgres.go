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

package dao

import (
	"database/sql"
	"fmt"
	"sync"
)

var (
	pqLock                                = &sync.Mutex{}
	DB_PQ_CONNECTIONS_PERCENT_PER_REQUEST = 5
	DB_PQ_MAX_CONNECTIONS_PERCENT         = 90
	DB_PQ_IDLE_CONNECTIONS_PERCENT        = 25
)

type postgres struct {
	conn *sql.DB
}

func (p *postgres) Open(dsn string) (Conn, error) {

	// Not necessary?
	// TODO remove, will fail if values have a space
	// connStr, err := pq.ParseURL(dsn)
	// if err != nil {
	// 	fmt.Println("Cannot parse DSN", dsn)
	// 	return nil, err
	// }
	// params := make(map[string]string)
	// for _, part := range strings.Split(connStr, " ") {
	// 	param := strings.Split(part, "=")
	// 	params[param[0]] = param[1]
	// }

	conn, err := sql.Open("postgres", dsn)
	if err != nil {
		return nil, err
	}

	err = conn.Ping()
	if err != nil {
		fmt.Println("Ping failed with: ", err)
		return nil, err
	}

	// TODO for the time being, lazy creatíon of the DB at startup is not supported:
	// We have to connect to another existing DB (usually template1) to create the Cells DB
	// Yet this DB might not exist on some production servers.

	p.conn = conn
	return conn, nil
}

func (p *postgres) GetConn() Conn {
	return p.conn
}

func (p *postgres) getMaxTotalConnections() int {
	db := p.conn

	var num int
	if err := db.QueryRow(`SELECT setting FROM pg_settings where name = 'max_connections';`).Scan(&num); err != nil {
		return 0
	}

	return (num * DB_PQ_MAX_CONNECTIONS_PERCENT) / 100
}

func (p *postgres) SetMaxConnectionsForWeight(num int) {

	pqLock.Lock()
	defer pqLock.Unlock()

	maxConns := p.getMaxTotalConnections() * (num * DB_PQ_CONNECTIONS_PERCENT_PER_REQUEST) / 100
	maxIdleConns := maxConns * DB_PQ_IDLE_CONNECTIONS_PERCENT / 100

	p.conn.SetMaxOpenConns(maxConns)
	p.conn.SetMaxIdleConns(maxIdleConns)
}
