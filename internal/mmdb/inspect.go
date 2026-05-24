package mmdb

import (
	"net"

	"github.com/oschwald/maxminddb-golang"
)

func Inspect(path, ip string) (Record, bool, error) {
	db, err := maxminddb.Open(path)
	if err != nil {
		return Record{}, false, err
	}
	defer db.Close()
	var rec Record
	_, ok, err := db.LookupNetwork(net.ParseIP(ip), &rec)
	if err != nil {
		return Record{}, false, err
	}
	return rec, ok, nil
}
