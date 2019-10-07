// Scrape `info_schema.global_status`.

package collector

import (
	"context"
	"database/sql"

	"github.com/prometheus/client_golang/prometheus"
)

const globalStatusRequestRatesQuery = `
	SELECT
		SUM(variable_value) AS SUM_REQUEST_RATE
		FROM information_schema.global_status
		WHERE variable_name IN ('com_select', 'com_update', 'com_delete', 'com_insert');
	`

// Metric descriptors.
var (
	informationSchemaGlobalStatusRequestSumTotalDesc = prometheus.NewDesc(
		prometheus.BuildFQName(namespace, informationSchema, "global_status_request_rate"),
		"The total count of requests.",
		nil, nil,
	)
)

// ScrapeGlobalStatusRequestRatesSum collects from `information_schema.global_status`.
type ScrapeGlobalStatusRequestRatesSum struct{}

// Name of the Scraper. Should be unique.
func (ScrapeGlobalStatusRequestRatesSum) Name() string {
	return "globalStatusRequestRatesSum"
}

// Help describes the role of the Scraper.
func (ScrapeGlobalStatusRequestRatesSum) Help() string {
	return "Returns request rate from information_schema.global_status"
}

// Version of MySQL from which scraper is available.
func (ScrapeGlobalStatusRequestRatesSum) Version() float64 {
	return 5.7
}

// Scrape collects data from database connection and sends it over channel as prometheus metric.
func (ScrapeGlobalStatusRequestRatesSum) Scrape(ctx context.Context, db *sql.DB, ch chan<- prometheus.Metric) error {
	// Timers here are returned in picoseconds.
	globalStatusRequestRatesRows, err := db.QueryContext(ctx, globalStatusRequestRatesQuery)
	if err != nil {
		return err
	}
	defer globalStatusRequestRatesRows.Close()

	var (
		total, createdTmpDiskTables, createdTmpTables, errors uint64
		lockTime, noGoodIndexUsed, noIndexUsed, rowsAffected  uint64
		rowsExamined, rowsSent, selectFullJoin                uint64
		selectFullRangeJoin, selectRange, selectRangeCheck    uint64
		selectScan, sortMergePasses, sortRange, sortRows      uint64
		sortScan, timerWait, warnings                         uint64
	)

	for globalStatusRequestRatesRows.Next() {
		if err := globalStatusRequestRatesRows.Scan(
			&total, &createdTmpDiskTables, &createdTmpTables, &errors,
			&lockTime, &noGoodIndexUsed, &noIndexUsed, &rowsAffected,
			&rowsExamined, &rowsSent, &selectFullJoin,
			&selectFullRangeJoin, &selectRange, &selectRangeCheck,
			&selectScan, &sortMergePasses, &sortRange, &sortRows,
			&sortScan, &timerWait, &warnings,
		); err != nil {
			return err
		}
		ch <- prometheus.MustNewConstMetric(
			informationSchemaGlobalStatusRequestSumTotalDesc, prometheus.CounterValue, float64(total),
		)
	}
	return nil
}

// check interface
var _ Scraper = ScrapeGlobalStatusRequestRatesSum{}
